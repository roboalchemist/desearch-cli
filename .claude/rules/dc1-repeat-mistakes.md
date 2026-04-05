---
scope: both
---

# desearch-cli Repeat Mistakes

## Auth Header — RAW KEY, NO "Bearer"

The Desearch API expects `Authorization: <API_KEY>` — raw key, **no "Bearer" prefix**.

The initial implementation had `Authorization: Bearer <API_KEY>` which caused 401 errors. The real API rejects "Bearer". Always verify: `Authorization: c.APIKey` (no prefix).

**Where this went wrong**: `pkg/api/client.go` lines 308 and 346 originally had `Bearer ` prefix.

## Integration Mock Server ≠ Real Client

The integration mock server in `integration_test.go` checks for `Authorization: Bearer <key>`. The real client sends the raw key. This means tests can pass against the mock but fail against the real API (or vice versa).

If you modify auth behavior, test against BOTH the mock AND a real API key (via `SKIP_INTEGRATION=0 make test-integration-live`).

## SSE Streaming: Packed Events on One Line

The Desearch API may send multiple SSE events on a single line without newline separators: `data: {...}data: {...}`. The parser in `cmd/completion.go` splits on `"data: "` boundaries. Only events with `"type": "text"` produce output; others (metadata, done signals, `[DONE]`) are silently skipped.

DC1-95 was attempted twice — the first SSE fix was incorrect. When modifying streaming parsing:
- Test with packed events (multiple `data: ` on one line)
- Test with the `[DONE]` sentinel
- Test with non-JSON garbage between events
- Test that only `"type": "text"` events produce output

## Default Tools: Always Send at Least One Tool

The Desearch API returns 422 when no `tools` field is sent. The `resolveTools()` chain (flag → config → default `["web"]`) must always produce a non-empty slice. DC1-91 was caused by an empty tools list being sent.

If you modify `cmd/tools.go` or `cmd/search.go`, verify that `resolveTools(nil, nil)` returns `["web"]` and that a bare `desearch "query"` (no `--tool` flag) still works.

## Tool Resolution Chain

```
1. --tool flag(s) on command line
2. default_tools in ~/.config/desearch-cli/config.toml
3. Hard-coded fallback: ["web"]
```

Never let an empty tools list reach the API. The API requires at least one tool.

## Exit Codes

| Code | Trigger | Implementation |
|------|---------|----------------|
| 3 | `errors.SystemError` (network, config permission) | `errors.Wrap()`, `errors.WrapF()` |
| 2 | `errors.UsageError` (unknown flag/command) | `errors.WrapUsage()` |
| 1 | API errors, user errors | plain error returns |

When adding new error paths, ensure they map to the correct exit code. Do not use `os.Exit` directly — return errors and let `main.go` handle the exit code dispatch.

## Coverage Gate: 75% Minimum

`make test-unit` fails if overall coverage drops below 75%. When adding new code in `cmd/` or `pkg/`, add corresponding tests. The `pkg/` package coverage is checked separately.

Run `make test-unit` locally before claiming work is done. Do not let coverage slip to unblock a merge.

## golangci-lint Must Pass

`make check` runs `golangci-lint` — it must pass before any PR is merged. Common issues:
- `go fmt ./...` fixes most formatting issues
- Unused variables or imports
- Missing error handling

Run `make lint` locally before committing.

## Makefile Targets

```
make test        # Smoke tests (no API key needed): build + help/version/docs/skill/completion
make test-unit   # Unit tests with -race and coverage (75% minimum)
make test-integration  # Integration tests via httptest.Server + exec.Command (build tag: integration)
make test-integration-live  # Live API tests (requires DESEARCH_API_KEY env var)
make check       # fmt + lint + test + test-unit (CI gate)
```

Always run `make check` before merging. `make test` alone is insufficient.

## SSE Streaming Output — Write Chunks Immediately

In `runCompletion()`, chunks must be written immediately via `StreamingFormatter.WriteChunk()` — do not buffer streaming output. Use `os.Stdout.Sync()` after each write for flush.

## Config File Permissions

Config is written with mode 0600 and config dir with 0700. If you modify `pkg/auth/api_key.go`, verify that `SaveConfig` creates directories with 0700 and files with 0600. Test with `make test-integration` using a readonly HOME to verify permission errors produce exit code 3.

## jq Filter Requires --json or --no-ai

`--jq` only works when output is JSON (via `--json` or `--no-ai`). Dry-run always outputs JSON so `--jq` is also allowed with `--dry-run`. Validation in `runSearch` enforces this.
