package cmd

import "github.com/roboalchemist/desearch-cli/pkg/auth"

// defaultTools is the hard-coded fallback when no tools are specified via flags
// or config. The Desearch API requires at least one tool in every request.
var defaultTools = []string{"web"}

// resolveTools returns the effective tool list for a search request.
// Priority order:
//  1. flagTools — tools provided via --tool flag(s)
//  2. cfg.DefaultTools — tools from ~/.config/desearch-cli/config.toml
//  3. defaultTools — hard-coded fallback (["web"])
func resolveTools(flagTools []string, cfg *auth.Config) []string {
	if len(flagTools) > 0 {
		return flagTools
	}
	if cfg != nil && len(cfg.DefaultTools) > 0 {
		return cfg.DefaultTools
	}
	return defaultTools
}
