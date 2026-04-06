package cmd

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/roboalchemist/desearch-cli/pkg/api"
	"github.com/roboalchemist/desearch-cli/pkg/auth"
	"github.com/roboalchemist/desearch-cli/pkg/output"
	"github.com/spf13/cobra"
)

var (
	flagTool             []string
	flagDateFilter       string
	flagStartDate        string
	flagEndDate          string
	flagStreaming        bool
	flagResultType       string
	flagCount            int
	flagSystemMsg        string
	flagScoringSystemMsg string
	flagNoAI             bool
	flagPlaintext        bool
	flagDryRun           bool
	flagJQ               string
	flagFields           string
	flagStdin            bool
	flagNoHistory        bool
)

func getAPIKey() string {
	key := apiKey
	if key == "" {
		key = auth.GetAPIKey()
	}
	return key
}

func buildSearchRequest(query string, cfg *auth.Config) *api.SearchRequest {
	req := &api.SearchRequest{
		Prompt: query,
	}

	req.Tools = resolveTools(flagTool, cfg)

	if flagDateFilter != "" {
		req.DateFilter = &flagDateFilter
	}
	if flagStartDate != "" {
		req.StartDate = &flagStartDate
	}
	if flagEndDate != "" {
		req.EndDate = &flagEndDate
	}
	if flagSystemMsg != "" {
		req.SystemMessage = &flagSystemMsg
	}
	if flagScoringSystemMsg != "" {
		req.ScoringSystemMessage = &flagScoringSystemMsg
	}
	if flagResultType != "" {
		req.ResultType = &flagResultType
	} else {
		defaultRT := "LINKS_WITH_FINAL_SUMMARY"
		req.ResultType = &defaultRT
	}
	if flagCount != 0 {
		req.Count = &flagCount
	}

	return req
}

func runSearch(cmd *cobra.Command, args []string) error {
	// Validate --count range (10-200)
	if flagCount != 0 && (flagCount < 10 || flagCount > 200) {
		return fmt.Errorf("--count must be between 10 and 200, got %d", flagCount)
	}

	// Validate --date-filter enum
	validDateFilters := map[string]bool{
		"PAST_24_HOURS": true,
		"PAST_2_DAYS":   true,
		"PAST_WEEK":     true,
		"PAST_2_WEEKS":  true,
		"PAST_MONTH":    true,
		"PAST_2_MONTHS": true,
		"PAST_YEAR":     true,
		"PAST_2_YEARS":  true,
	}
	if flagDateFilter != "" && !validDateFilters[flagDateFilter] {
		return fmt.Errorf("--date-filter %q is not valid; valid values: PAST_24_HOURS, PAST_2_DAYS, PAST_WEEK, PAST_2_WEEKS, PAST_MONTH, PAST_2_MONTHS, PAST_YEAR, PAST_2_YEARS", flagDateFilter)
	}

	// Validate --start-date format
	if flagStartDate != "" {
		if _, err := time.Parse("2006-01-02", flagStartDate); err != nil {
			return fmt.Errorf("--start-date must be YYYY-MM-DD, got %q", flagStartDate)
		}
	}

	// Validate --end-date format
	if flagEndDate != "" {
		if _, err := time.Parse("2006-01-02", flagEndDate); err != nil {
			return fmt.Errorf("--end-date must be YYYY-MM-DD, got %q", flagEndDate)
		}
	}

	// Validate --jq requires --json, --no-ai, or --dry-run (all produce JSON output)
	if flagJQ != "" && !jsonOut && !flagNoAI && !flagDryRun {
		return fmt.Errorf("--jq requires --json, --no-ai, or --dry-run to be set")
	}

	// Validate --fields requires --json (--fields filters API response).
	// --dry-run produces JSON output that can also be filtered, so it also qualifies.
	if flagFields != "" && !jsonOut && !flagDryRun {
		return fmt.Errorf("--fields requires --json to be set")
	}

	// If --stdin is set, read queries from stdin and run each one
	if flagStdin {
		scanner := bufio.NewScanner(os.Stdin)
		var queries []string
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				queries = append(queries, line)
			}
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}

		var lastErr error
		for i, q := range queries {
			if i > 0 {
				fmt.Fprintln(os.Stdout, "---")
			}
			fmt.Fprintf(os.Stdout, "Query: %s\n", q)
			if err := runSearchOne(q); err != nil {
				fmt.Fprintf(os.Stderr, "Error for query %q: %v\n", q, err)
				lastErr = err
			}
		}
		return lastErr
	}

	return runSearchOne(args[0])
}

func runSearchOne(query string) error {
	if flagVerbose && !flagQuiet {
		fmt.Fprintf(os.Stderr, "Searching %d source(s)...\n", len(flagTool))
	}

	cfg, _ := auth.LoadConfig()
	req := buildSearchRequest(query, cfg)

	// Dry-run: print the request as JSON and return without calling the API
	if flagDryRun {
		data, err := json.MarshalIndent(req, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		if flagJQ != "" {
			filtered, err := output.EvaluateJQ(data, flagJQ)
			if err != nil {
				return fmt.Errorf("jq filter failed: %w", err)
			}
			fmt.Fprint(os.Stdout, string(filtered))
		} else if flagFields != "" {
			filtered, err := output.FilterJSONFields(data, flagFields)
			if err != nil {
				return fmt.Errorf("filtering fields: %w", err)
			}
			fmt.Fprint(os.Stdout, string(filtered))
		} else {
			fmt.Fprint(os.Stdout, string(data))
		}
		return nil
	}

	apiKeyVal := getAPIKey()
	if apiKeyVal == "" {
		return fmt.Errorf("no API key found")
	}

	client := api.NewClient(apiKeyVal)

	if flagStreaming {
		return runSearchStream(nil, client, req)
	}
	return runSearchNormal(nil, client, req)
}

func runSearchNormal(cmd *cobra.Command, client *api.Client, req *api.SearchRequest) error {
	ctx := context.Background()
	t0 := time.Now()
	resp, err := client.Search(ctx, req)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}
	latencyMs := int(time.Since(t0).Milliseconds())

	formatter := output.NewFormatter(output.OutputFlags{
		JSON:         jsonOut || flagNoAI, // jsonOut from root.go, or --no-ai implies raw
		NoAI:         flagNoAI,
		Plaintext:    flagPlaintext,
		FilterFields: flagFields,
	})
	formatted := formatter.Format(resp)
	if flagJQ != "" {
		filtered, err := output.EvaluateJQ([]byte(formatted), flagJQ)
		if err != nil {
			return fmt.Errorf("jq filter failed: %w", err)
		}
		fmt.Fprint(os.Stdout, string(filtered))
	} else {
		fmt.Fprint(os.Stdout, formatted)
	}

	// Write history after successful response (non-fatal on error)
	cfg, _ := auth.LoadConfig()
	configDir, configDirErr := auth.ConfigDir()
	if configDirErr == nil {
		params := map[string]interface{}{
			"prompt": req.Prompt,
			"tools":  req.Tools,
		}
		if req.DateFilter != nil {
			params["date_filter"] = *req.DateFilter
		}
		if req.StartDate != nil {
			params["start_date"] = *req.StartDate
		}
		if req.EndDate != nil {
			params["end_date"] = *req.EndDate
		}
		if req.ResultType != nil {
			params["result_type"] = *req.ResultType
		}
		if req.Count != nil {
			params["count"] = *req.Count
		}
		historyEnabled := cfg != nil && cfg.HistoryEnabled && !flagNoHistory
		if histErr := output.WriteHistory(configDir, "search", params, resp, latencyMs, historyEnabled); histErr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to write history: %v\n", histErr)
		}
	}

	return nil
}

func runSearchStream(cmd *cobra.Command, client *api.Client, req *api.SearchRequest) error {
	if flagVerbose && !flagQuiet {
		fmt.Fprintf(os.Stderr, "Streaming results...\n")
	}
	ctx := context.Background()
	t0 := time.Now()
	reader, err := client.SearchStream(ctx, req)
	if err != nil {
		return fmt.Errorf("stream search failed: %w", err)
	}
	defer reader.Close()

	streamer := &output.StreamingFormatter{JSON: jsonOut}

	// Accumulate SSE text chunks for history (mirrors how runCompletion does it
	// for the ai command — DC1-118 pattern).
	var buf strings.Builder

	// Read line-by-line using ReadBytes; split each line on "data: " boundaries
	// to handle multiple SSE events packed on a single line.
	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			segments := bytes.Split(line, []byte("data: "))
			for _, seg := range segments {
				content := output.ParseSSEEvent(seg)
				if content != "" {
					buf.WriteString(content)
					streamer.WriteChunk(content)
				}
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("error reading stream: %w", err)
		}
	}

	latencyMs := int(time.Since(t0).Milliseconds())
	streamer.Finalize(req.Prompt, "")

	// Write history after stream is fully consumed (non-fatal on error)
	cfg, _ := auth.LoadConfig()
	configDir, configDirErr := auth.ConfigDir()
	if configDirErr == nil {
		params := map[string]interface{}{
			"prompt":    req.Prompt,
			"tools":     req.Tools,
			"streaming": true,
		}
		if req.DateFilter != nil {
			params["date_filter"] = *req.DateFilter
		}
		if req.StartDate != nil {
			params["start_date"] = *req.StartDate
		}
		if req.EndDate != nil {
			params["end_date"] = *req.EndDate
		}
		if req.ResultType != nil {
			params["result_type"] = *req.ResultType
		}
		if req.Count != nil {
			params["count"] = *req.Count
		}
		historyEnabled := cfg != nil && cfg.HistoryEnabled && !flagNoHistory
		// Accumulate the streamed text chunks into the response so history
		// files contain the actual completion instead of null (DC1-119).
		streamResponse := map[string]interface{}{"completion": buf.String()}
		if histErr := output.WriteHistory(configDir, "search", params, streamResponse, latencyMs, historyEnabled); histErr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to write history: %v\n", histErr)
		}
	}

	return nil
}

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search using Desearch AI",
	Long: `Search the web using Desearch AI's contextual search engine.

Supports multiple sources including web, hackernews, reddit, wikipedia,
youtube, twitter, and arxiv. Results can be streamed in real-time or
returned as a complete response with AI summarization.`,
	Example: `  desearch "golang best practices"
  desearch "rust vs go" --tool web --count 20
  desearch "AI news" --date-filter PAST_2_DAYS --streaming`,
	Args: func(cmd *cobra.Command, args []string) error {
		if flagStdin {
			return nil // stdin mode: no positional args required
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE:       runSearch,
	SuggestFor: []string{"serch", "srch", "seach", "searc"},
}

func init() {
	searchCmd.Flags().StringSliceVar(&flagTool, "tool", nil, "Sources to query (web, hackernews, reddit, wikipedia, youtube, twitter, arxiv). Empty queries all.")
	searchCmd.Flags().StringVar(&flagDateFilter, "date-filter", "", "Predefined date range (PAST_24_HOURS, PAST_2_DAYS, PAST_WEEK, PAST_2_WEEKS, PAST_MONTH, PAST_2_MONTHS, PAST_YEAR, PAST_2_YEARS)")
	searchCmd.Flags().StringVar(&flagStartDate, "start-date", "", "Start date in YYYY-MM-DD format")
	searchCmd.Flags().StringVar(&flagEndDate, "end-date", "", "End date in YYYY-MM-DD format")
	searchCmd.Flags().BoolVar(&flagStreaming, "streaming", false, "Stream results as they arrive")
	searchCmd.Flags().StringVar(&flagResultType, "result-type", "", "Result type: ONLY_LINKS or LINKS_WITH_FINAL_SUMMARY (default LINKS_WITH_FINAL_SUMMARY)")
	searchCmd.Flags().IntVar(&flagCount, "count", 0, "Number of results per source (10-200)")
	searchCmd.Flags().StringVar(&flagSystemMsg, "system-message", "", "System message to influence AI behavior")
	searchCmd.Flags().StringVar(&flagScoringSystemMsg, "scoring-system-message", "", "System message to influence result scoring/ranking")
	searchCmd.Flags().BoolVar(&flagNoAI, "no-ai", false, "Skip AI completion/summary")
	searchCmd.Flags().BoolVarP(&flagPlaintext, "plaintext", "p", false, "Output as tab-separated values (title\\turl\\tsnippet)")
	searchCmd.Flags().BoolVarP(&flagDryRun, "dry-run", "D", false, "Build request and print as JSON without calling the API")
	searchCmd.Flags().StringVar(&flagJQ, "jq", "", "jq expression to filter JSON output (requires --json or --no-ai)")
	searchCmd.Flags().StringVar(&flagFields, "fields", "", "Comma-separated top-level JSON fields to include in output (requires --json)")
	searchCmd.Flags().BoolVar(&flagStdin, "stdin", false, "Read queries from stdin (one per line)")
	searchCmd.Flags().BoolVar(&flagNoHistory, "no-history", false, "Skip writing to history even when history_enabled is set in config")

	searchCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		cmd.Parent().HelpFunc()(cmd, args)
	})

	rootCmd.AddCommand(searchCmd)
}
