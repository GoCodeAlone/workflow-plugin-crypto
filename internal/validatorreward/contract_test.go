package validatorreward

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/GoCodeAlone/workflow-plugin-crypto/catalog"
)

func TestProviderContractArtifactAlignsWithCatalogAndSchemas(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "contracts", "ethereum-validator-reward-provider.json"))
	if err != nil {
		t.Fatalf("read contract: %v", err)
	}
	var contract catalog.ProviderContract
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&contract); err != nil {
		t.Fatalf("decode contract: %v", err)
	}
	if err := contract.Validate(); err != nil {
		t.Fatalf("contract should validate: %v", err)
	}
	profile, ok := catalog.CryptoEthereumTestnetValidatorRewardProfile()
	if !ok {
		t.Fatal("missing catalog profile")
	}
	catalogContract := profile.ProviderContract()
	if contract.PluginID != catalogContract.PluginID ||
		contract.ProviderID != catalogContract.ProviderID ||
		contract.ContractID != catalogContract.ContractID ||
		contract.Operations[0].ID != Operation ||
		contract.Operations[0].Artifacts[0] != EvidenceArtifact {
		t.Fatalf("static contract drifted from catalog: %+v", contract)
	}
	if got, want := contract.ConfigSchemaDigest, schemaDigest(t, "ethereum-validator-reward-provider.schema.json"); got != want {
		t.Fatalf("config schema digest = %q, want %q", got, want)
	}
	if got, want := contract.Operations[0].InputSchemaDigest, schemaDigest(t, "ethereum-validator-reward-operation-input.schema.json"); got != want {
		t.Fatalf("input schema digest = %q, want %q", got, want)
	}
	if got, want := contract.Operations[0].OutputSchemaDigest, schemaDigest(t, "ethereum-validator-reward-operation-output.schema.json"); got != want {
		t.Fatalf("output schema digest = %q, want %q", got, want)
	}
}

func TestPluginManifestsExposeValidatorRewardContract(t *testing.T) {
	for _, path := range []string{"plugin.json", "plugin.contracts.json"} {
		data, err := os.ReadFile(filepath.Join("..", "..", path))
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		contracts, err := decodeContractRefs(path, data)
		if err != nil {
			t.Fatalf("decode %s: %v", path, err)
		}
		found := false
		for _, contract := range contracts {
			if contract.ID == "crypto.ethereum-testnet-validator-reward.v1" &&
				contract.Path == "contracts/ethereum-validator-reward-provider.json" &&
				contract.Schema == "schemas/ethereum-validator-reward-provider.schema.json" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("%s does not expose ethereum validator reward provider contract: %+v", path, contracts)
		}
	}
}

type contractRef struct {
	ID     string `json:"id"`
	Path   string `json:"path"`
	Schema string `json:"schema"`
}

func decodeContractRefs(path string, data []byte) ([]contractRef, error) {
	if path == "plugin.contracts.json" {
		var manifest struct {
			Version   string        `json:"version"`
			Contracts []contractRef `json:"contracts"`
		}
		dec := json.NewDecoder(bytes.NewReader(data))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&manifest); err != nil {
			return nil, err
		}
		return manifest.Contracts, nil
	}
	var manifest struct {
		Contracts []contractRef `json:"contracts"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}
	return manifest.Contracts, nil
}

func schemaDigest(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "schemas", name))
	if err != nil {
		t.Fatalf("read schema %s: %v", name, err)
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
