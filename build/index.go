package build

import (
	"context"

	"github.com/tokenize-x/tx-callisto/build/callisto"
	"github.com/tokenize-x/tx-callisto/build/hasura"
	"github.com/tokenize-x/tx-crust/build/golang"
	crust "github.com/tokenize-x/tx-crust/build/tx-crust"
	"github.com/tokenize-x/tx-crust/build/types"
)

// Commands is a definition of commands available in build system.
var Commands = map[string]types.Command{
	"build/me":    {Fn: crust.BuildBuilder, Description: "Builds the builder"},
	"build/znet":  {Fn: crust.BuildZNet, Description: "Builds znet binary"},
	"build":       {Fn: callisto.Build, Description: "Builds callisto binary"},
	"build/amd64": {Fn: callisto.BuildAMD64, Description: "Builds callisto binary for arm64 platform"},
	"build/arm64": {Fn: callisto.BuildARM64, Description: "Builds callisto binary for amd64 platform"},
	"images": {Fn: func(ctx context.Context, deps types.DepsFunc) error {
		deps(callisto.BuildDockerImage, hasura.BuildDockerImage)
		return nil
	}, Description: "Builds callisto and hasura docker images"},
	"release/images": {Fn: func(ctx context.Context, deps types.DepsFunc) error {
		deps(callisto.ReleaseDockerImage, hasura.ReleaseDockerImage)
		return nil
	}, Description: "Builds callisto and hasura docker images and releases them"},
	"test": {Fn: golang.Test, Description: "Runs unit tests"},
	"tidy": {Fn: golang.Tidy, Description: "Runs go mod tidy"},
}
