// Command workflow-plugin-crypto runs as an external workflow plugin and
// exposes crypto network provider catalog metadata.
package main

import (
	"github.com/GoCodeAlone/workflow-plugin-crypto/internal"
	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

func main() {
	sdk.Serve(internal.NewPlugin(),
		sdk.WithBuildVersion(sdk.ResolveBuildVersion(internal.Version)),
	)
}
