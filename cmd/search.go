package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"

	"github.com/roboalchemist/desearch-cli/pkg/api"
	"github.com/roboalchemist/desearch-cli/pkg/auth"
	"github.com/roboalchemist/desearch-cli/pkg/output"
	"github.com/spf13/cobra"
)

var (
	flagTool        []string
	flagDateFilter  string
	flagStartDate   string
	flagEndDate     string
	flagStreaming   bool
	flagResultType  string
	flagCount       int
	flagSystemMsg   string
	flagNoAI        bool
	flagPlaintext   bool
	flagDryRun      bool
)

func getAPIKey() string {
	key := apiKey
	if key == "" {
		key = auth.GetAPIKey()
	}
	return key
}

func buildSearchRequest(query string) *api.SearchRequest {
	req := &api.SearchRequest{
		Prompt: query,
	}

	if len(flagTool) > 0 {
		req.Tools = flagTool
	}

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
	query := args[0]

	req := buildSearchRequest(query)

	// Dry-run: print the request as JSON and return without calling the API
	if flagDryRun {
		data, err := json.MarshalIndent(req, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		fmt.Print(string(data))
		return nil
	}

	apiKeyVal := getAPIKey()
	if apiKeyVal == "" {
		return fmt.Errorf("no API key found")
	}

	client := api.NewClient(apiKeyVal)

	if flagStreaming {
		return runSearchStream(cmd, client, req)
	}
	return runSearchNormal(cmd, client, req)
}

func runSearchNormal(cmd *cobra.Command, client *api.Client, req *api.SearchRequest) error {
	ctx := context.Background()
	resp, err := client.Search(ctx, req)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	formatter := output.NewFormatter(output.OutputFlags{
		JSON:      jsonOut || flagNoAI, // jsonOut from root.go, or --no-ai implies raw
		NoAI:      flagNoAI,
		Plaintext: flagPlaintext,
	})
	fmt.Print(formatter.Format(resp))
	return nil
}

func runSearchStream(cmd *cobra.Command, client *api.Client, req *api.SearchRequest) error {
	ctx := context.Background()
	reader, err := client.SearchStream(ctx, req)
	if err != nil {
		return fmt.Errorf("stream search failed: %w", err)
	}

	// Stream output directly to stdout using scanner for line-by-line output
	scanner := bufio.NewScanner(reader)
	// Increase scanner buffer for potentially long lines
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			fmt.Println(line)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stream: %w", err)
	}
	return nil
}

var searchCmd = &cobra.Command{
	Use:     "search [query]",
	Short:   "Search using Desearch AI",
	Long:    `Search the web using Desearch AI's contextual search engine.

Supports multiple sources including web, hackernews, reddit, wikipedia,
youtube, twitter, and arxiv. Results can be streamed in real-time or
returned as a complete response with AI summarization.`,
	Example: `  desearch "golang best practices"
  desearch "rust vs go" --tool web --count 20
  desearch "AI news" --date-filter PAST_2_DAYS --streaming`,
	Args:       cobra.ExactArgs(1),
	RunE:       runSearch,
	SuggestFor: []string{"serch", "srch", "seach", "searc"},
}

func init() {
	searchCmd.Flags().StringSliceVar(&flagTool, "tool", nil, "Sources to query (web, hackernews, reddit, wikipedia, youtube, twitter, arxiv). Empty queries all.")
	searchCmd.Flags().StringVar(&flagDateFilter, "date-filter", "", "Predefined date range (PAST_24_HOURS, PAST_2_DAYS, PAST_WEEK, PAST_2_WEEKS, PAST_MONTH, PAST_2_MONTHS, PAST_YEAR, PAST_2_YEARS)")
	searchCmd.Flags().StringVar(&flagStartDate, "start-date", "", "Start date in ISO8601 UTC format")
	searchCmd.Flags().StringVar(&flagEndDate, "end-date", "", "End date in ISO8601 UTC format")
	searchCmd.Flags().BoolVar(&flagStreaming, "streaming", false, "Stream results as they arrive")
	searchCmd.Flags().StringVar(&flagResultType, "result-type", "", "Result type: ONLY_LINKS or LINKS_WITH_FINAL_SUMMARY (default LINKS_WITH_FINAL_SUMMARY)")
	searchCmd.Flags().IntVar(&flagCount, "count", 0, "Number of results per source (10-200)")
	searchCmd.Flags().StringVar(&flagSystemMsg, "system-message", "", "System message to influence AI behavior")
	searchCmd.Flags().BoolVar(&flagNoAI, "no-ai", false, "Skip AI completion/summary")
	searchCmd.Flags().BoolVarP(&flagPlaintext, "plaintext", "p", false, "Output as tab-separated values (title\\turl\\tsnippet)")
	searchCmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Build request and print as JSON without calling the API")

	searchCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		cmd.Parent().HelpFunc()(cmd, args)
	})

	rootCmd.AddCommand(searchCmd)
}
