package catalog

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

const (
	CryptoOperationalEvidenceProtocolVersion = "crypto-operational-evidence/v1"
	CryptoRPCPolicyPrivateOnly               = "private-only"
	CryptoMinerEvidenceModeDevnetBlock       = "devnet-block"
	CryptoMinerEvidenceModePoolShare         = "pool-share"
)

type CryptoOperationalEvidenceContract struct {
	Role                    string   `json:"role"`
	ProofMode               string   `json:"proof_mode"`
	ActivationStatus        string   `json:"activation_status"`
	SupportedModes          []string `json:"supported_modes,omitempty"`
	RequiredRefs            []string `json:"required_refs"`
	ProtocolRewardProof     bool     `json:"protocol_reward_proof"`
	RequiresCustodyContract bool     `json:"requires_custody_contract,omitempty"`
	RequiresSlashingRisk    bool     `json:"requires_slashing_risk,omitempty"`
	DeferredReason          string   `json:"deferred_reason,omitempty"`
}

type CryptoOperationalEvidenceDocument struct {
	ProtocolVersion     string                                        `json:"protocol_version"`
	PluginID            string                                        `json:"plugin_id"`
	Chain               string                                        `json:"chain"`
	Role                string                                        `json:"role"`
	ExternalRefs        map[string]string                             `json:"external_refs,omitempty"`
	FullNode            *CryptoFullNodeOperationalEvidence            `json:"full_node,omitempty"`
	TransactionVerifier *CryptoTransactionVerifierOperationalEvidence `json:"transaction_verifier,omitempty"`
	Miner               *CryptoMinerOperationalEvidence               `json:"miner,omitempty"`
	Validator           *CryptoValidatorOperationalEvidence           `json:"validator,omitempty"`
	ProtocolReward      *CryptoProtocolRewardOperationalEvidence      `json:"protocol_reward,omitempty"`
}

type CryptoFullNodeOperationalEvidence struct {
	UpstreamClientName        string `json:"upstream_client_name"`
	UpstreamClientVersion     string `json:"upstream_client_version"`
	RuntimeImageRef           string `json:"runtime_image_ref"`
	RuntimeImageDigest        string `json:"runtime_image_digest"`
	DurableDataRef            string `json:"durable_data_ref"`
	ServiceHealthReceiptRef   string `json:"service_health_receipt_ref"`
	PeerPolicyEvidenceRef     string `json:"peer_policy_evidence_ref"`
	RPCPolicy                 string `json:"rpc_policy"`
	ProtocolNativeRewardProof bool   `json:"protocol_native_reward_proof"`
}

type CryptoTransactionVerifierOperationalEvidence struct {
	RawTransactionDigest      string `json:"raw_transaction_digest"`
	ExpectedTxID              string `json:"expected_txid"`
	ComputedTxID              string `json:"computed_txid"`
	OutputAccountingRef       string `json:"output_accounting_ref"`
	RuntimeReceiptRef         string `json:"runtime_receipt_ref"`
	ProtocolNativeRewardProof bool   `json:"protocol_native_reward_proof"`
}

type CryptoMinerOperationalEvidence struct {
	Mode                      string `json:"mode"`
	PoolShareEvidenceRef      string `json:"pool_share_evidence_ref,omitempty"`
	DevnetBlockEvidenceRef    string `json:"devnet_block_evidence_ref,omitempty"`
	TreasuryRewardAddressRef  string `json:"treasury_reward_address_ref"`
	HashrateBudgetRef         string `json:"hashrate_budget_ref"`
	ResourceBudgetRef         string `json:"resource_budget_ref"`
	StaleShareAccountingRef   string `json:"stale_share_accounting_ref"`
	ProtocolNativeRewardProof bool   `json:"protocol_native_reward_proof"`
}

type CryptoValidatorOperationalEvidence struct {
	ValidatorDutyAttestationRef string `json:"validator_duty_attestation_ref"`
	CustodyContractRef          string `json:"custody_contract_ref"`
	SlashingRiskContractRef     string `json:"slashing_risk_contract_ref"`
	TreasuryCreditEvidenceRef   string `json:"treasury_credit_evidence_ref"`
}

type CryptoProtocolRewardOperationalEvidence struct {
	ChainRewardEventEvidenceRef string `json:"chain_reward_event_evidence_ref"`
	TreasuryCreditEvidenceRef   string `json:"treasury_credit_evidence_ref"`
	AttributionPolicyRef        string `json:"attribution_policy_ref"`
}

func CryptoOperationalEvidenceContracts() []CryptoOperationalEvidenceContract {
	return []CryptoOperationalEvidenceContract{
		{
			Role:             CryptoRoleFullNode,
			ProofMode:        CryptoProofModeOperationalEvidence,
			ActivationStatus: CryptoRoleStatusSupported,
			RequiredRefs: []string{
				"upstream_client_version",
				"runtime_image_digest",
				"durable_data_ref",
				"service_health_receipt_ref",
				"peer_policy_evidence_ref",
				"rpc_policy=private-only",
			},
		},
		{
			Role:             CryptoRoleTransactionVerifier,
			ProofMode:        CryptoProofModeTransactionVerify,
			ActivationStatus: CryptoRoleStatusSupported,
			RequiredRefs: []string{
				"raw_transaction_digest",
				"expected_txid",
				"computed_txid",
				"output_accounting",
				"runtime_receipt_ref",
			},
		},
		{
			Role:             CryptoRoleMiner,
			ProofMode:        CryptoProofModeMiningShare,
			ActivationStatus: CryptoRoleStatusDeferred,
			SupportedModes:   []string{CryptoMinerEvidenceModeDevnetBlock, CryptoMinerEvidenceModePoolShare},
			RequiredRefs: []string{
				"devnet_block_evidence_ref or pool_share_evidence_ref",
				"treasury_reward_address_ref",
				"hashrate_budget_ref",
				"resource_budget_ref",
				"stale_share_accounting_ref",
			},
			DeferredReason: "live mining activation requires pool/hardware eligibility, stale-share, and treasury-credit accounting contracts",
		},
		{
			Role:                    CryptoRoleValidator,
			ProofMode:               CryptoProofModeValidatorDuty,
			ActivationStatus:        CryptoRoleStatusDeferred,
			RequiresCustodyContract: true,
			RequiresSlashingRisk:    true,
			RequiredRefs: []string{
				"validator_duty_attestation_ref",
				"custody_contract_ref",
				"slashing_risk_contract_ref",
				"treasury_credit_evidence_ref",
			},
			DeferredReason: "validator activation requires custody, signing authority, slashing-risk, and reward-credit contracts",
		},
		{
			Role:             CryptoRoleProtocolReward,
			ProofMode:        CryptoProofModeProtocolReward,
			ActivationStatus: CryptoRoleStatusDeferred,
			RequiredRefs: []string{
				"chain_reward_event_evidence_ref",
				"treasury_credit_evidence_ref",
				"attribution_policy_ref",
			},
			ProtocolRewardProof: true,
			DeferredReason:      "protocol reward activation requires chain reward event, treasury credit, and attribution contracts",
		},
	}
}

func (d CryptoOperationalEvidenceDocument) Validate() error {
	var errs []error
	if d.ProtocolVersion != CryptoOperationalEvidenceProtocolVersion {
		errs = append(errs, fmt.Errorf("protocol_version must be %q", CryptoOperationalEvidenceProtocolVersion))
	}
	if d.PluginID != cryptoPluginID {
		errs = append(errs, fmt.Errorf("plugin_id must be %q", cryptoPluginID))
	}
	role, ok := CryptoRoleProfileByID(d.Role)
	if !ok {
		errs = append(errs, fmt.Errorf("role %q is unsupported", d.Role))
	}
	if role.ID == CryptoRoleTransactionVerifier {
		if _, ok := CryptoTransactionVerifierProfile(d.Chain); !ok {
			errs = append(errs, fmt.Errorf("transaction-verifier chain %q is unsupported", d.Chain))
		}
	} else if _, ok := CryptoNetworkProfile(d.Chain); !ok {
		errs = append(errs, fmt.Errorf("chain %q is unsupported", d.Chain))
	}
	if err := rejectRawSecrets(d); err != nil {
		errs = append(errs, err)
	}
	switch role.ID {
	case CryptoRoleFullNode:
		errs = append(errs, validateFullNodeEvidence(d.FullNode)...)
	case CryptoRoleTransactionVerifier:
		errs = append(errs, validateTransactionVerifierEvidence(d.TransactionVerifier)...)
	case CryptoRoleMiner:
		errs = append(errs, validateMinerEvidence(d.Miner)...)
	case CryptoRoleValidator:
		errs = append(errs, validateValidatorEvidence(d.Validator)...)
	case CryptoRoleProtocolReward:
		errs = append(errs, validateProtocolRewardEvidence(d.ProtocolReward)...)
	default:
		errs = append(errs, fmt.Errorf("role %q has no evidence validator", role.ID))
	}
	return errors.Join(errs...)
}

func validateFullNodeEvidence(e *CryptoFullNodeOperationalEvidence) []error {
	if e == nil {
		return []error{errors.New("full_node evidence is required")}
	}
	var errs []error
	requireNonEmpty(&errs, "upstream_client_name", e.UpstreamClientName)
	requireNonEmpty(&errs, "upstream_client_version", e.UpstreamClientVersion)
	requireNonEmpty(&errs, "durable_data_ref", e.DurableDataRef)
	requireNonEmpty(&errs, "service_health_receipt_ref", e.ServiceHealthReceiptRef)
	requireNonEmpty(&errs, "peer_policy_evidence_ref", e.PeerPolicyEvidenceRef)
	if !strings.Contains(e.RuntimeImageRef, "@sha256:") {
		errs = append(errs, errors.New("runtime_image_ref must be digest-pinned"))
	}
	if !sha256DigestPattern.MatchString(e.RuntimeImageDigest) {
		errs = append(errs, errors.New("runtime_image_digest must be sha256:<64 hex>"))
	}
	if e.RPCPolicy != CryptoRPCPolicyPrivateOnly {
		errs = append(errs, fmt.Errorf("rpc_policy must be %q", CryptoRPCPolicyPrivateOnly))
	}
	if e.ProtocolNativeRewardProof {
		errs = append(errs, errors.New("full-node operational evidence must not claim protocol-native reward proof"))
	}
	return errs
}

func validateTransactionVerifierEvidence(e *CryptoTransactionVerifierOperationalEvidence) []error {
	if e == nil {
		return []error{errors.New("transaction_verifier evidence is required")}
	}
	var errs []error
	if !sha256DigestPattern.MatchString(e.RawTransactionDigest) {
		errs = append(errs, errors.New("raw_transaction_digest must be sha256:<64 hex>"))
	}
	if !txidPattern.MatchString(e.ExpectedTxID) {
		errs = append(errs, errors.New("expected_txid must be 64 hex characters"))
	}
	if !txidPattern.MatchString(e.ComputedTxID) {
		errs = append(errs, errors.New("computed_txid must be 64 hex characters"))
	}
	if !strings.EqualFold(e.ExpectedTxID, e.ComputedTxID) {
		errs = append(errs, errors.New("computed_txid must match expected_txid"))
	}
	requireNonEmpty(&errs, "output_accounting_ref", e.OutputAccountingRef)
	requireNonEmpty(&errs, "runtime_receipt_ref", e.RuntimeReceiptRef)
	if e.ProtocolNativeRewardProof {
		errs = append(errs, errors.New("transaction verifier evidence must not claim protocol-native reward proof"))
	}
	return errs
}

func validateMinerEvidence(e *CryptoMinerOperationalEvidence) []error {
	if e == nil {
		return []error{errors.New("miner evidence is required")}
	}
	var errs []error
	switch e.Mode {
	case CryptoMinerEvidenceModeDevnetBlock:
		requireNonEmpty(&errs, "devnet_block_evidence_ref", e.DevnetBlockEvidenceRef)
	case CryptoMinerEvidenceModePoolShare:
		requireNonEmpty(&errs, "pool_share_evidence_ref", e.PoolShareEvidenceRef)
	default:
		errs = append(errs, errors.New("miner evidence mode must be devnet-block or pool-share"))
	}
	requireNonEmpty(&errs, "treasury_reward_address_ref", e.TreasuryRewardAddressRef)
	requireNonEmpty(&errs, "hashrate_budget_ref", e.HashrateBudgetRef)
	requireNonEmpty(&errs, "resource_budget_ref", e.ResourceBudgetRef)
	requireNonEmpty(&errs, "stale_share_accounting_ref", e.StaleShareAccountingRef)
	if e.ProtocolNativeRewardProof {
		errs = append(errs, errors.New("miner evidence must not claim direct protocol-native reward payout"))
	}
	return errs
}

func validateValidatorEvidence(e *CryptoValidatorOperationalEvidence) []error {
	if e == nil {
		return []error{errors.New("validator evidence is required")}
	}
	var errs []error
	requireNonEmpty(&errs, "validator_duty_attestation_ref", e.ValidatorDutyAttestationRef)
	requireNonEmpty(&errs, "custody contract", e.CustodyContractRef)
	requireNonEmpty(&errs, "slashing-risk contract", e.SlashingRiskContractRef)
	requireNonEmpty(&errs, "treasury_credit_evidence_ref", e.TreasuryCreditEvidenceRef)
	return errs
}

func validateProtocolRewardEvidence(e *CryptoProtocolRewardOperationalEvidence) []error {
	if e == nil {
		return []error{errors.New("protocol_reward evidence is required")}
	}
	var errs []error
	requireNonEmpty(&errs, "chain_reward_event_evidence_ref", e.ChainRewardEventEvidenceRef)
	requireNonEmpty(&errs, "treasury_credit_evidence_ref", e.TreasuryCreditEvidenceRef)
	requireNonEmpty(&errs, "attribution policy", e.AttributionPolicyRef)
	return errs
}

func requireNonEmpty(errs *[]error, name, value string) {
	if strings.TrimSpace(value) == "" {
		*errs = append(*errs, fmt.Errorf("%s is required", name))
	}
}

var sha256DigestPattern = regexp.MustCompile(`^sha256:[a-fA-F0-9]{64}$`)
var txidPattern = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)

func rejectRawSecrets(v any) error {
	var errs []error
	walkRawSecrets(reflect.ValueOf(v), &errs)
	return errors.Join(errs...)
}

func walkRawSecrets(v reflect.Value, errs *[]error) {
	if !v.IsValid() {
		return
	}
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			name := jsonFieldName(t.Field(i))
			if isRawSecretField(name) {
				*errs = append(*errs, fmt.Errorf("raw secret field %q is not allowed; use a secret ref", name))
			}
			walkRawSecrets(v.Field(i), errs)
		}
	case reflect.Map:
		iter := v.MapRange()
		for iter.Next() {
			key := fmt.Sprint(iter.Key().Interface())
			if isRawSecretField(key) {
				*errs = append(*errs, fmt.Errorf("raw secret field %q is not allowed; use a secret ref", key))
			}
			walkRawSecrets(iter.Value(), errs)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			walkRawSecrets(v.Index(i), errs)
		}
	case reflect.String:
		if isRawSecretValue(v.String()) {
			*errs = append(*errs, errors.New("raw secret value is not allowed; use a secret ref"))
		}
	}
}

func jsonFieldName(field reflect.StructField) string {
	name := strings.Split(field.Tag.Get("json"), ",")[0]
	if name == "" {
		name = field.Name
	}
	return name
}

func isRawSecretField(name string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" || strings.HasSuffix(name, "_ref") {
		return false
	}
	switch name {
	case "client_secret", "signing_secret", "private_key", "seed_phrase", "mnemonic", "bearer_token", "webhook_secret", "pool_secret":
		return true
	default:
		return false
	}
}

func isRawSecretValue(value string) bool {
	value = strings.TrimSpace(value)
	lower := strings.ToLower(value)
	upper := strings.ToUpper(value)
	return strings.HasPrefix(lower, "bearer ") ||
		strings.HasPrefix(lower, "sk_live_") ||
		strings.HasPrefix(lower, "sk_test_") ||
		strings.HasPrefix(value, "ghp_") ||
		strings.Contains(upper, "PRIVATE KEY-----")
}

func (c CryptoOperationalEvidenceContract) Validate() error {
	var errs []error
	if _, ok := CryptoRoleProfileByID(c.Role); !ok {
		errs = append(errs, fmt.Errorf("role %q is unsupported", c.Role))
	}
	if strings.TrimSpace(c.ProofMode) == "" {
		errs = append(errs, errors.New("proof_mode is required"))
	}
	if c.ActivationStatus != CryptoRoleStatusSupported && c.ActivationStatus != CryptoRoleStatusDeferred {
		errs = append(errs, errors.New("activation_status is invalid"))
	}
	if len(c.RequiredRefs) == 0 {
		errs = append(errs, errors.New("required_refs is required"))
	}
	if c.ActivationStatus == CryptoRoleStatusDeferred && strings.TrimSpace(c.DeferredReason) == "" {
		errs = append(errs, errors.New("deferred_reason is required for deferred evidence contracts"))
	}
	return errors.Join(errs...)
}
