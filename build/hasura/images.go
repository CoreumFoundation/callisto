package hasura

import (
	"context"
	_ "embed"

	"github.com/tokenize-x/tx-crust/build/config"
	"github.com/tokenize-x/tx-crust/build/docker"
	"github.com/tokenize-x/tx-crust/build/tools"
	"github.com/tokenize-x/tx-crust/build/types"
)

//go:embed Dockerfile
var dockerfile []byte

// BuildDockerImage builds docker image.
func BuildDockerImage(ctx context.Context, deps types.DepsFunc) error {
	return buildDockerImage(ctx, false)
}

// ReleaseDockerImage builds and releases docker image.
func ReleaseDockerImage(ctx context.Context, deps types.DepsFunc) error {
	return buildDockerImage(ctx, true)
}

func buildDockerImage(ctx context.Context, push bool) error {
	var action docker.Action
	if push {
		action = docker.ActionPush
	} else {
		action = docker.ActionLoad
	}
	return docker.BuildImage(ctx, docker.BuildImageConfig{
		ContextDir:      ".", // TODO (wojciech): Later on, move `hasura` dir here
		ImageName:       config.ContainerRegistryOrgName + "/hasura",
		TargetPlatforms: []tools.TargetPlatform{tools.TargetPlatformLinuxLocalArchInDocker},
		Dockerfile:      dockerfile,
		Action:          action,
		Versions: []string{
			"latest",
		},
	})
}
