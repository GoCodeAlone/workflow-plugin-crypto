package internal_test

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/GoCodeAlone/workflow-plugin-crypto/catalog"
	"github.com/GoCodeAlone/workflow-plugin-crypto/internal"
	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

func TestNewPlugin_ImplementsPluginProvider(t *testing.T) {
	var _ sdk.PluginProvider = internal.NewPlugin()
}

func TestManifest_HasRequiredFields(t *testing.T) {
	m := internal.NewPlugin().Manifest()
	if m.Name == "" {
		t.Error("manifest Name is empty")
	}
	if m.Version == "" {
		t.Error("manifest Version is empty — build-time ldflags injection missing")
	}
	if m.Description == "" {
		t.Error("manifest Description is empty")
	}
	if strings.Contains(m.Description, "TEMPLATE") || strings.Contains(strings.ToLower(m.Description), "scaffold") {
		t.Fatalf("manifest still carries scaffold placeholder text: %q", m.Description)
	}
}

func TestPluginJSON_AdvertisesProviderCatalogOnly(t *testing.T) {
	data, err := os.ReadFile("../plugin.json")
	if err != nil {
		t.Fatalf("read plugin.json: %v", err)
	}
	var manifest struct {
		Name         string   `json:"name"`
		Description  string   `json:"description"`
		Type         string   `json:"type"`
		Private      bool     `json:"private"`
		Keywords     []string `json:"keywords"`
		Capabilities struct {
			ConfigProvider bool     `json:"configProvider"`
			ModuleTypes    []string `json:"moduleTypes"`
			StepTypes      []string `json:"stepTypes"`
			TriggerTypes   []string `json:"triggerTypes"`
		} `json:"capabilities"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("parse plugin.json: %v", err)
	}
	if manifest.Name != "workflow-plugin-crypto" || manifest.Type != "external" || manifest.Private {
		t.Fatalf("unexpected plugin identity: %+v", manifest)
	}
	joined := strings.Join(append(manifest.Keywords, manifest.Description), " ")
	if strings.Contains(joined, "TEMPLATE") || strings.Contains(strings.ToLower(joined), "scaffold") {
		t.Fatalf("plugin.json still carries scaffold placeholder text: %s", joined)
	}
	if manifest.Capabilities.ConfigProvider ||
		len(manifest.Capabilities.ModuleTypes) != 0 ||
		len(manifest.Capabilities.StepTypes) != 0 ||
		len(manifest.Capabilities.TriggerTypes) != 0 {
		t.Fatalf("crypto provider catalog should not advertise runtime capabilities: %+v", manifest.Capabilities)
	}
	if !slices.Contains(manifest.Keywords, "provider-catalog") || !slices.Contains(manifest.Keywords, "workflow-compute") {
		t.Fatalf("provider-catalog keywords missing: %+v", manifest.Keywords)
	}
}

func TestCryptoProviderManifest_ValidatesStableCatalog(t *testing.T) {
	manifest := catalog.CryptoProviderManifest()
	if err := manifest.Validate(); err != nil {
		t.Fatalf("crypto provider manifest invalid: %v", err)
	}
	if manifest.ProtocolVersion != catalog.Version || manifest.PluginID != "workflow-plugin-crypto" || manifest.Version != "v1.0.0" {
		t.Fatalf("manifest identity: %+v", manifest)
	}
	if len(manifest.Profiles) != 3 || manifest.Profiles[0].Chain != "btc" || manifest.Profiles[1].Chain != "bch" || manifest.Profiles[2].Chain != "ethereum" {
		t.Fatalf("manifest profiles are not stable btc/bch/ethereum order: %+v", manifest.Profiles)
	}
	if digest := catalog.CryptoProviderManifestDigest(); digest != "sha256:73faf36582844b9399686517b9cfd6ffe695fa47f14fedf87f4c02e74ec20b63" {
		t.Fatalf("crypto provider manifest digest drifted: got %s", digest)
	}
}

func TestCryptoNetworkCatalog_BuildsStrictContractsAndProducts(t *testing.T) {
	for _, tc := range []struct {
		chain      string
		productID  string
		providerID string
		contractID string
		peerPort   int
	}{
		{chain: "btc", productID: "btc-full-node", providerID: "btc-full-node", contractID: "crypto.btc-full-node.v1", peerPort: 8333},
		{chain: "bch", productID: "bch-full-node", providerID: "bch-full-node", contractID: "crypto.bch-full-node.v1", peerPort: 8333},
		{chain: "ethereum", productID: "ethereum-full-node", providerID: "ethereum-full-node", contractID: "crypto.ethereum-full-node.v1", peerPort: 30303},
	} {
		t.Run(tc.chain, func(t *testing.T) {
			profile, ok := catalog.CryptoNetworkProfile(tc.chain)
			if !ok {
				t.Fatalf("missing profile %q", tc.chain)
			}
			contract := profile.ProviderContract()
			if err := contract.Validate(); err != nil {
				t.Fatalf("provider contract invalid: %v", err)
			}
			if contract.PluginID != "workflow-plugin-crypto" || contract.ProviderID != tc.providerID || contract.ContractID != tc.contractID {
				t.Fatalf("contract identity: %+v", contract)
			}
			product := profile.NetworkProduct("public")
			if err := product.Validate(); err != nil {
				t.Fatalf("network product invalid: %v", err)
			}
			if product.ID != tc.productID || product.ProviderConfig.ProviderID != tc.providerID || product.ProviderConfig.ContractID != tc.contractID {
				t.Fatalf("product identity/provider: %+v", product)
			}
			if product.PlacementConstraints.Chain != tc.chain || product.PlacementConstraints.Role != catalog.CryptoRoleFullNode || !product.PlacementConstraints.RequiresIngress {
				t.Fatalf("placement constraints: %+v", product.PlacementConstraints)
			}
			if profile.Network.PeerPort != tc.peerPort || !profile.Network.RPCPrivateOnly || !profile.Network.AuditRequired {
				t.Fatalf("public-chain network metadata: %+v", profile.Network)
			}
			if err := contract.SupportsProduct(product); err != nil {
				t.Fatalf("contract should support product: %v", err)
			}
		})
	}
}

func TestCryptoUpstreamRequirements_ValidateImagePolicy(t *testing.T) {
	for _, chain := range []string{"btc", "bch", "ethereum"} {
		t.Run(chain, func(t *testing.T) {
			req, ok := catalog.CryptoUpstreamClientRequirement(chain)
			if !ok {
				t.Fatalf("missing upstream requirement %q", chain)
			}
			if err := req.Validate(); err != nil {
				t.Fatalf("upstream requirement invalid: %v", err)
			}
			if req.PluginID != "workflow-plugin-crypto" || req.DefaultConformance != catalog.UpstreamClientConformanceShapeOnly {
				t.Fatalf("upstream requirement identity/conformance: %+v", req)
			}
			if !req.ImagePolicy.DigestPinnedImageRequired {
				t.Fatalf("upstream images must require digest pins: %+v", req.ImagePolicy)
			}
		})
	}
}
