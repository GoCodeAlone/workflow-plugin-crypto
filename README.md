# workflow-plugin-crypto

Public Workflow plugin for crypto network provider catalog metadata used by
`workflow-compute`. The catalog package intentionally avoids private module
dependencies so this public plugin can build on public CI and be imported by
future public tooling.

The plugin currently owns provider contracts for:

- BTC full node
- BCH full node
- Ethereum full node
- Ethereum testnet validator reward proof

Each profile declares the provider contract, network product shape, storage and
network guidance, treasury reward routing, and upstream-client image policy.

The public catalog also exposes versioned operational evidence contracts:

- full-node evidence validates real-client/runtime identity, durable data,
  service-health, peer-policy, and private-RPC proof without claiming
  protocol-native rewards;
- Ethereum testnet validator reward evidence validates retained-agent validator
  duty execution, signer-mode refs, slashing protection refs, reward accrual,
  wallet receipt status, and source-state hashes while rejecting mainnet funds
  and raw key material;
- miner evidence is limited to devnet block or pool-share evidence with
  treasury routing, resource budgets, and stale-share accounting fields;
- validator evidence requires custody and slashing-risk contract refs before it
  can validate; and
- protocol-native reward evidence requires chain reward event, treasury credit,
  and attribution-policy refs.

Generic mainnet-capable mining, validator, and protocol-native reward roles
remain deferred for product activation until host applications consume those
contracts and supply the remaining custody, slashing-risk, payout attribution,
and treasury-credit controls. Evidence payloads reject raw secret field names
and obvious raw secret values; callers must pass `secret://...` or other
host-managed refs instead.

The Ethereum validator reward runtime component is:

```text
provider://workflow-plugin-crypto/ethereum/validator-reward-runner
```

It accepts generic workflow-compute provider envelopes for
`run_validator_reward_proof`, writes `validator_reward_evidence.json` for
fixture observations with an explicit `fixture_mode` marker, and writes
`validator_reward_blocker.json` when required live testnet validator inputs are
unavailable. Staging acceptance must reject `fixture_mode` artifacts.

Live testnet observations require a Beacon API URL and active validator pubkey.
The operation input may provide `beacon_api_url` and `validator_pubkey`
directly. For retained agents, the runner also accepts those same live inputs
from the dynamic provider envelope or process environment using
`WORKFLOW_CRYPTO_ETHEREUM_VALIDATOR_REWARD_BEACON_API_URL` and
`WORKFLOW_CRYPTO_ETHEREUM_VALIDATOR_REWARD_VALIDATOR_PUBKEY`. Signer, wallet,
and slashing-protection values remain refs in the operation input; raw validator
keys are not accepted.

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
