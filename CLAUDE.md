# CLAUDE.md — workflow-plugin-crypto

External gRPC plugin for crypto network provider catalog metadata consumed by
Workflow and `workflow-compute`.

## Build & Test

```sh
go build ./...
go test ./... -v -race -count=1
```

## Cross-compile

```sh
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o workflow-plugin-crypto ./cmd/workflow-plugin-crypto/
```

## Structure

- `cmd/workflow-plugin-crypto/main.go` — external plugin entrypoint
- `internal/plugin.go` — Workflow plugin manifest
- `catalog/crypto.go` — public `workflow-compute` provider catalog metadata
- `plugin.json` — registry-facing plugin manifest
- `.goreleaser.yaml` — GoReleaser v2 config for releases
- `.github/workflows/ci.yml` — build, test, vet, and plugin contract validation
- `.github/workflows/release.yml` — tagged release pipeline

## Releasing

```sh
git tag v0.1.0
git push origin v0.1.0
```
