package catalog

import (
	"strings"
	"testing"
)

func TestCryptoFullNodeOperationalEvidence_RequiresRealConformanceWithoutRewardClaim(t *testing.T) {
	doc := CryptoOperationalEvidenceDocument{
		ProtocolVersion: CryptoOperationalEvidenceProtocolVersion,
		PluginID:        "workflow-plugin-crypto",
		Chain:           "bch",
		Role:            CryptoRoleFullNode,
		FullNode: &CryptoFullNodeOperationalEvidence{
			UpstreamClientName:        "bitcoind",
			UpstreamClientVersion:     "Bitcoin Cash Node Daemon version v28.0.0",
			RuntimeImageRef:           "ghcr.io/example/bchn@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			RuntimeImageDigest:        "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			DurableDataRef:            "volume://agents/agent-1/bch",
			ServiceHealthReceiptRef:   "artifact://tasks/task-1/service-health",
			PeerPolicyEvidenceRef:     "artifact://tasks/task-1/peer-policy",
			RPCPolicy:                 CryptoRPCPolicyPrivateOnly,
			ProtocolNativeRewardProof: false,
		},
	}
	if err := doc.Validate(); err != nil {
		t.Fatalf("full-node operational evidence should validate: %v", err)
	}

	doc.FullNode.ProtocolNativeRewardProof = true
	err := doc.Validate()
	if err == nil || !strings.Contains(err.Error(), "protocol-native reward proof") {
		t.Fatalf("full-node reward proof claim should be rejected, got %v", err)
	}
}

func TestCryptoTransactionVerifierEvidence_RequiresBoundedVerificationRefsWithoutRewardClaim(t *testing.T) {
	doc := CryptoOperationalEvidenceDocument{
		ProtocolVersion: CryptoOperationalEvidenceProtocolVersion,
		PluginID:        "workflow-plugin-crypto",
		Chain:           "btc",
		Role:            CryptoRoleTransactionVerifier,
		TransactionVerifier: &CryptoTransactionVerifierOperationalEvidence{
			RawTransactionDigest:      "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			ExpectedTxID:              "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			ComputedTxID:              "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			OutputAccountingRef:       "artifact://tasks/task-6/output-accounting",
			RuntimeReceiptRef:         "artifact://tasks/task-6/runtime-receipt",
			ProtocolNativeRewardProof: false,
		},
	}
	if err := doc.Validate(); err != nil {
		t.Fatalf("transaction verifier evidence should validate: %v", err)
	}

	doc.TransactionVerifier.ExpectedTxID = strings.ToUpper(doc.TransactionVerifier.ExpectedTxID)
	if err := doc.Validate(); err != nil {
		t.Fatalf("transaction verifier txid comparison should be case-insensitive: %v", err)
	}
	doc.TransactionVerifier.ExpectedTxID = strings.ToLower(doc.TransactionVerifier.ExpectedTxID)

	doc.TransactionVerifier.ProtocolNativeRewardProof = true
	err := doc.Validate()
	if err == nil || !strings.Contains(err.Error(), "protocol-native reward proof") {
		t.Fatalf("transaction verifier reward proof claim should be rejected, got %v", err)
	}
	doc.TransactionVerifier.ProtocolNativeRewardProof = false

	doc.Chain = "ethereum"
	err = doc.Validate()
	if err == nil || !strings.Contains(err.Error(), "transaction-verifier chain") {
		t.Fatalf("unsupported transaction verifier chain should be rejected, got %v", err)
	}
}

func TestCryptoMinerEvidence_AllowsOnlyDevnetOrPoolShareTreasuryEvidence(t *testing.T) {
	doc := CryptoOperationalEvidenceDocument{
		ProtocolVersion: CryptoOperationalEvidenceProtocolVersion,
		PluginID:        "workflow-plugin-crypto",
		Chain:           "btc",
		Role:            CryptoRoleMiner,
		Miner: &CryptoMinerOperationalEvidence{
			Mode:                      CryptoMinerEvidenceModePoolShare,
			PoolShareEvidenceRef:      "artifact://tasks/task-2/pool-share",
			TreasuryRewardAddressRef:  "wallet://btc-mining/treasury",
			HashrateBudgetRef:         "artifact://tasks/task-2/hashrate-budget",
			ResourceBudgetRef:         "artifact://tasks/task-2/resource-budget",
			StaleShareAccountingRef:   "artifact://tasks/task-2/stale-share-accounting",
			ProtocolNativeRewardProof: false,
		},
	}
	if err := doc.Validate(); err != nil {
		t.Fatalf("pool-share miner evidence should validate: %v", err)
	}

	doc.Miner.Mode = "mainnet-solo"
	err := doc.Validate()
	if err == nil || !strings.Contains(err.Error(), "devnet-block or pool-share") {
		t.Fatalf("unsupported miner mode should be rejected, got %v", err)
	}
}

func TestCryptoValidatorAndProtocolRewardEvidence_RequireCustodyAndRewardContracts(t *testing.T) {
	validator := CryptoOperationalEvidenceDocument{
		ProtocolVersion: CryptoOperationalEvidenceProtocolVersion,
		PluginID:        "workflow-plugin-crypto",
		Chain:           "ethereum",
		Role:            CryptoRoleValidator,
		Validator: &CryptoValidatorOperationalEvidence{
			ValidatorDutyAttestationRef: "artifact://tasks/task-3/validator-duty",
			TreasuryCreditEvidenceRef:   "artifact://tasks/task-3/treasury-credit",
		},
	}
	err := validator.Validate()
	if err == nil || !strings.Contains(err.Error(), "custody contract") || !strings.Contains(err.Error(), "slashing-risk contract") {
		t.Fatalf("validator evidence without custody/slashing contracts should be rejected, got %v", err)
	}
	validator.Validator.CustodyContractRef = "contract://custody/ethereum/validator"
	validator.Validator.SlashingRiskContractRef = "contract://slashing-risk/ethereum/validator"
	if err := validator.Validate(); err != nil {
		t.Fatalf("validator evidence with custody/slashing contracts should validate: %v", err)
	}

	reward := CryptoOperationalEvidenceDocument{
		ProtocolVersion: CryptoOperationalEvidenceProtocolVersion,
		PluginID:        "workflow-plugin-crypto",
		Chain:           "ethereum",
		Role:            CryptoRoleProtocolReward,
		ProtocolReward: &CryptoProtocolRewardOperationalEvidence{
			ChainRewardEventEvidenceRef: "artifact://tasks/task-4/reward-event",
			TreasuryCreditEvidenceRef:   "artifact://tasks/task-4/treasury-credit",
		},
	}
	err = reward.Validate()
	if err == nil || !strings.Contains(err.Error(), "attribution policy") {
		t.Fatalf("protocol reward without attribution policy should be rejected, got %v", err)
	}
	reward.ProtocolReward.AttributionPolicyRef = "policy://crypto/ethereum/reward-attribution"
	if err := reward.Validate(); err != nil {
		t.Fatalf("protocol reward evidence with attribution policy should validate: %v", err)
	}
}

func TestEthereumTestnetValidatorRewardEvidence_RequiresRealDutyAndRewardRefs(t *testing.T) {
	doc := CryptoOperationalEvidenceDocument{
		ProtocolVersion: CryptoOperationalEvidenceProtocolVersion,
		PluginID:        "workflow-plugin-crypto",
		Chain:           "ethereum",
		Role:            CryptoRoleEthereumTestnetValidatorReward,
		EthereumTestnetValidatorReward: &CryptoEthereumTestnetValidatorRewardOperationalEvidence{
			Network:                          "hoodi",
			Testnet:                          true,
			ValidatorPubkey:                  "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			ValidatorClientVersion:           "Lighthouse/v7.1.0",
			ValidatorClientIdentityRef:       "artifact://tasks/task-7/validator-client-identity",
			SignerModeRef:                    "secret-ref://agents/agent-1/validator-signer",
			ValidatorDutyEvidenceRef:         "artifact://tasks/task-7/validator-duties",
			RewardAccrualEvidenceRef:         "artifact://tasks/task-7/reward-accrual",
			RewardAccrualStatus:              CryptoValidatorRewardStatusObserved,
			RewardDeltaGwei:                  42,
			WalletReceiptStatusRef:           "artifact://tasks/task-7/wallet-receipt-status",
			WalletReceiptStatus:              CryptoValidatorWalletReceiptPending,
			WithdrawalAddressRef:             "wallet://ethereum-testnet-validator-reward/withdrawal",
			FeeRecipientAddressRef:           "wallet://ethereum-testnet-validator-reward/fee-recipient",
			SlashingProtectionEvidenceRef:    "artifact://tasks/task-7/slashing-protection",
			RuntimeReceiptRef:                "artifact://tasks/task-7/runtime-receipt",
			ObservationWindowSeconds:         96,
			SourceStateDigest:                "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			FixtureMode:                      true,
			ProtocolNativeRewardProof:        true,
			ValueBearingMainnetFundsInvolved: false,
		},
	}
	if err := doc.Validate(); err != nil {
		t.Fatalf("ethereum testnet validator reward evidence should validate: %v", err)
	}

	doc.EthereumTestnetValidatorReward.Testnet = false
	err := doc.Validate()
	if err == nil || !strings.Contains(err.Error(), "testnet") {
		t.Fatalf("mainnet validator reward evidence should be rejected, got %v", err)
	}
	doc.EthereumTestnetValidatorReward.Testnet = true

	doc.EthereumTestnetValidatorReward.RewardAccrualStatus = "unobserved"
	err = doc.Validate()
	if err == nil || !strings.Contains(err.Error(), "reward_accrual_status") {
		t.Fatalf("missing reward accrual should be rejected, got %v", err)
	}
	doc.EthereumTestnetValidatorReward.RewardAccrualStatus = CryptoValidatorRewardStatusObserved

	doc.EthereumTestnetValidatorReward.ValueBearingMainnetFundsInvolved = true
	err = doc.Validate()
	if err == nil || !strings.Contains(err.Error(), "mainnet funds") {
		t.Fatalf("mainnet funds should be rejected, got %v", err)
	}
	doc.EthereumTestnetValidatorReward.ValueBearingMainnetFundsInvolved = false

	doc.ExternalRefs = map[string]string{"validator_private_key": "secret://agents/agent-1/key"}
	err = doc.Validate()
	if err == nil || !strings.Contains(err.Error(), "raw secret field") {
		t.Fatalf("raw validator key fields should be rejected, got %v", err)
	}
}

func TestCryptoOperationalEvidence_RejectsRawSecretFieldsAndValues(t *testing.T) {
	doc := CryptoOperationalEvidenceDocument{
		ProtocolVersion: CryptoOperationalEvidenceProtocolVersion,
		PluginID:        "workflow-plugin-crypto",
		Chain:           "btc",
		Role:            CryptoRoleMiner,
		ExternalRefs: map[string]string{
			"client_secret_ref": "secret://providers/mining/client-secret",
			"pool_secret_ref":   "secret://providers/mining/pool-secret",
		},
		Miner: &CryptoMinerOperationalEvidence{
			Mode:                     CryptoMinerEvidenceModeDevnetBlock,
			DevnetBlockEvidenceRef:   "artifact://tasks/task-5/devnet-block",
			TreasuryRewardAddressRef: "wallet://btc-mining/treasury",
			HashrateBudgetRef:        "artifact://tasks/task-5/hashrate-budget",
			ResourceBudgetRef:        "artifact://tasks/task-5/resource-budget",
			StaleShareAccountingRef:  "artifact://tasks/task-5/stale-share-accounting",
		},
	}
	if err := doc.Validate(); err != nil {
		t.Fatalf("secret refs should validate: %v", err)
	}

	doc.ExternalRefs["client_secret"] = "secret://providers/mining/client-secret"
	err := doc.Validate()
	if err == nil || !strings.Contains(err.Error(), "raw secret field") {
		t.Fatalf("raw secret field name should be rejected, got %v", err)
	}
	delete(doc.ExternalRefs, "client_secret")

	doc.ExternalRefs["provider_auth"] = "Bearer abc123"
	err = doc.Validate()
	if err == nil || !strings.Contains(err.Error(), "raw secret value") {
		t.Fatalf("raw secret value should be rejected, got %v", err)
	}
}
