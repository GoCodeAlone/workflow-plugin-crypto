package validatorreward

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/GoCodeAlone/workflow-plugin-crypto/catalog"
)

func TestWriteProbeReportsValidatorRewardCapabilities(t *testing.T) {
	var out bytes.Buffer
	if err := WriteProbe(&out); err != nil {
		t.Fatalf("probe: %v", err)
	}
	var got probeResponse
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("decode probe: %v", err)
	}
	if got.ExecutorProvider != ExecutorProvider || got.Operation != Operation || got.WorkloadKind != WorkloadKind {
		t.Fatalf("probe identity drifted: %+v", got)
	}
}

func TestMainRunsFixtureBackedDynamicProviderEnvelope(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	input := `{
	  "protocol_version":"compute.v1alpha1",
	  "task_id":"task-eth-1",
	  "lease_id":"lease-eth-1",
	  "provider_config":{
	    "plugin_id":"workflow-plugin-crypto",
	    "provider_id":"ethereum-testnet-validator-reward",
	    "contract_id":"crypto.ethereum-testnet-validator-reward.v1",
	    "version":"v1.0.0",
	    "config_ref":"config://network-products/ethereum-testnet-validator-reward/ethereum-testnet-validator-reward"
	  },
	  "operation":"run_validator_reward_proof",
	  "input":{
	    "chain":"ethereum",
	    "network":"hoodi",
	    "validator_client_identity_ref":"artifact://tasks/task-eth-1/validator-client-identity",
	    "signer_mode_ref":"secret-ref://agents/agent-1/validator-signer",
	    "withdrawal_address_ref":"wallet://ethereum-testnet-validator-reward/withdrawal",
	    "fee_recipient_address_ref":"wallet://ethereum-testnet-validator-reward/fee-recipient",
	    "slashing_protection_ref":"artifact://tasks/task-eth-1/slashing-protection",
	    "observation_window_seconds":96,
	    "fixture":{
	      "validator_pubkey":"0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	      "validator_client_version":"Lighthouse/v7.1.0",
	      "duty_evidence_ref":"artifact://tasks/task-eth-1/validator-duties",
	      "reward_accrual_ref":"artifact://tasks/task-eth-1/reward-accrual",
	      "wallet_receipt_status_ref":"artifact://tasks/task-eth-1/wallet-receipt-status",
	      "wallet_receipt_status":"pending",
	      "reward_delta_gwei":42,
	      "source_state":"hoodi fixture epoch 1024"
	    }
	  }
	}`

	var stdout, stderr bytes.Buffer
	code := Main(nil, &stdout, &stderr, strings.NewReader(input))
	if code != 0 {
		t.Fatalf("runner failed with code %d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var result providerResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("decode result: %v\n%s", err, stdout.String())
	}
	if len(result.Artifacts) != 1 || result.Artifacts[0] != EvidenceArtifact {
		t.Fatalf("artifacts = %+v", result.Artifacts)
	}
	data, err := os.ReadFile(filepath.Join(dir, EvidenceArtifact))
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	var doc catalog.CryptoOperationalEvidenceDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("decode evidence: %v", err)
	}
	if err := doc.Validate(); err != nil {
		t.Fatalf("evidence should validate: %v", err)
	}
	reward := doc.EthereumTestnetValidatorReward
	if reward.ValidatorPubkey == "" ||
		reward.ObservationWindowSeconds != 96 ||
		reward.SourceStateDigest == "" ||
		!reward.FixtureMode ||
		reward.RewardDeltaGwei != 42 ||
		reward.RewardAccrualStatus != catalog.CryptoValidatorRewardStatusObserved {
		t.Fatalf("fixture evidence missing reward details: %+v", reward)
	}
	if bytes.Contains(data, []byte("PRIVATE KEY")) || bytes.Contains(data, []byte("validator_private_key")) {
		t.Fatalf("evidence leaked raw secret material: %s", string(data))
	}
}

func TestMainRunsLiveBeaconObservation(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	oldSleep := sleep
	sleep = func(time.Duration) {}
	t.Cleanup(func() { sleep = oldSleep })

	balanceCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/eth/v1/node/version":
			fmt.Fprint(w, `{"data":{"version":"Lighthouse/v7.1.0"}}`)
		case strings.HasPrefix(r.URL.Path, "/eth/v1/beacon/states/head/validators/"):
			balanceCalls++
			balance := int64(32_000_000_000)
			if balanceCalls > 1 {
				balance += 64
			}
			fmt.Fprintf(w, `{"data":{"balance":"%d"}}`, balance)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	input := `{
	  "protocol_version":"compute.v1alpha1",
	  "task_id":"task-eth-live",
	  "lease_id":"lease-eth-live",
	  "provider_config":{
	    "plugin_id":"workflow-plugin-crypto",
	    "provider_id":"ethereum-testnet-validator-reward",
	    "contract_id":"crypto.ethereum-testnet-validator-reward.v1",
	    "version":"v1.0.0",
	    "config_ref":"config://network-products/ethereum-testnet-validator-reward/ethereum-testnet-validator-reward"
	  },
	  "operation":"run_validator_reward_proof",
	  "input":{
	    "chain":"ethereum",
	    "network":"hoodi",
	    "validator_client_identity_ref":"artifact://tasks/task-eth-live/validator-client-identity",
	    "signer_mode_ref":"secret-ref://agents/agent-1/validator-signer",
	    "withdrawal_address_ref":"wallet://ethereum-testnet-validator-reward/withdrawal",
	    "fee_recipient_address_ref":"wallet://ethereum-testnet-validator-reward/fee-recipient",
	    "slashing_protection_ref":"artifact://tasks/task-eth-live/slashing-protection",
	    "beacon_api_url":"` + server.URL + `",
	    "validator_pubkey":"0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	    "observation_window_seconds":1
	  }
	}`

	var stdout, stderr bytes.Buffer
	code := Main(nil, &stdout, &stderr, strings.NewReader(input))
	if code != 0 {
		t.Fatalf("runner failed with code %d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	data, err := os.ReadFile(filepath.Join(dir, EvidenceArtifact))
	if err != nil {
		t.Fatalf("read evidence: %v", err)
	}
	var doc catalog.CryptoOperationalEvidenceDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("decode evidence: %v", err)
	}
	if err := doc.Validate(); err != nil {
		t.Fatalf("evidence should validate: %v", err)
	}
	reward := doc.EthereumTestnetValidatorReward
	if reward.FixtureMode || reward.RewardDeltaGwei != 64 || reward.ValidatorClientVersion != "Lighthouse/v7.1.0" {
		t.Fatalf("live evidence fields drifted: %+v", reward)
	}
}

func TestMainRejectsMainnetAndWritesBlocker(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	input := `{
	  "protocol_version":"compute.v1alpha1",
	  "task_id":"task-eth-2",
	  "lease_id":"lease-eth-2",
	  "provider_config":{
	    "plugin_id":"workflow-plugin-crypto",
	    "provider_id":"ethereum-testnet-validator-reward",
	    "contract_id":"crypto.ethereum-testnet-validator-reward.v1",
	    "version":"v1.0.0",
	    "config_ref":"config://network-products/ethereum-testnet-validator-reward/ethereum-testnet-validator-reward"
	  },
	  "operation":"run_validator_reward_proof",
	  "input":{
	    "chain":"ethereum",
	    "network":"mainnet",
	    "validator_client_identity_ref":"artifact://tasks/task-eth-2/validator-client-identity",
	    "signer_mode_ref":"secret-ref://agents/agent-1/validator-signer",
	    "withdrawal_address_ref":"wallet://ethereum-testnet-validator-reward/withdrawal",
	    "fee_recipient_address_ref":"wallet://ethereum-testnet-validator-reward/fee-recipient",
	    "slashing_protection_ref":"artifact://tasks/task-eth-2/slashing-protection"
	  }
	}`

	var stdout, stderr bytes.Buffer
	code := Main(nil, &stdout, &stderr, strings.NewReader(input))
	if code == 0 {
		t.Fatalf("expected mainnet rejection, stdout=%s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "testnet") {
		t.Fatalf("stderr should explain testnet rejection: %s", stderr.String())
	}
	if _, err := os.Stat(filepath.Join(dir, BlockerArtifact)); err != nil {
		t.Fatalf("blocker artifact missing: %v", err)
	}
}
