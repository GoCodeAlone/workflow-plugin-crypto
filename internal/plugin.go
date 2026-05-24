// Package internal implements the workflow-plugin-crypto plugin.
package internal

import (
	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

// Version is set at build time via -ldflags
// "-X github.com/GoCodeAlone/workflow-plugin-crypto/internal.Version=X.Y.Z".
// Default is a bare semver so plugin loaders that validate semver accept
// unreleased dev builds; goreleaser overrides with the real release tag.
var Version = "0.0.0"

// CryptoPlugin exposes crypto network provider catalog metadata.
type CryptoPlugin struct{}

// NewPlugin returns a new plugin instance. main.go calls sdk.Serve(NewPlugin()).
func NewPlugin() sdk.PluginProvider {
	return &CryptoPlugin{}
}

// Manifest returns the plugin metadata used by the workflow engine for
// discovery and capability negotiation.
func (p *CryptoPlugin) Manifest() sdk.PluginManifest {
	return sdk.PluginManifest{
		Name:        "workflow-plugin-crypto",
		Version:     Version,
		Author:      "GoCodeAlone",
		Description: "Crypto network provider catalog plugin for workflow-compute.",
	}
}
