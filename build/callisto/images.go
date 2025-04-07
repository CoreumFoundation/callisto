package callisto

import (
	"context"
	"path/filepath"

	"github.com/CoreumFoundation/callisto/build/callisto/image"
	"github.com/CoreumFoundation/crust/build/config"
	"github.com/CoreumFoundation/crust/build/docker"
	"github.com/CoreumFoundation/crust/build/tools"
	"github.com/CoreumFoundation/crust/build/types"
)

// BuildDockerImage builds docker image.
func BuildDockerImage(ctx context.Context, deps types.DepsFunc) error {
	return buildDockerImage(ctx, deps, false)
}

// ReleaseDockerImage releases docker image.
func ReleaseDockerImage(ctx context.Context, deps types.DepsFunc) error {
	return buildDockerImage(ctx, deps, true)
}

func buildDockerImage(ctx context.Context, deps types.DepsFunc, push bool) error {
	deps(Build)

	dockerfile, err := image.Execute(image.Data{
		From:         docker.AlpineImage,
		BinaryPath:   filepath.Join("bin", ".cache", binaryName, tools.TargetPlatformLinuxLocalArchInDocker.String(), "bin", binaryName),
		BinaryName:   binaryName,
		DBSchemaPath: filepath.Join("database", "schema"),
	})
	if err != nil {
		return err
	}

	var action docker.Action
	if push {
		action = docker.ActionPush
	} else {
		action = docker.ActionLoad
	}

	return docker.BuildImage(ctx, docker.BuildImageConfig{
		ContextDir:      ".",
		ImageName:       config.DockerHubUsername + "/" + binaryName,
		TargetPlatforms: []tools.TargetPlatform{tools.TargetPlatformLinuxLocalArchInDocker},
		Dockerfile:      dockerfile,
		Action:          action,
		Versions: []string{
			"latest",
		},
	})
}
