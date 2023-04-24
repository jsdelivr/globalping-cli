package probe

import (
	"os"
	"os/exec"
	"strings"
)

type ContainerEngine string

const (
	ContainerEngineUnknown ContainerEngine = "Unknown"
	ContainerEngineDocker  ContainerEngine = "Docker"
	ContainerEnginePodman  ContainerEngine = "Podman"
)

func DetectContainerEngine() (ContainerEngine, error) {
	// check if docker is installed
	dockerInfoCmd := exec.Command("docker", "info")
	dockerInfoCmd.Stderr = os.Stderr
	dockerInfoErr := dockerInfoCmd.Run()
	if dockerInfoErr == nil {
		// check if docker is aliased to podman
		aliasCmd := exec.Command("type", "docker")
		aliasResults, _ := aliasCmd.Output()
		if strings.Contains(string(aliasResults), "podman") {
			return ContainerEnginePodman, nil
		}

		return ContainerEngineDocker, nil
	}

	// check if podman is installed
	podmanInfoCmd := exec.Command("podman", "info")
	podmanInfoCmd.Stderr = os.Stderr
	podmanInfoErr := podmanInfoCmd.Run()
	if podmanInfoErr == nil {
		return ContainerEnginePodman, nil
	}

	return ContainerEngineUnknown, dockerInfoErr
}
