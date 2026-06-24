package validatorreward

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/GoCodeAlone/workflow-plugin-compute-core/protocol"
	"github.com/GoCodeAlone/workflow-plugin-crypto/catalog"
)

const (
	ProviderName             = "workflow-plugin-crypto"
	ExecutorProvider         = "ethereum-testnet-validator-reward"
	WorkloadKind             = "provider"
	Operation                = "run_validator_reward_proof"
	EvidenceArtifact         = "validator_reward_evidence.json"
	BlockerArtifact          = "validator_reward_blocker.json"
	ComputeProtocolVersion   = "compute.v1alpha1"
	artifactMode             = 0o644
	defaultObservationWindow = 60
)

var Version = "0.1.0"

var sleep = time.Sleep

type dynamicEnvelope struct {
	ProtocolVersion string                  `json:"protocol_version"`
	TaskID          string                  `json:"task_id"`
	LeaseID         string                  `json:"lease_id"`
	WorkloadKind    protocol.WorkloadKind   `json:"workload_kind,omitempty"`
	ProviderConfig  protocol.ProviderConfig `json:"provider_config"`
	Operation       string                  `json:"operation"`
	Input           json.RawMessage         `json:"input"`
	Env             map[string]string       `json:"env,omitempty"`
	Limits          protocol.ResourceLimits `json:"limits,omitzero"`
}

type Request struct {
	Workload Workload `json:"workload"`
}

type Workload struct {
	Chain                      string  `json:"chain"`
	Network                    string  `json:"network"`
	ValidatorClientIdentityRef string  `json:"validator_client_identity_ref"`
	SignerModeRef              string  `json:"signer_mode_ref"`
	WithdrawalAddressRef       string  `json:"withdrawal_address_ref"`
	FeeRecipientAddressRef     string  `json:"fee_recipient_address_ref"`
	SlashingProtectionRef      string  `json:"slashing_protection_ref"`
	BeaconRPCRef               string  `json:"beacon_rpc_ref,omitempty"`
	BeaconAPIURL               string  `json:"beacon_api_url,omitempty"`
	ValidatorPubkey            string  `json:"validator_pubkey,omitempty"`
	ObservationWindowSeconds   int     `json:"observation_window_seconds,omitempty"`
	Fixture                    Fixture `json:"fixture,omitempty"`
}

type Fixture struct {
	ValidatorPubkey        string `json:"validator_pubkey,omitempty"`
	ValidatorClientVersion string `json:"validator_client_version,omitempty"`
	DutyEvidenceRef        string `json:"duty_evidence_ref,omitempty"`
	RewardAccrualRef       string `json:"reward_accrual_ref,omitempty"`
	WalletReceiptStatusRef string `json:"wallet_receipt_status_ref,omitempty"`
	WalletReceiptStatus    string `json:"wallet_receipt_status,omitempty"`
	RewardDeltaGwei        int64  `json:"reward_delta_gwei,omitempty"`
	SourceState            string `json:"source_state,omitempty"`
}

type providerResult struct {
	Artifacts []string `json:"artifacts"`
}

type probeResponse struct {
	Provider              string   `json:"provider"`
	ProviderVersion       string   `json:"provider_version"`
	Status                string   `json:"status"`
	WorkloadKind          string   `json:"workload_kind"`
	ExecutorProvider      string   `json:"executor_provider"`
	ExecutionSecurityTier string   `json:"execution_security_tier"`
	ProofTier             string   `json:"proof_tier"`
	Operation             string   `json:"operation"`
	SupportedNetworks     []string `json:"supported_networks"`
	RuntimeTools          []string `json:"runtime_tools"`
}

type blockerDocument struct {
	ProtocolVersion string    `json:"protocol_version"`
	PluginID        string    `json:"plugin_id"`
	ProviderID      string    `json:"provider_id"`
	Operation       string    `json:"operation"`
	Status          string    `json:"status"`
	Reason          string    `json:"reason"`
	ObservedAt      time.Time `json:"observed_at"`
}

type liveObservation struct {
	ValidatorPubkey        string
	ValidatorClientVersion string
	DutyEvidenceRef        string
	RewardAccrualRef       string
	WalletReceiptStatusRef string
	WalletReceiptStatus    string
	RewardDeltaGwei        int64
	SourceState            string
	FixtureMode            bool
}

type nodeVersionResponse struct {
	Data struct {
		Version string `json:"version"`
	} `json:"data"`
}

type validatorResponse struct {
	Data struct {
		Balance string `json:"balance"`
	} `json:"data"`
}

func WriteProbe(w io.Writer) error {
	resp := probeResponse{
		Provider:              ProviderName,
		ProviderVersion:       Version,
		Status:                "supported",
		WorkloadKind:          WorkloadKind,
		ExecutorProvider:      ExecutorProvider,
		ExecutionSecurityTier: string(catalog.ExecutionSandboxedContainer),
		ProofTier:             string(catalog.ProofArtifactHash),
		Operation:             Operation,
		SupportedNetworks:     []string{"hoodi", "sepolia"},
		RuntimeTools:          []string{"ethereum-validator-client"},
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(resp)
}

func Main(args []string, stdout, stderr io.Writer, stdin ...io.Reader) int {
	fs := flag.NewFlagSet("ethereum-validator-reward-runner", flag.ContinueOnError)
	fs.SetOutput(stderr)
	requestPath := fs.String("request", "", "path to validator reward request JSON")
	outputPath := fs.String("output", "", "path to write validator reward evidence JSON")
	probe := fs.Bool("probe", false, "print provider capability probe")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *probe {
		if err := WriteProbe(stdout); err != nil {
			fmt.Fprintf(stderr, "probe: %v\n", err)
			return 1
		}
		return 0
	}
	if *requestPath == "" && *outputPath == "" {
		input := io.Reader(os.Stdin)
		if len(stdin) > 0 && stdin[0] != nil {
			input = stdin[0]
		}
		if err := runDynamic(input, stdout); err != nil {
			fmt.Fprintf(stderr, "ethereum validator reward: %v\n", err)
			return 1
		}
		return 0
	}
	if *requestPath == "" || *outputPath == "" {
		fmt.Fprintln(stderr, "--request and --output are both required")
		return 2
	}
	if err := runRequest(*requestPath, *outputPath); err != nil {
		fmt.Fprintf(stderr, "ethereum validator reward: %v\n", err)
		return 1
	}
	return 0
}

func runDynamic(r io.Reader, stdout io.Writer) error {
	env, err := readDynamicEnvelope(r)
	if err != nil {
		return err
	}
	if err := validateDynamicEnvelope(env); err != nil {
		return err
	}
	workload, err := decodeWorkload(env.Input)
	if err != nil {
		return err
	}
	if err := runWorkload(workload, EvidenceArtifact); err != nil {
		if writeErr := writeBlocker(BlockerArtifact, err); writeErr != nil {
			return errors.Join(err, writeErr)
		}
		return err
	}
	return json.NewEncoder(stdout).Encode(providerResult{Artifacts: []string{EvidenceArtifact}})
}

func runRequest(requestPath, outputPath string) error {
	data, err := os.ReadFile(requestPath)
	if err != nil {
		return fmt.Errorf("read request: %w", err)
	}
	var req Request
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		return fmt.Errorf("decode request: %w", err)
	}
	var extra struct{}
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("decode request: multiple JSON values")
	}
	return runWorkload(req.Workload, outputPath)
}

func readDynamicEnvelope(r io.Reader) (dynamicEnvelope, error) {
	var env dynamicEnvelope
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&env); err != nil {
		return dynamicEnvelope{}, fmt.Errorf("decode provider envelope: %w", err)
	}
	var extra struct{}
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return dynamicEnvelope{}, errors.New("decode provider envelope: multiple JSON values")
	}
	return env, nil
}

func validateDynamicEnvelope(env dynamicEnvelope) error {
	if env.ProtocolVersion != ComputeProtocolVersion {
		return fmt.Errorf("unsupported protocol_version %q", env.ProtocolVersion)
	}
	if env.WorkloadKind != "" && env.WorkloadKind != protocol.WorkloadProvider {
		return fmt.Errorf("unsupported workload_kind %q", env.WorkloadKind)
	}
	if err := env.ProviderConfig.Validate(); err != nil {
		return fmt.Errorf("provider_config: %w", err)
	}
	if env.ProviderConfig.PluginID != ProviderName ||
		env.ProviderConfig.ProviderID != ExecutorProvider ||
		env.ProviderConfig.ContractID != "crypto.ethereum-testnet-validator-reward.v1" ||
		env.ProviderConfig.Version != "v1.0.0" {
		return errors.New("provider_config does not match ethereum testnet validator reward v1")
	}
	if env.Operation != Operation {
		return fmt.Errorf("unsupported operation %q", env.Operation)
	}
	return nil
}

func decodeWorkload(raw json.RawMessage) (Workload, error) {
	var workload Workload
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&workload); err != nil {
		return Workload{}, fmt.Errorf("decode operation input: %w", err)
	}
	var extra struct{}
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return Workload{}, errors.New("decode operation input: multiple JSON values")
	}
	return workload, nil
}

func runWorkload(workload Workload, outputPath string) error {
	doc, err := buildEvidence(workload)
	if err != nil {
		return err
	}
	if err := doc.Validate(); err != nil {
		return fmt.Errorf("validate evidence: %w", err)
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("encode evidence: %w", err)
	}
	return os.WriteFile(outputPath, append(data, '\n'), artifactMode)
}

func buildEvidence(workload Workload) (catalog.CryptoOperationalEvidenceDocument, error) {
	normalizeWorkload(&workload)
	if workload.Chain != "ethereum" {
		return catalog.CryptoOperationalEvidenceDocument{}, fmt.Errorf("chain must be ethereum")
	}
	switch workload.Network {
	case "hoodi", "sepolia":
	default:
		return catalog.CryptoOperationalEvidenceDocument{}, fmt.Errorf("network must be hoodi or sepolia testnet")
	}
	window := workload.ObservationWindowSeconds
	if window <= 0 {
		window = defaultObservationWindow
	}
	observation, err := observe(workload, window)
	if err != nil {
		return catalog.CryptoOperationalEvidenceDocument{}, err
	}
	sourceDigest := sha256.Sum256([]byte(observation.SourceState))
	doc := catalog.CryptoOperationalEvidenceDocument{
		ProtocolVersion: catalog.CryptoOperationalEvidenceProtocolVersion,
		PluginID:        ProviderName,
		Chain:           "ethereum",
		Role:            catalog.CryptoRoleEthereumTestnetValidatorReward,
		EthereumTestnetValidatorReward: &catalog.CryptoEthereumTestnetValidatorRewardOperationalEvidence{
			Network:                          workload.Network,
			Testnet:                          true,
			ValidatorPubkey:                  observation.ValidatorPubkey,
			ValidatorClientVersion:           observation.ValidatorClientVersion,
			ValidatorClientIdentityRef:       workload.ValidatorClientIdentityRef,
			SignerModeRef:                    workload.SignerModeRef,
			ValidatorDutyEvidenceRef:         observation.DutyEvidenceRef,
			RewardAccrualEvidenceRef:         observation.RewardAccrualRef,
			RewardAccrualStatus:              catalog.CryptoValidatorRewardStatusObserved,
			RewardDeltaGwei:                  observation.RewardDeltaGwei,
			WalletReceiptStatusRef:           observation.WalletReceiptStatusRef,
			WalletReceiptStatus:              observation.WalletReceiptStatus,
			WithdrawalAddressRef:             workload.WithdrawalAddressRef,
			FeeRecipientAddressRef:           workload.FeeRecipientAddressRef,
			SlashingProtectionEvidenceRef:    workload.SlashingProtectionRef,
			RuntimeReceiptRef:                "artifact://validator-reward/runtime-receipt",
			ObservationWindowSeconds:         window,
			SourceStateDigest:                "sha256:" + hex.EncodeToString(sourceDigest[:]),
			FixtureMode:                      observation.FixtureMode,
			ProtocolNativeRewardProof:        true,
			ValueBearingMainnetFundsInvolved: false,
		},
	}
	return doc, nil
}

func observe(workload Workload, windowSeconds int) (liveObservation, error) {
	if workload.Fixture != (Fixture{}) {
		if workload.Fixture.RewardDeltaGwei <= 0 {
			return liveObservation{}, errors.New("reward accrual fixture must include a positive reward_delta_gwei")
		}
		return liveObservation{
			ValidatorPubkey:        workload.Fixture.ValidatorPubkey,
			ValidatorClientVersion: workload.Fixture.ValidatorClientVersion,
			DutyEvidenceRef:        workload.Fixture.DutyEvidenceRef,
			RewardAccrualRef:       workload.Fixture.RewardAccrualRef,
			WalletReceiptStatusRef: workload.Fixture.WalletReceiptStatusRef,
			WalletReceiptStatus:    workload.Fixture.WalletReceiptStatus,
			RewardDeltaGwei:        workload.Fixture.RewardDeltaGwei,
			SourceState:            workload.Fixture.SourceState,
			FixtureMode:            true,
		}, nil
	}
	if strings.TrimSpace(workload.BeaconAPIURL) == "" || strings.TrimSpace(workload.ValidatorPubkey) == "" {
		return liveObservation{}, errors.New("live validator execution requires beacon_api_url and validator_pubkey, or fixture mode")
	}
	base, err := url.Parse(workload.BeaconAPIURL)
	if err != nil {
		return liveObservation{}, fmt.Errorf("beacon_api_url: %w", err)
	}
	if base.Scheme != "https" && base.Scheme != "http" {
		return liveObservation{}, errors.New("beacon_api_url must use http or https")
	}
	client := &http.Client{Timeout: 10 * time.Second}
	version, err := fetchNodeVersion(client, base)
	if err != nil {
		return liveObservation{}, err
	}
	before, err := fetchValidatorBalance(client, base, workload.ValidatorPubkey)
	if err != nil {
		return liveObservation{}, err
	}
	sleep(time.Duration(windowSeconds) * time.Second)
	after, err := fetchValidatorBalance(client, base, workload.ValidatorPubkey)
	if err != nil {
		return liveObservation{}, err
	}
	delta := after - before
	if delta <= 0 {
		return liveObservation{}, errors.New("no positive validator reward delta observed")
	}
	source := fmt.Sprintf("%s|%s|%d|%d", version, workload.ValidatorPubkey, before, after)
	return liveObservation{
		ValidatorPubkey:        workload.ValidatorPubkey,
		ValidatorClientVersion: version,
		DutyEvidenceRef:        "artifact://validator-reward/beacon-validator-state",
		RewardAccrualRef:       "artifact://validator-reward/beacon-balance-delta",
		WalletReceiptStatusRef: "artifact://validator-reward/wallet-receipt-status",
		WalletReceiptStatus:    catalog.CryptoValidatorWalletReceiptPending,
		RewardDeltaGwei:        delta,
		SourceState:            source,
	}, nil
}

func fetchNodeVersion(client *http.Client, base *url.URL) (string, error) {
	var resp nodeVersionResponse
	if err := fetchJSON(client, base, "/eth/v1/node/version", &resp); err != nil {
		return "", fmt.Errorf("node version: %w", err)
	}
	if strings.TrimSpace(resp.Data.Version) == "" {
		return "", errors.New("node version response missing version")
	}
	return resp.Data.Version, nil
}

func fetchValidatorBalance(client *http.Client, base *url.URL, pubkey string) (int64, error) {
	var resp validatorResponse
	if err := fetchJSON(client, base, "/eth/v1/beacon/states/head/validators/"+strings.TrimSpace(pubkey), &resp); err != nil {
		return 0, fmt.Errorf("validator balance: %w", err)
	}
	balance, err := strconv.ParseInt(resp.Data.Balance, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse validator balance: %w", err)
	}
	return balance, nil
}

func fetchJSON(client *http.Client, base *url.URL, path string, target any) error {
	u := *base
	u.Path = strings.TrimRight(base.Path, "/") + path
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	dec := json.NewDecoder(resp.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		return err
	}
	var extra struct{}
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("multiple JSON values")
	}
	return nil
}

func normalizeWorkload(workload *Workload) {
	workload.Chain = strings.ToLower(strings.TrimSpace(workload.Chain))
	workload.Network = strings.ToLower(strings.TrimSpace(workload.Network))
	if workload.Fixture != (Fixture{}) && workload.Fixture.WalletReceiptStatus == "" {
		workload.Fixture.WalletReceiptStatus = catalog.CryptoValidatorWalletReceiptPending
	}
}

func writeBlocker(path string, cause error) error {
	doc := blockerDocument{
		ProtocolVersion: ComputeProtocolVersion,
		PluginID:        ProviderName,
		ProviderID:      ExecutorProvider,
		Operation:       Operation,
		Status:          "blocked",
		Reason:          cause.Error(),
		ObservedAt:      time.Now().UTC(),
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("encode blocker: %w", err)
	}
	return os.WriteFile(path, append(data, '\n'), artifactMode)
}
