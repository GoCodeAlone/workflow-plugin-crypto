# workflow-plugin-crypto

Public Workflow plugin for crypto network provider catalog metadata used by
`workflow-compute`. The catalog package intentionally avoids private module
dependencies so this public plugin can build on public CI and be imported by
future public tooling.

The plugin currently owns shape-only provider contracts for:

- BTC full node
- BCH full node
- Ethereum full node

Each profile declares the provider contract, network product shape, storage and
network guidance, treasury reward routing, and upstream-client image policy.
Mining, validator, and protocol-native reward roles are represented as deferred
role profiles until custody, slashing-risk, payout attribution, and evidence
contracts exist.

## Build & Test

```sh
go build ./...
go test ./... -race -count=1
```

## Release

```sh
git tag v0.1.0
git push origin v0.1.0
```

The release workflow validates `plugin.json`, builds cross-platform binaries
with GoReleaser, and verifies the runtime plugin manifest against the shipped
contract metadata.
