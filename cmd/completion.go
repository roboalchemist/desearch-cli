package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/roboalchemist/desearch-cli/pkg/api"
	"github.com/roboalchemist/desearch-cli/pkg/auth"
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
	Args: cobra.ExactArgs(1),
	RunE: runCompletion,
}

// Shell completion commands
var completionBashCmd = &cobra.Command{
	Use:   "bash",
	Short: "Generate Bash completion script",
	RunE: func(cmd *cobra.Command, args []string) error {
		return rootCmd.GenBashCompletionFile(os.Stdout)
	},
}

var completionZshCmd = &cobra.Command{
	Use:   "zsh",
	Short: "Generate Zsh completion script",
	RunE: func(cmd *cobra.Command, args []string) error {
		return rootCmd.GenZshCompletionFile(os.Stdout)
	},
}

var completionFishCmd = &cobra.Command{
	Use:   "fish",
	Short: "Generate Fish completion script",
	RunE: func(cmd *cobra.Command, args []string) error {
		return rootCmd.GenFishCompletionFile(os.Stdout)
	},
}

var completionPowerShellCmd = &cobra.Command{
	Use:   "powershell",
	Short: "Generate PowerShell completion script",
	RunE: func(cmd *cobra.Command, args []string) error {
		return rootCmd.GenPowerShellCompletionFile(os.Stdout)
	},
}

var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generate shell completion scripts",
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

	streaming := true
	resultType := "LINKS_WITH_FINAL_SUMMARY"

	req := &api.SearchRequest{
		Prompt:     query,
		Streaming:  &streaming,
		ResultType: &resultType,
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

	// Stream completion text chunks as they arrive
	for {
		select {
		case <-ctx.Done():
			// User cancelled (e.g., Ctrl+C)
			return ctx.Err()
		default:
		}

		chunk, err := reader.ReadBytes('\n')
		if len(chunk) > 0 {
			if completionJSON {
				// Accumulate completion text from JSON chunks
				var partial map[string]interface{}
				if err := json.Unmarshal(chunk, &partial); err == nil {
					if completion, ok := partial["completion"].(string); ok && completion != "" {
						completionBuilder.WriteString(completion)
					}
					if text, ok := partial["text"].(string); ok && text != "" {
						if _, hasCompletion := partial["completion"]; !hasCompletion {
							completionBuilder.WriteString(text)
						}
					}
				}
			} else {
				// Try to parse as a partial response to extract completion chunks
				// The stream may send JSON objects with completion text
				var partial map[string]interface{}
				if err := json.Unmarshal(chunk, &partial); err == nil {
					// Look for completion field
					if completion, ok := partial["completion"].(string); ok && completion != "" {
						// Print without extra newline, flush immediately
						fmt.Print(completion)
						os.Stdout.Sync()
					}
					// Also check for text field which may contain completion chunks
					if text, ok := partial["text"].(string); ok && text != "" {
						// Only print if completion is not set or empty
						if _, hasCompletion := partial["completion"]; !hasCompletion {
							fmt.Print(text)
							os.Stdout.Sync()
						}
					}
				} else {
					// Not JSON - try to print raw chunk if it looks like text
					trimmed := strings.TrimSpace(string(chunk))
					if trimmed != "" {
						// Check if it might be plain text completion
						if !strings.HasPrefix(trimmed, "{") && !strings.HasPrefix(trimmed, "[") {
							fmt.Print(trimmed)
							os.Stdout.Sync()
						}
					}
				}
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

	// Print final newline when done
	if completionJSON {
		// Output structured JSON at the end
		output := map[string]string{
			"query":      query,
			"completion": completionBuilder.String(),
		}
		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
	} else {
		fmt.Println()
	}

	return nil
}
