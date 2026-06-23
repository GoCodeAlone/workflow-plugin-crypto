package catalog

import core "github.com/GoCodeAlone/workflow-plugin-compute-core/protocol"

const (
	Version = core.Version

	NetworkModeBatch       = core.NetworkModeBatch
	NetworkModeNodeService = core.NetworkModeNodeService

	WorkloadProvider = core.WorkloadProvider

	ExecutionSandboxedContainer = core.ExecutionSandboxedContainer
	ProofArtifactHash           = core.ProofArtifactHash

	NetworkModeDirect  = core.NetworkModeDirect
	NetworkModeRelay   = core.NetworkModeRelay
	NetworkModeOffline = core.NetworkModeOffline

	UpstreamClientConformanceShapeOnly  = core.UpstreamClientConformanceShapeOnly
	UpstreamClientConformanceRealClient = core.UpstreamClientConformanceRealClient

	SettlementTargetTreasuryWallet = core.SettlementTargetTreasuryWallet

	CryptoRewardCustodyTreasuryThenDistribute = core.CryptoRewardCustodyTreasuryThenDistribute
	CryptoRewardDistributionContributionShare = core.CryptoRewardDistributionContributionShare
	CryptoRewardParticipantAccountWallet      = core.CryptoRewardParticipantAccountWallet
)

type (
	NetworkOperatingMode = core.NetworkOperatingMode
	NetworkMode          = core.NetworkMode

	ExecutionSecurityTier = core.ExecutionSecurityTier
	ProofTier             = core.ProofTier

	ProviderContract        = core.ProviderContract
	ProviderOperation       = core.ProviderOperation
	ProviderArtifactSpec    = core.ProviderArtifactSpec
	ProviderRuntimeContract = core.ProviderRuntimeContract
	ProviderRuntimeProfile  = core.ProviderRuntimeProfile

	NetworkProduct        = core.NetworkProduct
	PlacementRequirements = core.PlacementRequirements
	SessionPolicy         = core.SessionPolicy
	ProviderConfig        = core.ProviderConfig
	PlacementConstraints  = core.PlacementConstraints
	StorageGuidance       = core.StorageGuidance

	SettlementTarget          = core.SettlementTarget
	CryptoRewardRoutingPolicy = core.CryptoRewardRoutingPolicy

	ProviderUpstreamClientRequirement = core.ProviderUpstreamClientRequirement
	ProviderUpstreamImagePolicy       = core.ProviderUpstreamImagePolicy
)

var (
	CanonicalHash                 = core.CanonicalHash
	DefaultProviderRuntimeProfile = core.DefaultProviderRuntimeProfile
)
