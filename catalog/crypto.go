package catalog

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const cryptoPluginID = "workflow-plugin-crypto"

const (
	CryptoStorageArchiveFull            = "archive-full"
	CryptoNetworkProfilePublicChainPeer = "public-chain-peer"
	CryptoRewardDestinationTreasury     = "treasury"
	CryptoProofModeOperationalEvidence  = "operational-evidence"
	CryptoProofModeMiningShare          = "mining-share"
	CryptoProofModeValidatorDuty        = "validator-duty"
	CryptoProofModeProtocolReward       = "protocol-reward"
	CryptoRoleFullNode                  = "full-node"
	CryptoRoleMiner                     = "miner"
	CryptoRoleValidator                 = "validator"
	CryptoRoleProtocolReward            = "protocol-reward"
	CryptoRoleStatusSupported           = "supported"
	CryptoRoleStatusDeferred            = "deferred"
)

type CryptoRoleMetadata struct {
	ID                       string `json:"id"`
	ShapeOnly                bool   `json:"shape_only"`
	ProtocolRewardsAssumed   bool   `json:"protocol_rewards_assumed"`
	OperationalConformanceID string `json:"operational_conformance_id,omitempty"`
}

type CryptoStorageMetadata struct {
	Mode                         string `json:"mode"`
	MinDiskBytes                 int64  `json:"min_disk_bytes"`
	MinDiskDisplay               string `json:"min_disk_display"`
	RecommendedDiskBytes         int64  `json:"recommended_disk_bytes"`
	RecommendedDiskDisplay       string `json:"recommended_disk_display"`
	GrowthMarginBytes            int64  `json:"growth_margin_bytes"`
	GrowthMarginDisplay          string `json:"growth_margin_display"`
	DurableVolumeRequired        bool   `json:"durable_volume_required"`
	PreserveOnUpdate             bool   `json:"preserve_on_update"`
	PreserveOnUninstall          bool   `json:"preserve_on_uninstall"`
	PruningSupported             bool   `json:"pruning_supported"`
	SnapshotVerificationRequired bool   `json:"snapshot_verification_required"`
}

type CryptoNetworkMetadata struct {
	ProfileID                 string `json:"profile_id"`
	PeerPort                  int    `json:"peer_port"`
	AllowedPeerPorts          []int  `json:"allowed_peer_ports,omitempty"`
	RequiresIngress           bool   `json:"requires_ingress"`
	UsesDNSSeeds              bool   `json:"uses_dns_seeds"`
	RPCPrivateOnly            bool   `json:"rpc_private_only"`
	AuditRequired             bool   `json:"audit_required"`
	MaxOutboundPeers          int    `json:"max_outbound_peers"`
	MaxOutboundBytesPerSecond int64  `json:"max_outbound_bytes_per_second"`
	KillClosesPeers           bool   `json:"kill_closes_peers"`
}

type CryptoRewardMetadata struct {
	ProtocolRewardDestination  string `json:"protocol_reward_destination"`
	TreasuryAccountID          string `json:"treasury_account_id,omitempty"`
	TreasuryWalletRef          string `json:"treasury_wallet_ref,omitempty"`
	ManagementFeeBasisPoints   int    `json:"management_fee_basis_points"`
	DirectWorkerPayout         bool   `json:"direct_worker_payout"`
	ProtocolRewardProofClaimed bool   `json:"protocol_reward_proof_claimed"`
}

type CryptoProofMetadata struct {
	Mode                      string   `json:"mode"`
	ShapeOnly                 bool     `json:"shape_only"`
	ProtocolNativeRewardProof bool     `json:"protocol_native_reward_proof"`
	EvidenceRefs              []string `json:"evidence_refs,omitempty"`
}

type CryptoImageMetadata struct {
	UpstreamClientName       string   `json:"upstream_client_name"`
	DigestPinnedRequired     bool     `json:"digest_pinned_required"`
	OperatorSuppliedRequired bool     `json:"operator_supplied_required,omitempty"`
	RecommendedImageRef      string   `json:"recommended_image_ref,omitempty"`
	KnownImageRefs           []string `json:"known_image_refs,omitempty"`
}

type CryptoRoleProfile struct {
	ID                           string   `json:"id"`
	Status                       string   `json:"status"`
	DisplayName                  string   `json:"display_name"`
	Description                  string   `json:"description"`
	ProofMode                    string   `json:"proof_mode"`
	TreasuryRequired             bool     `json:"treasury_required"`
	DirectWorkerPayout           bool     `json:"direct_worker_payout"`
	ProductCreationSupported     bool     `json:"product_creation_supported"`
	RequiresCustodyContract      bool     `json:"requires_custody_contract,omitempty"`
	RequiresSlashingRiskContract bool     `json:"requires_slashing_risk_contract,omitempty"`
	RequiredEvidence             []string `json:"required_evidence,omitempty"`
	DeferredReason               string   `json:"deferred_reason,omitempty"`
}

type CryptoProfile struct {
	Chain             string                `json:"chain"`
	ProductID         string                `json:"product_id"`
	DisplayName       string                `json:"display_name"`
	Purpose           string                `json:"purpose"`
	PoolID            string                `json:"pool_id"`
	ProviderID        string                `json:"provider_id"`
	ContractID        string                `json:"contract_id"`
	SchemaRef         string                `json:"schema_ref"`
	SchemaDigest      string                `json:"schema_digest"`
	ConfigRef         string                `json:"config_ref"`
	SettlementNetwork string                `json:"settlement_network"`
	WalletRef         string                `json:"wallet_ref"`
	MinDiskBytes      int64                 `json:"min_disk_bytes"`
	MinMemoryBytes    int64                 `json:"min_memory_bytes"`
	MinBandwidthMbps  int64                 `json:"min_bandwidth_mbps"`
	Role              CryptoRoleMetadata    `json:"role"`
	Storage           CryptoStorageMetadata `json:"storage"`
	Network           CryptoNetworkMetadata `json:"network"`
	Rewards           CryptoRewardMetadata  `json:"rewards"`
	Proof             CryptoProofMetadata   `json:"proof"`
	Image             CryptoImageMetadata   `json:"image"`
}

type CryptoProviderManifestDocument struct {
	ProtocolVersion   string                              `json:"protocol_version"`
	PluginID          string                              `json:"plugin_id"`
	Version           string                              `json:"version"`
	RoleProfiles      []CryptoRoleProfile                 `json:"role_profiles"`
	EvidenceContracts []CryptoOperationalEvidenceContract `json:"evidence_contracts"`
	Profiles          []CryptoProfile                     `json:"profiles"`
}

func CryptoProviderManifest() CryptoProviderManifestDocument {
	profiles := make([]CryptoProfile, 0, 3)
	for _, chain := range []string{"btc", "bch", "ethereum"} {
		profile, _ := CryptoNetworkProfile(chain)
		profiles = append(profiles, profile)
	}
	return CryptoProviderManifestDocument{
		ProtocolVersion:   Version,
		PluginID:          cryptoPluginID,
		Version:           "v1.0.0",
		RoleProfiles:      CryptoRoleProfiles(),
		EvidenceContracts: CryptoOperationalEvidenceContracts(),
		Profiles:          profiles,
	}
}

func CryptoRoleProfiles() []CryptoRoleProfile {
	return []CryptoRoleProfile{
		{
			ID:                       CryptoRoleFullNode,
			Status:                   CryptoRoleStatusSupported,
			DisplayName:              "Full node",
			Description:              "Runs a chain client that validates and serves chain data with durable storage and public peer networking.",
			ProofMode:                CryptoProofModeOperationalEvidence,
			TreasuryRequired:         true,
			DirectWorkerPayout:       false,
			ProductCreationSupported: true,
			RequiredEvidence: []string{
				"service health receipt",
				"upstream client version evidence",
				"durable data volume ref",
				"public-chain peer policy evidence",
			},
		},
		{
			ID:                       CryptoRoleMiner,
			Status:                   CryptoRoleStatusDeferred,
			DisplayName:              "Miner",
			Description:              "Runs mining or mining-pool work with pool/share evidence and treasury reward routing.",
			ProofMode:                CryptoProofModeMiningShare,
			TreasuryRequired:         true,
			DirectWorkerPayout:       false,
			ProductCreationSupported: false,
			RequiredEvidence: []string{
				"pool share or devnet block evidence",
				"treasury reward address configuration",
				"hashrate/resource budget evidence",
			},
			DeferredReason: "mining pool, hardware eligibility, stale-share, and treasury-credit accounting contracts are not implemented",
		},
		{
			ID:                           CryptoRoleValidator,
			Status:                       CryptoRoleStatusDeferred,
			DisplayName:                  "Validator",
			Description:                  "Performs validator duties only after custody, slashing-risk, uptime, and reward evidence contracts exist.",
			ProofMode:                    CryptoProofModeValidatorDuty,
			TreasuryRequired:             true,
			DirectWorkerPayout:           false,
			ProductCreationSupported:     false,
			RequiresCustodyContract:      true,
			RequiresSlashingRiskContract: true,
			RequiredEvidence: []string{
				"validator duty attestation",
				"slashing-risk controls",
				"staking/custody authority refs",
				"treasury credit evidence",
			},
			DeferredReason: "validator custody, key-management, and slashing-risk contracts are not implemented",
		},
		{
			ID:                       CryptoRoleProtocolReward,
			Status:                   CryptoRoleStatusDeferred,
			DisplayName:              "Protocol reward",
			Description:              "Captures chain-specific reward duties that are neither pure full-node service nor mining/validator work.",
			ProofMode:                CryptoProofModeProtocolReward,
			TreasuryRequired:         true,
			DirectWorkerPayout:       false,
			ProductCreationSupported: false,
			RequiredEvidence: []string{
				"chain-specific reward event evidence",
				"treasury credit evidence",
				"contribution attribution policy",
			},
			DeferredReason: "chain-specific reward evidence and payout attribution contracts are not implemented",
		},
	}
}

func CryptoRoleProfileByID(roleID string) (CryptoRoleProfile, bool) {
	roleID = strings.ToLower(strings.TrimSpace(roleID))
	if roleID == "" {
		roleID = CryptoRoleFullNode
	}
	for _, role := range CryptoRoleProfiles() {
		if role.ID == roleID {
			return role, true
		}
	}
	return CryptoRoleProfile{}, false
}

func CryptoProviderManifestDigest() string {
	return CanonicalHash(CryptoProviderManifest())
}

func (m CryptoProviderManifestDocument) Validate() error {
	var errs []error
	if m.ProtocolVersion != Version {
		errs = append(errs, fmt.Errorf("protocol_version must be %q", Version))
	}
	if m.PluginID != cryptoPluginID {
		errs = append(errs, fmt.Errorf("plugin_id must be %q", cryptoPluginID))
	}
	if strings.TrimSpace(m.Version) == "" || strings.ContainsAny(m.Version, " \t\r\n") {
		errs = append(errs, errors.New("version is required"))
	}
	if len(m.Profiles) == 0 {
		errs = append(errs, errors.New("profiles is required"))
	}
	if len(m.RoleProfiles) == 0 {
		errs = append(errs, errors.New("role_profiles is required"))
	}
	if len(m.EvidenceContracts) == 0 {
		errs = append(errs, errors.New("evidence_contracts is required"))
	}
	seenRoles := map[string]struct{}{}
	for i, role := range m.RoleProfiles {
		if strings.TrimSpace(role.ID) == "" {
			errs = append(errs, fmt.Errorf("role_profiles[%d].id is required", i))
		}
		if _, ok := seenRoles[role.ID]; ok {
			errs = append(errs, fmt.Errorf("role_profiles[%d].id %q is duplicated", i, role.ID))
		}
		seenRoles[role.ID] = struct{}{}
		if role.Status != CryptoRoleStatusSupported && role.Status != CryptoRoleStatusDeferred {
			errs = append(errs, fmt.Errorf("role_profiles[%d].status is invalid", i))
		}
		if strings.TrimSpace(role.ProofMode) == "" {
			errs = append(errs, fmt.Errorf("role_profiles[%d].proof_mode is required", i))
		}
		if !role.TreasuryRequired || role.DirectWorkerPayout {
			errs = append(errs, fmt.Errorf("role_profiles[%d] must use treasury routing without direct worker payout", i))
		}
		if role.Status == CryptoRoleStatusDeferred && role.ProductCreationSupported {
			errs = append(errs, fmt.Errorf("role_profiles[%d] deferred role must not support product creation", i))
		}
		if role.Status == CryptoRoleStatusDeferred && strings.TrimSpace(role.DeferredReason) == "" {
			errs = append(errs, fmt.Errorf("role_profiles[%d].deferred_reason is required", i))
		}
	}
	seenEvidenceContracts := map[string]struct{}{}
	for i, contract := range m.EvidenceContracts {
		if err := contract.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("evidence_contracts[%d]: %w", i, err))
		}
		if _, ok := seenEvidenceContracts[contract.Role]; ok {
			errs = append(errs, fmt.Errorf("evidence_contracts[%d].role %q is duplicated", i, contract.Role))
		}
		seenEvidenceContracts[contract.Role] = struct{}{}
	}
	seen := map[string]struct{}{}
	for i, profile := range m.Profiles {
		if strings.TrimSpace(profile.Chain) == "" {
			errs = append(errs, fmt.Errorf("profiles[%d].chain is required", i))
		}
		if _, ok := seen[profile.Chain]; ok {
			errs = append(errs, fmt.Errorf("profiles[%d].chain %q is duplicated", i, profile.Chain))
		}
		seen[profile.Chain] = struct{}{}
		contract := profile.ProviderContract()
		if err := contract.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("profiles[%d].provider_contract: %w", i, err))
		}
		product := profile.NetworkProduct("public")
		if err := product.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("profiles[%d].network_product: %w", i, err))
		}
		if err := contract.SupportsProduct(product); err != nil {
			errs = append(errs, fmt.Errorf("profiles[%d].contract_support: %w", i, err))
		}
		if profile.Storage.Mode == "" || profile.Storage.MinDiskBytes <= 0 || profile.Storage.RecommendedDiskBytes < profile.Storage.MinDiskBytes {
			errs = append(errs, fmt.Errorf("profiles[%d].storage is invalid", i))
		}
		if profile.Storage.MinDiskDisplay == "" || profile.Storage.RecommendedDiskDisplay == "" || profile.Storage.GrowthMarginDisplay == "" {
			errs = append(errs, fmt.Errorf("profiles[%d].storage display guidance is required", i))
		}
		if profile.Network.ProfileID == "" || profile.Network.PeerPort <= 0 {
			errs = append(errs, fmt.Errorf("profiles[%d].network is invalid", i))
		}
		if profile.Rewards.ProtocolRewardDestination != CryptoRewardDestinationTreasury || profile.Rewards.DirectWorkerPayout || profile.Rewards.ProtocolRewardProofClaimed {
			errs = append(errs, fmt.Errorf("profiles[%d].rewards must route through treasury without protocol proof claim", i))
		}
		if profile.Rewards.TreasuryAccountID != profile.ProductID+"-treasury" || profile.Rewards.TreasuryWalletRef != profile.WalletRef || profile.Rewards.ManagementFeeBasisPoints < 0 {
			errs = append(errs, fmt.Errorf("profiles[%d].rewards treasury policy is invalid", i))
		}
		if profile.Proof.Mode != CryptoProofModeOperationalEvidence || !profile.Proof.ShapeOnly || profile.Proof.ProtocolNativeRewardProof {
			errs = append(errs, fmt.Errorf("profiles[%d].proof must be operational evidence only", i))
		}
		if !profile.Image.DigestPinnedRequired || profile.Image.UpstreamClientName == "" {
			errs = append(errs, fmt.Errorf("profiles[%d].image is invalid", i))
		}
	}
	return errors.Join(errs...)
}

func CryptoNetworkProfile(chain string) (CryptoProfile, bool) {
	switch strings.ToLower(strings.TrimSpace(chain)) {
	case "btc", "bitcoin":
		return CryptoProfile{
			Chain:             "btc",
			ProductID:         "btc-full-node",
			DisplayName:       "BTC Full Node",
			Purpose:           "Bitcoin full-node capacity with treasury settlement and participant attribution",
			PoolID:            "btc",
			ProviderID:        "btc-full-node",
			ContractID:        "crypto.btc-full-node.v1",
			SchemaRef:         "schema://providers/workflow-plugin-crypto/btc-full-node/v1",
			SchemaDigest:      cryptoSchemaDigest("btc"),
			ConfigRef:         "config://network-products/btc-full-node/btc-full-node",
			SettlementNetwork: "bitcoin",
			WalletRef:         "wallet://btc-full-node/primary",
			MinDiskBytes:      800000000000,
			MinMemoryBytes:    8000000000,
			MinBandwidthMbps:  50,
			Role:              cryptoFullNodeRole(),
			Storage:           cryptoFullNodeStorage(800000000000, 1000000000000, 200000000000),
			Network:           cryptoPublicChainNetwork(8333, 125, 50_000_000),
			Rewards:           cryptoTreasuryRewards("btc-full-node", "wallet://btc-full-node/primary"),
			Proof:             cryptoOperationalProof(),
			Image: CryptoImageMetadata{
				UpstreamClientName:       "bitcoind",
				DigestPinnedRequired:     true,
				OperatorSuppliedRequired: true,
				KnownImageRefs:           []string{"bitcoin/bitcoin@sha256:<operator-confirmed-digest>"},
			},
		}, true
	case "bch", "bitcoin-cash":
		return CryptoProfile{
			Chain:             "bch",
			ProductID:         "bch-full-node",
			DisplayName:       "BCH Full Node",
			Purpose:           "Bitcoin Cash full-node capacity with treasury settlement and participant attribution",
			PoolID:            "bch",
			ProviderID:        "bch-full-node",
			ContractID:        "crypto.bch-full-node.v1",
			SchemaRef:         "schema://providers/workflow-plugin-crypto/bch-full-node/v1",
			SchemaDigest:      cryptoSchemaDigest("bch"),
			ConfigRef:         "config://network-products/bch-full-node/bch-full-node",
			SettlementNetwork: "bitcoin-cash",
			WalletRef:         "wallet://bch-full-node/primary",
			MinDiskBytes:      800000000000,
			MinMemoryBytes:    8000000000,
			MinBandwidthMbps:  50,
			Role:              cryptoFullNodeRole(),
			Storage:           cryptoFullNodeStorage(800000000000, 1000000000000, 200000000000),
			Network:           cryptoPublicChainNetwork(8333, 125, 50_000_000),
			Rewards:           cryptoTreasuryRewards("bch-full-node", "wallet://bch-full-node/primary"),
			Proof:             cryptoOperationalProof(),
			Image: CryptoImageMetadata{
				UpstreamClientName:       "bitcoind",
				DigestPinnedRequired:     true,
				OperatorSuppliedRequired: true,
				KnownImageRefs:           []string{"zquestz/bitcoin-cash-node@sha256:<digest>"},
			},
		}, true
	case "ethereum", "eth":
		return CryptoProfile{
			Chain:             "ethereum",
			ProductID:         "ethereum-full-node",
			DisplayName:       "Ethereum Full Node",
			Purpose:           "Ethereum full-node capacity with treasury settlement and participant attribution",
			PoolID:            "ethereum",
			ProviderID:        "ethereum-full-node",
			ContractID:        "crypto.ethereum-full-node.v1",
			SchemaRef:         "schema://providers/workflow-plugin-crypto/ethereum-full-node/v1",
			SchemaDigest:      cryptoSchemaDigest("ethereum"),
			ConfigRef:         "config://network-products/ethereum-full-node/ethereum-full-node",
			SettlementNetwork: "ethereum",
			WalletRef:         "wallet://ethereum-full-node/primary",
			MinDiskBytes:      1200000000000,
			MinMemoryBytes:    16000000000,
			MinBandwidthMbps:  100,
			Role:              cryptoFullNodeRole(),
			Storage:           cryptoFullNodeStorage(1200000000000, 2000000000000, 500000000000),
			Network:           cryptoPublicChainNetwork(30303, 50, 100_000_000),
			Rewards:           cryptoTreasuryRewards("ethereum-full-node", "wallet://ethereum-full-node/primary"),
			Proof:             cryptoOperationalProof(),
			Image: CryptoImageMetadata{
				UpstreamClientName:   "geth",
				DigestPinnedRequired: true,
				RecommendedImageRef:  "ethereum/client-go@sha256:<digest>",
				KnownImageRefs:       []string{"ethereum/client-go@sha256:<digest>"},
			},
		}, true
	default:
		return CryptoProfile{}, false
	}
}

func cryptoFullNodeRole() CryptoRoleMetadata {
	return CryptoRoleMetadata{
		ID:                       "full-node",
		ShapeOnly:                true,
		ProtocolRewardsAssumed:   false,
		OperationalConformanceID: "shape-only",
	}
}

func cryptoFullNodeStorage(minDisk, recommendedDisk, growthMargin int64) CryptoStorageMetadata {
	return CryptoStorageMetadata{
		Mode:                         CryptoStorageArchiveFull,
		MinDiskBytes:                 minDisk,
		MinDiskDisplay:               cryptoDecimalByteDisplay(minDisk),
		RecommendedDiskBytes:         recommendedDisk,
		RecommendedDiskDisplay:       cryptoDecimalByteDisplay(recommendedDisk),
		GrowthMarginBytes:            growthMargin,
		GrowthMarginDisplay:          cryptoDecimalByteDisplay(growthMargin),
		DurableVolumeRequired:        true,
		PreserveOnUpdate:             true,
		PreserveOnUninstall:          true,
		PruningSupported:             false,
		SnapshotVerificationRequired: false,
	}
}

func cryptoDecimalByteDisplay(bytes int64) string {
	const (
		gb = int64(1_000_000_000)
		tb = int64(1_000_000_000_000)
	)
	if bytes >= tb {
		whole := bytes / tb
		remainder := bytes % tb
		if remainder == 0 {
			return fmt.Sprintf("%d TB", whole)
		}
		return fmt.Sprintf("%d.%d TB", whole, remainder/(tb/10))
	}
	if bytes >= gb {
		whole := bytes / gb
		remainder := bytes % gb
		if remainder == 0 {
			return fmt.Sprintf("%d GB", whole)
		}
		return fmt.Sprintf("%d.%d GB", whole, remainder/(gb/10))
	}
	return fmt.Sprintf("%d bytes", bytes)
}

func cryptoPublicChainNetwork(peerPort, maxOutboundPeers int, maxBytesPerSecond int64) CryptoNetworkMetadata {
	return CryptoNetworkMetadata{
		ProfileID:                 CryptoNetworkProfilePublicChainPeer,
		PeerPort:                  peerPort,
		AllowedPeerPorts:          []int{peerPort},
		RequiresIngress:           true,
		UsesDNSSeeds:              true,
		RPCPrivateOnly:            true,
		AuditRequired:             true,
		MaxOutboundPeers:          maxOutboundPeers,
		MaxOutboundBytesPerSecond: maxBytesPerSecond,
		KillClosesPeers:           true,
	}
}

func cryptoTreasuryRewards(productID, walletRef string) CryptoRewardMetadata {
	return CryptoRewardMetadata{
		ProtocolRewardDestination:  CryptoRewardDestinationTreasury,
		TreasuryAccountID:          productID + "-treasury",
		TreasuryWalletRef:          walletRef,
		ManagementFeeBasisPoints:   1,
		DirectWorkerPayout:         false,
		ProtocolRewardProofClaimed: false,
	}
}

func cryptoOperationalProof() CryptoProofMetadata {
	return CryptoProofMetadata{
		Mode:                      CryptoProofModeOperationalEvidence,
		ShapeOnly:                 true,
		ProtocolNativeRewardProof: false,
		EvidenceRefs: []string{
			"artifact://node-service/health-check",
			"artifact://node-service/version-probe",
			"artifact://node-service/resource-usage",
		},
	}
}

func cryptoSchemaDigest(chain string) string {
	schema := `{"$schema":"https://json-schema.org/draft/2020-12/schema","type":"object","additionalProperties":false,"properties":{"chain":{"const":"` + chain + `"},"network":{"type":"string"},"data_dir_ref":{"type":"string"},"rpc_secret_ref":{"type":"string"},"wallet_ref":{"type":"string"}}}`
	sum := sha256.Sum256([]byte(schema))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func (p CryptoProfile) ProviderContract() ProviderContract {
	runtime := DefaultProviderRuntimeProfile("node-service-sandboxed-container", ExecutionSandboxedContainer, ProofArtifactHash)
	runtime.UpstreamClientConformance = UpstreamClientConformanceShapeOnly
	return ProviderContract{
		ProtocolVersion:        Version,
		ID:                     p.ProviderID + "-v1",
		DisplayName:            p.DisplayName,
		PluginID:               cryptoPluginID,
		ProviderID:             p.ProviderID,
		ContractID:             p.ContractID,
		Version:                "v1.0.0",
		ConfigSchemaRef:        p.SchemaRef,
		ConfigSchemaDigest:     p.SchemaDigest,
		OperatingModes:         []NetworkOperatingMode{NetworkModeNodeService},
		WorkloadKinds:          []string{"node-service"},
		ExecutorProviders:      []string{"node-service-sandboxed-container"},
		ExecutionSecurityTiers: []ExecutionSecurityTier{ExecutionSandboxedContainer},
		ProofTiers:             []ProofTier{ProofArtifactHash},
		NetworkModes:           cryptoNodeNetworkModes(),
		RuntimeContract: ProviderRuntimeContract{Profiles: []ProviderRuntimeProfile{
			runtime,
		}},
	}
}

func CryptoUpstreamClientRequirement(chain string) (ProviderUpstreamClientRequirement, bool) {
	profile, ok := CryptoNetworkProfile(chain)
	if !ok {
		return ProviderUpstreamClientRequirement{}, false
	}
	runtime := DefaultProviderRuntimeProfile("node-service-sandboxed-container", ExecutionSandboxedContainer, ProofArtifactHash)
	req := ProviderUpstreamClientRequirement{
		ProtocolVersion:       Version,
		PluginID:              cryptoPluginID,
		ProviderID:            profile.ProviderID,
		ContractID:            profile.ContractID,
		Version:               "v1.0.0",
		RuntimeProfileID:      runtime.ID,
		ConformanceProfile:    "upstream-client-v1",
		DefaultConformance:    UpstreamClientConformanceShapeOnly,
		RealClientConformance: UpstreamClientConformanceRealClient,
		UpstreamClientName:    "bitcoind",
		VersionProbeCommand:   []string{"bitcoind", "--version"},
		ImagePolicy: ProviderUpstreamImagePolicy{
			DigestPinnedImageRequired:     true,
			OperatorSuppliedImageRequired: true,
		},
		RequiredEvidence: []string{
			"digest-pinned OCI image reference",
			"artifact:// provider conformance evidence with sha256 digest",
			"upstream-client-v1 version probe artifact",
		},
	}
	switch profile.Chain {
	case "ethereum":
		req.UpstreamClientName = "geth"
		req.VersionProbeCommand = []string{"geth", "version"}
		req.ImagePolicy.OperatorSuppliedImageRequired = false
		req.ImagePolicy.RecommendedImageRef = "ethereum/client-go@sha256:<digest>"
		req.ImagePolicy.KnownImageRefs = []string{"ethereum/client-go@sha256:<digest>"}
		req.Notes = []string{
			"ethereum/client-go is the Geth image family; operators must pin and prove the digest they deploy",
			"shape-only remains the default until real upstream-client-v1 evidence is attached to the provider contract",
		}
	case "bch":
		req.ImagePolicy.KnownImageRefs = []string{"zquestz/bitcoin-cash-node@sha256:<digest>"}
		req.Notes = []string{
			"Bitcoin Cash uses implementation-specific image families rather than a single wfcompute-owned canonical image",
			"operators must choose a BCH implementation image, pin the digest, and prove the upstream client version before real-client promotion",
		}
	case "btc":
		req.ImagePolicy.KnownImageRefs = []string{"bitcoin/bitcoin@sha256:<operator-confirmed-digest>"}
		req.Notes = []string{
			"Bitcoin Core has no single canonical official OCI runtime image for wfcompute to assume",
			"operators must supply a digest-pinned implementation image and prove the upstream client version before real-client promotion",
		}
	}
	return req, true
}

func (p CryptoProfile) NetworkProduct(orgID string) NetworkProduct {
	if strings.TrimSpace(orgID) == "" {
		orgID = "public"
	}
	return NetworkProduct{
		ProtocolVersion: Version,
		ID:              p.ProductID,
		DisplayName:     p.DisplayName,
		Purpose:         p.Purpose,
		OperatingMode:   NetworkModeNodeService,
		OrgID:           orgID,
		PoolID:          p.PoolID,
		WorkloadKinds:   []string{"node-service"},
		SecurityFloor: PlacementRequirements{
			ExecutorProvider:      "node-service-sandboxed-container",
			ExecutionSecurityTier: ExecutionSandboxedContainer,
			ProofTier:             ProofArtifactHash,
		},
		SessionPolicy: SessionPolicy{WarmSeconds: 3600, MinRenewals: 1, MaxBatchRequests: 1},
		ProviderConfig: ProviderConfig{
			PluginID:   cryptoPluginID,
			ProviderID: p.ProviderID,
			ContractID: p.ContractID,
			Version:    "v1.0.0",
			ConfigRef:  p.ConfigRef,
		},
		NetworkModes: cryptoNodeNetworkModes(),
		PlacementConstraints: PlacementConstraints{
			Chain:            p.Chain,
			Role:             "full-node",
			MinDiskBytes:     p.MinDiskBytes,
			MinMemoryBytes:   p.MinMemoryBytes,
			MinBandwidthMbps: p.MinBandwidthMbps,
			RequiresIngress:  true,
			WalletRef:        p.WalletRef,
			StorageGuidance: StorageGuidance{
				Mode:                         p.Storage.Mode,
				MinDiskBytes:                 p.Storage.MinDiskBytes,
				MinDiskDisplay:               p.Storage.MinDiskDisplay,
				RecommendedDiskBytes:         p.Storage.RecommendedDiskBytes,
				RecommendedDiskDisplay:       p.Storage.RecommendedDiskDisplay,
				GrowthMarginBytes:            p.Storage.GrowthMarginBytes,
				GrowthMarginDisplay:          p.Storage.GrowthMarginDisplay,
				DurableVolumeRequired:        p.Storage.DurableVolumeRequired,
				PreserveOnUpdate:             p.Storage.PreserveOnUpdate,
				PreserveOnUninstall:          p.Storage.PreserveOnUninstall,
				PruningSupported:             p.Storage.PruningSupported,
				SnapshotVerificationRequired: p.Storage.SnapshotVerificationRequired,
			},
		},
		RewardPolicy:        "profit-share",
		AbusePolicy:         "crypto-node-v1",
		SettlementAccountID: p.Rewards.TreasuryAccountID,
		SettlementTarget: SettlementTarget{
			Kind:      SettlementTargetTreasuryWallet,
			AccountID: p.Rewards.TreasuryAccountID,
			Network:   p.SettlementNetwork,
			WalletRef: p.Rewards.TreasuryWalletRef,
		},
		CryptoRewardRouting: CryptoRewardRoutingPolicy{
			Network:                 p.SettlementNetwork,
			TreasuryAccountID:       p.Rewards.TreasuryAccountID,
			TreasuryWalletRef:       p.Rewards.TreasuryWalletRef,
			CustodyMode:             CryptoRewardCustodyTreasuryThenDistribute,
			DistributionMode:        CryptoRewardDistributionContributionShare,
			ParticipantWalletSource: CryptoRewardParticipantAccountWallet,
			ManagementFeeBps:        1,
		},
	}
}

func cryptoNodeNetworkModes() []NetworkMode {
	return []NetworkMode{NetworkModeDirect, NetworkModeRelay}
}
