# Contributing

Contributions are welcome, especially reproducible diagnostics for real TOTVS Protheus environments.

Please keep new tools:

1. read-only by default;
2. narrowly scoped;
3. structured in their output;
4. safe to run in production-like environments;
5. independent from the LLM provider.

Before opening a pull request:

```bash
gofmt -w ./cmd ./internal
go test ./...
```

Never include customer data, credentials, internal hostnames, production IP addresses, or proprietary source code in examples/tests.
