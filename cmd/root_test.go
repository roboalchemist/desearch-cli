package cmd

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func TestHasChangedFlag(t *testing.T) {
	t.Run("returns false when flag not changed", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().Bool("test-flag", false, "test")
		if hasChangedFlag(cmd, "test-flag") {
			t.Error("expected false for unchanged flag")
		}
	})

	t.Run("returns true when flag changed on cmd", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().Bool("dry-run", false, "dry-run")
		if err := cmd.Flags().Set("dry-run", "true"); err != nil {
			t.Fatal(err)
		}
		if !hasChangedFlag(cmd, "dry-run") {
			t.Error("expected true for changed flag")
		}
	})

	t.Run("returns true when flag changed on parent", func(t *testing.T) {
		parent := &cobra.Command{Use: "parent"}
		child := &cobra.Command{Use: "child"}
		parent.AddCommand(child)
		parent.Flags().Bool("dry-run", false, "dry-run")
		if err := parent.Flags().Set("dry-run", "true"); err != nil {
			t.Fatal(err)
		}
		if !hasChangedFlag(child, "dry-run") {
			t.Error("expected true for flag changed on parent")
		}
	})

	t.Run("returns false when neither cmd nor parent has flag", func(t *testing.T) {
		parent := &cobra.Command{Use: "parent"}
		child := &cobra.Command{Use: "child"}
		parent.AddCommand(child)
		if hasChangedFlag(child, "nonexistent") {
			t.Error("expected false for nonexistent flag")
		}
	})
}

func TestHasDryRunInArgs(t *testing.T) {
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })

	t.Run("returns false when no -- separator", func(t *testing.T) {
		os.Args = []string{"desearch", "search", "--dry-run"}
		if hasDryRunInArgs() {
			t.Error("expected false when -- not present")
		}
	})

	t.Run("returns true when --dry-run after --", func(t *testing.T) {
		os.Args = []string{"desearch", "--", "search", "--dry-run"}
		if !hasDryRunInArgs() {
			t.Error("expected true when --dry-run follows --")
		}
	})

	t.Run("returns true when --fields after --", func(t *testing.T) {
		os.Args = []string{"desearch", "--", "search", "--fields"}
		if !hasDryRunInArgs() {
			t.Error("expected true when --fields follows --")
		}
	})

	t.Run("returns true when -D after --", func(t *testing.T) {
		os.Args = []string{"desearch", "--", "search", "-D"}
		if !hasDryRunInArgs() {
			t.Error("expected true when -D follows --")
		}
	})

	t.Run("returns false when -- is the last arg", func(t *testing.T) {
		os.Args = []string{"desearch", "--"}
		if hasDryRunInArgs() {
			t.Error("expected false when -- has no following args")
		}
	})

	t.Run("returns false when after -- no dry-run flags", func(t *testing.T) {
		os.Args = []string{"desearch", "--", "search", "query"}
		if hasDryRunInArgs() {
			t.Error("expected false when no dry-run flags after --")
		}
	})
}

func TestIsNoAuthCommand(t *testing.T) {
	tests := []struct {
		name    string
		cmdName string
		want    bool
	}{
		{"version command", "version", true},
		{"help command", "help", true},
		{"docs command", "docs", true},
		{"skill command", "skill", true},
		{"print command", "print", true},
		{"add command", "add", true},
		{"completion command", "completion", true},
		{"ai command", "ai", false},
		{"bash command", "bash", true},
		{"zsh command", "zsh", true},
		{"fish command", "fish", true},
		{"powershell command", "powershell", true},
		{"clear command", "clear", true},
		{"desearch root", "desearch", true},
		{"search command", "search", false},
		{"config command", "config", false},
		{"show command", "show", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: tt.cmdName}
			got := isNoAuthCommand(cmd)
			if got != tt.want {
				t.Errorf("isNoAuthCommand(%q) = %v, want %v", tt.cmdName, got, tt.want)
			}
		})
	}

	t.Run("child of no-auth parent is also no-auth", func(t *testing.T) {
		parent := &cobra.Command{Use: "skill"}
		child := &cobra.Command{Use: "somechild"}
		parent.AddCommand(child)
		if !isNoAuthCommand(child) {
			t.Error("expected child of skill to be no-auth")
		}
	})
}

func TestGetJSONOut(t *testing.T) {
	orig := jsonOut
	t.Cleanup(func() { jsonOut = orig })

	jsonOut = false
	if GetJSONOut() {
		t.Error("expected false when jsonOut is false")
	}

	jsonOut = true
	if !GetJSONOut() {
		t.Error("expected true when jsonOut is true")
	}
}

func TestRootCmd(t *testing.T) {
	cmd := RootCmd()
	if cmd == nil {
		t.Error("RootCmd() should not return nil")
		return
	}
	if cmd.Use != "desearch" {
		t.Errorf("RootCmd().Use = %q, want %q", cmd.Use, "desearch")
	}
}
