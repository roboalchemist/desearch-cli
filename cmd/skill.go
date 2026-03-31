package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/roboalchemist/desearch-cli/skill"
	"github.com/spf13/cobra"
)

var skillPrintCmd = &cobra.Command{
	Use:     "print",
	Short:   "Print SKILL.md to stdout",
	Example: `  desearch skill print`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(skill.SKILLMD)
		return nil
	},
}

var skillAddCmd = &cobra.Command{
	Use:     "add",
	Short:   "Install skill to ~/.claude/skills/",
	Example: `  desearch skill add`,
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("could not determine home directory: %w", err)
		}
		destDir := filepath.Join(home, ".claude", "skills", "desearch-cli")
		destPath := filepath.Join(destDir, "SKILL.md")

		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("creating skill directory: %w", err)
		}

		if err := os.WriteFile(destPath, []byte(skill.SKILLMD), 0644); err != nil {
			return fmt.Errorf("writing SKILL.md: %w", err)
		}

		fmt.Printf("Skill installed to %s\n", destPath)
		return nil
	},
}

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Manage Claude Code skill",
}

func init() {
	rootCmd.AddCommand(skillCmd)
	skillCmd.AddCommand(skillPrintCmd, skillAddCmd)
}
