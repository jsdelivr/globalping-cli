package probe

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

type Probe interface {
	DetectContainerEngine() (ContainerEngine, error)
	InspectContainer(containerEngine ContainerEngine) error
	RunContainer(containerEngine ContainerEngine) error
}

type probe struct{}

func NewProbe() Probe {
	return &probe{}
}

func (p *probe) InspectContainer(containerEngine ContainerEngine) error {
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
		return fmt.Errorf("unknown container engine %s", containerEngine)
	}

	return nil
}

func (p *probe) RunContainer(containerEngine ContainerEngine) error {
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
		return fmt.Errorf("unknown container engine %s", containerEngine)
	}

	return nil
}

func inspectContainerDocker() error {
	cmd := exec.Command("docker", "inspect", "globalping-probe", "-f", "{{.State.Status}}")
	containerStatus, err := cmd.Output()
	if err == nil {
		containerStatusStr := string(bytes.TrimSpace(containerStatus))
		return fmt.Errorf("the globalping-probe container is already installed on your system. Current status: %s", containerStatusStr)
	}

	return nil
}

func inspectContainerPodman() error {
	cmd := exec.Command("sudo", "podman", "inspect", "globalping-probe", "-f", "{{.State.Status}}")
	containerStatus, err := cmd.Output()
	if err == nil {
		containerStatusStr := string(bytes.TrimSpace(containerStatus))
		if containerStatusStr == "" {
			// false positive as podmain keeps container info after deletion
			return nil
		}

		return fmt.Errorf("the globalping-probe container is already installed on your system. Current status: %s", containerStatusStr)
	}

	return nil
}

func runContainerDocker() error {
	cmd := exec.Command("docker", "run", "-d", "--log-driver", "local", "--network", "host", "--restart", "always", "--name", "globalping-probe", "globalping/globalping-probe")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run container: %v", err)
	}

	return nil
}

func runContainerPodman() error {
	cmd := exec.Command("sudo", "podman", "run", "--cap-add=NET_RAW", "-d", "--network", "host", "--restart=always", "--name", "globalping-probe", "globalping/globalping-probe")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run container: %v", err)
	}

	return nil
}
