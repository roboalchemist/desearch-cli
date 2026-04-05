package output

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/roboalchemist/desearch-cli/pkg/api"
)

// historyMeta holds metadata written to the top-level "meta" envelope in each
// history file.
type historyMeta struct {
	Timestamp string                 `json:"timestamp"`
	Hostname  string                 `json:"hostname"`
	Command   string                 `json:"command"`
	Params    map[string]interface{} `json:"params"`
	LatencyMs int                    `json:"latency_ms"`
}

// historyEnvelope is the top-level JSON structure written to each history file.
type historyEnvelope struct {
	Meta     historyMeta         `json:"meta"`
	Response *api.SearchResponse `json:"response"`
}

// slugRe matches any character that is NOT alphanumeric. Replaced with "-".
var slugRe = regexp.MustCompile(`[^a-zA-Z0-9]+`)

// makeSlug returns a filesystem-safe slug derived from the prompt field of
// params. It replaces non-alphanumeric characters with "-", trims leading/
// trailing dashes, and caps the result at 40 characters.
func makeSlug(params map[string]interface{}) string {
	raw := ""
	if p, ok := params["prompt"]; ok {
		raw = fmt.Sprintf("%v", p)
	}
	slug := slugRe.ReplaceAllString(raw, "-")
	slug = strings.Trim(slug, "-")
	if len(slug) > 40 {
		slug = slug[:40]
	}
	return slug
}

// WriteHistory writes a JSON history file for an API call. It is a no-op when
// historyEnabled is false.
//
// The file is created at:
//
//	<configDir>/history/<cmd>/<YYYY>/<MM>/<DD>/<HH-MM-SS>_<slug>_<hostname>.json
//
// The JSON envelope contains a "meta" block with request metadata and a
// "response" block with the full API response. The file is written with mode
// 0600. On error, a warning is printed to stderr (the caller's command is NOT
// failed).
func WriteHistory(configDir string, cmd string, params map[string]interface{}, response *api.SearchResponse, latencyMs int, historyEnabled bool) error {
	if !historyEnabled {
		return nil
	}

	now := time.Now().UTC()

	// Sanitize params — strip api_key if somehow present.
	cleanParams := make(map[string]interface{}, len(params))
	for k, v := range params {
		if k == "api_key" {
			continue
		}
		cleanParams[k] = v
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	// Sanitize hostname for use in filenames.
	safeHost := regexp.MustCompile(`[^a-zA-Z0-9._-]`).ReplaceAllString(hostname, "-")

	slug := makeSlug(params)

	dirPath := fmt.Sprintf("%s/history/%s/%s/%s/%s",
		configDir,
		cmd,
		now.Format("2006"),
		now.Format("01"),
		now.Format("02"),
	)
	if err := os.MkdirAll(dirPath, 0700); err != nil {
		fmt.Fprintf(os.Stderr, "[history] error creating directory %s: %v\n", dirPath, err)
		return err
	}

	filename := fmt.Sprintf("%s_%s_%s.json",
		now.Format("15-04-05"),
		slug,
		safeHost,
	)
	filePath := dirPath + "/" + filename

	// Build timestamp in the same format as the tavily-cli reference:
	// "2006-01-02T15:04:05.000Z" (millisecond precision, UTC, Z suffix)
	timestamp := now.Format("2006-01-02T15:04:05.") + fmt.Sprintf("%03dZ", now.Nanosecond()/1e6)

	envelope := historyEnvelope{
		Meta: historyMeta{
			Timestamp: timestamp,
			Hostname:  hostname,
			Command:   cmd,
			Params:    cleanParams,
			LatencyMs: latencyMs,
		},
		Response: response,
	}

	data, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "[history] error marshaling JSON for %s: %v\n", filePath, err)
		return err
	}

	if err := os.WriteFile(filePath, data, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "[history] error writing %s: %v\n", filePath, err)
		return err
	}

	fmt.Fprintf(os.Stderr, "[history] %s\n", filePath)
	return nil
}
