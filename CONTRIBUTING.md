# Contributing

## Development

```bash
# Clone
git clone https://github.com/roboalchemist/desearch-cli.git
cd desearch-cli

# Build
make build

# Run tests
make check
```

## Testing

```bash
make test              # Smoke tests (no API key needed)
make test-unit         # Unit tests with coverage
make test-integration  # Integration tests via mock server
```

For live API tests, set `DESEARCH_API_KEY`:
```bash
DESEARCH_API_KEY=your_key make test-integration-live
```

## Pull Requests

1. Fork the repository
2. Create a feature branch: `git checkout -b my-feature`
3. Make your changes and run `make check`
4. Submit a pull request

## Reporting Issues

Open an issue at https://github.com/roboalchemist/desearch-cli/issues
