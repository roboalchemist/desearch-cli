package cmd

import (
	"testing"

	"github.com/roboalchemist/desearch-cli/pkg/auth"
)

func TestResolveTools(t *testing.T) {
	tests := []struct {
		name      string
		flagTools []string
		cfg       *auth.Config
		want      []string
	}{
		{
			name:      "empty flags and nil config returns default",
			flagTools: nil,
			cfg:       nil,
			want:      []string{"web"},
		},
		{
			name:      "empty flags and empty config returns default",
			flagTools: nil,
			cfg:       &auth.Config{DefaultTools: nil},
			want:      []string{"web"},
		},
		{
			name:      "empty flags with config tools uses config",
			flagTools: nil,
			cfg:       &auth.Config{DefaultTools: []string{"hackernews", "reddit"}},
			want:      []string{"hackernews", "reddit"},
		},
		{
			name:      "flags provided override config",
			flagTools: []string{"arxiv"},
			cfg:       &auth.Config{DefaultTools: []string{"hackernews"}},
			want:      []string{"arxiv"},
		},
		{
			name:      "flags provided override nil config",
			flagTools: []string{"web", "twitter"},
			cfg:       nil,
			want:      []string{"web", "twitter"},
		},
		{
			name:      "empty config DefaultTools slice returns default",
			flagTools: nil,
			cfg:       &auth.Config{DefaultTools: []string{}},
			want:      []string{"web"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveTools(tt.flagTools, tt.cfg)
			if len(got) != len(tt.want) {
				t.Errorf("resolveTools() = %v (len %d), want %v (len %d)", got, len(got), tt.want, len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("resolveTools()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
