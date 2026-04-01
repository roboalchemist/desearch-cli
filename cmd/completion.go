package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/roboalchemist/desearch-cli/pkg/api"
	"github.com/roboalchemist/desearch-cli/pkg/auth"
	"github.com/roboalchemist/desearch-cli/pkg/output"
	"github.com/spf13/cobra"
)

var (
	completionSystemMessage string
	completionJSON          bool
)

var aiCmd = &cobra.Command{
	Use:   "ai <query>",
	Short: "Get an AI-generated summary without per-source results",
	Long: `Streams an AI-generated summary for the given query.

This command always streams results. It does not return per-source search results,
only the final AI summary.

Example:
  desearch ai "what is bittensor"
  desearch ai "explain transformers" --system-message "Summarize in simple terms"`,
	Args:    cobra.ExactArgs(1),
	RunE:    runCompletion,
	Example: `desearch ai "what is bittensor"`,
}

// Shell completion commands
var completionBashCmd = &cobra.Command{
	Use:     "bash",
	Short:   "Generate Bash completion script",
	Example: `desearch completion bash`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return rootCmd.GenBashCompletion(os.Stdout)
	},
}

var completionZshCmd = &cobra.Command{
	Use:     "zsh",
	Short:   "Generate Zsh completion script",
	Example: `desearch completion zsh`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return rootCmd.GenZshCompletion(os.Stdout)
	},
}

var completionFishCmd = &cobra.Command{
	Use:     "fish",
	Short:   "Generate Fish completion script",
	Example: `desearch completion fish`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return rootCmd.GenFishCompletion(os.Stdout, true)
	},
}

var completionPowerShellCmd = &cobra.Command{
	Use:     "powershell",
	Short:   "Generate PowerShell completion script",
	Example: `desearch completion powershell`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return rootCmd.GenPowerShellCompletion(os.Stdout)
	},
}

var completionCmd = &cobra.Command{
	Use:     "completion",
	Short:   "Generate shell completion scripts",
	Example: `desearch completion bash`,
}

func init() {
	rootCmd.AddCommand(aiCmd)
	aiCmd.Flags().StringVar(&completionSystemMessage, "system-message", "", "Optional system message to override the default")
	aiCmd.Flags().BoolVar(&completionJSON, "json", false, "Output raw JSON response")

	rootCmd.AddCommand(completionCmd)
	completionCmd.AddCommand(completionBashCmd, completionZshCmd, completionFishCmd, completionPowerShellCmd)
}

func runCompletion(cmd *cobra.Command, args []string) error {
	query := args[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Get API key
	apiKey := apiKey
	if apiKey == "" {
		apiKey = auth.GetAPIKey()
	}
	if apiKey == "" {
		return fmt.Errorf("no API key found")
	}

	client := api.NewClient(apiKey)

	cfg, _ := auth.LoadConfig()

	streaming := true
	resultType := "LINKS_WITH_FINAL_SUMMARY"

	req := &api.SearchRequest{
		Prompt:     query,
		Streaming:  &streaming,
		ResultType: &resultType,
		Tools:      resolveTools(nil, cfg),
	}

	if completionSystemMessage != "" {
		req.SystemMessage = &completionSystemMessage
	}

	reader, err := client.SearchStream(ctx, req)
	if err != nil {
		return fmt.Errorf("search stream failed: %w", err)
	}

	// When --json flag is set, accumulate completion and output at end
	var completionBuilder strings.Builder
	streamer := &output.StreamingFormatter{JSON: completionJSON}

	// processSSESegment parses a single SSE data segment (after stripping "data: " prefix)
	// and extracts the completion text, writing it to the streamer or accumulating it.
	//
	// The Desearch API sends streaming events in the format:
	//   {"type": "text", "role": "summary", "content": "..."}
	// Only events with "type" == "text" carry output; others (metadata, done signals) are skipped silently.
	processSSESegment := func(seg []byte) {
		seg = bytes.TrimSpace(seg)
		if len(seg) == 0 {
			return
		}
		// Skip the SSE stream-end sentinel
		if string(seg) == "[DONE]" {
			return
		}
		var partial map[string]interface{}
		if err := json.Unmarshal(seg, &partial); err != nil {
			// Not valid JSON — skip silently (could be partial/garbled data)
			return
		}
		// Only emit content for "type": "text" events; skip metadata and done signals silently.
		eventType, _ := partial["type"].(string)
		if eventType != "text" {
			return
		}
		content, _ := partial["content"].(string)
		if content == "" {
			return
		}
		if completionJSON {
			completionBuilder.WriteString(content)
		} else {
			streamer.WriteChunk(content)
		}
	}

	// Stream completion text chunks as they arrive.
	// The Desearch API may send multiple SSE events on a single line without
	// newline separators (e.g. `data: {...}data: {...}`). We split each read
	// on "data: " boundaries so every event is parsed independently.
	for {
		select {
		case <-ctx.Done():
			// User cancelled (e.g., Ctrl+C)
			return ctx.Err()
		default:
		}

		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			// Split on "data: " to handle multiple events packed on one line.
			segments := bytes.Split(line, []byte("data: "))
			for _, seg := range segments {
				processSSESegment(seg)
			}
		}

		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			// Check if context was cancelled
			if ctx.Err() != nil {
				return ctx.Err()
			}
			// Some other error - could be done streaming
			break
		}
	}

	// Print final newline when done (or JSON object in --json mode)
	streamer.Finalize(query, completionBuilder.String())

	return nil
}
