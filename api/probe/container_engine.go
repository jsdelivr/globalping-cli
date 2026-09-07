package probe

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type ContainerEngine string

const (
	ContainerEngineUnknown ContainerEngine = "Unknown"
	ContainerEngineDocker  ContainerEngine = "Docker"
	ContainerEnginePodman  ContainerEngine = "Podman"
)

func (p *probe) DetectContainerEngine() (ContainerEngine, error) {
	// check if docker is installed
	dockerInfo := p.execute("docker", "info")

	if dockerInfo.err == nil {
		// check if docker is aliased to podman
		aliasResult := p.execute("type", "docker")

		if aliasResult.err == nil && strings.Contains(strings.ToLower(string(aliasResult.stdout)), "podman") {
			return ContainerEnginePodman, nil
		}

		return ContainerEngineDocker, nil
	}

	// check if podman is installed
	podmanInfo := p.execute("podman", "info")

	if podmanInfo.err == nil {
		return ContainerEnginePodman, nil
	}

	return ContainerEngineUnknown, errors.Join(
		containerEngineError("Docker", dockerInfo),
		containerEngineError("Podman", podmanInfo),
	)
}

func containerEngineError(name string, result executionResult) error {
	err := executionError(result)
	var execErr *exec.Error

	if errors.As(result.err, &execErr) {
		return fmt.Errorf("  %s: not installed or not available in PATH: %w", name, err)
	}

	return fmt.Errorf("  %s: installed but unavailable: %w", name, err)
}
