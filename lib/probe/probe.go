package probe

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

func InspectContainer(containerEngine ContainerEngine) error {
	switch containerEngine {
	case ContainerEngineDocker:
		err := inspectContainerDocker()
		if err != nil {
			return err
		}
	case ContainerEnginePodman:
		err := inspectContainerPodman()
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unknown container engine %s", containerEngine)
	}

	return nil
}

func inspectContainerDocker() error {
	cmd := exec.Command("docker", "inspect", "globalping-probe", "-f", "{{.State.Status}}")
	containerStatus, err := cmd.Output()
	if err == nil {
		containerStatusStr := string(bytes.TrimSpace(containerStatus))
		return fmt.Errorf("The globalping-probe container is already installed on your system. Current status: %s", containerStatusStr)
	}

	return nil
}

func inspectContainerPodman() error {
	cmd := exec.Command("podman", "inspect", "globalping-probe", "-f", "{{.State.Status}}")
	containerStatus, err := cmd.Output()
	if err == nil {
		containerStatusStr := string(bytes.TrimSpace(containerStatus))
		if containerStatusStr == "" {
			// false positive as podmain keeps container info after deletion
			return nil
		}

		return fmt.Errorf("The globalping-probe container is already installed on your system. Current status: %s", containerStatusStr)
	}

	return nil
}

func RunContainer(containerEngine ContainerEngine) error {
	switch containerEngine {
	case ContainerEngineDocker:
		err := runContainerDocker()
		if err != nil {
			return err
		}
	case ContainerEnginePodman:
		err := runContainerPodman()
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unknown container engine %s", containerEngine)
	}

	return nil
}

func runContainerDocker() error {
	cmd := exec.Command("docker", "run", "-d", "--log-driver", "local", "--network", "host", "--restart", "always", "--name", "globalping-probe", "ghcr.io/jsdelivr/globalping-probe")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to run container: %v", err)
	}

	return nil
}

func runContainerPodman() error {
	cmd := exec.Command("podman", "run", "--cap-add=NET_RAW", "-d", "--network", "host", "--restart=always", "--name", "globalping-probe", "ghcr.io/jsdelivr/globalping-probe")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to run container: %v", err)
	}

	return nil
}
