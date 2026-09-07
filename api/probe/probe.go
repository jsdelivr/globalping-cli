package probe

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

var ErrContainerAlreadyInstalled = errors.New("the globalping-probe container is already installed on your system")

const (
	containerName       = "globalping-probe"
	containerListFormat = "{{.Names}}\t{{.State}}\t{{.Status}}"
)

type Probe interface {
	DetectContainerEngine() (ContainerEngine, error)
	InspectContainer(containerEngine ContainerEngine) error
	RunContainer(containerEngine ContainerEngine) error
}

type executionResult struct {
	stdout []byte
	stderr []byte
	err    error
}

type executor func(name string, args ...string) executionResult

type probe struct {
	execute executor
	stdout  io.Writer
	stderr  io.Writer
}

func NewProbe() Probe {
	return &probe{
		execute: executeCommand,
		stdout:  os.Stdout,
		stderr:  os.Stderr,
	}
}

func (p *probe) InspectContainer(containerEngine ContainerEngine) error {
	var result executionResult

	switch containerEngine {
	case ContainerEngineDocker:
		result = p.execute("docker", "ps", "--all", "--format", containerListFormat)
	case ContainerEnginePodman:
		result = p.execute("sudo", "podman", "ps", "--all", "--format", containerListFormat)
	default:
		return fmt.Errorf("unknown container engine %s", containerEngine)
	}

	if result.err != nil {
		return fmt.Errorf("failed to query %s containers: %w", containerEngine, executionError(result))
	}

	err := classifyContainerList(string(result.stdout))

	if err != nil && !errors.Is(err, ErrContainerAlreadyInstalled) {
		return fmt.Errorf("failed to classify %s container list: %w", containerEngine, err)
	}

	return err
}

func (p *probe) RunContainer(containerEngine ContainerEngine) error {
	var result executionResult

	switch containerEngine {
	case ContainerEngineDocker:
		result = p.execute("docker", "run", "-d", "--log-driver", "local", "--network", "host", "--restart", "always", "--name", containerName, "globalping/globalping-probe")
	case ContainerEnginePodman:
		result = p.execute("sudo", "podman", "run", "--cap-add=NET_RAW", "-d", "--network", "host", "--restart=always", "--name", containerName, "globalping/globalping-probe")
	default:
		return fmt.Errorf("unknown container engine %s", containerEngine)
	}

	if result.err != nil {
		return fmt.Errorf("failed to run %s container: %w", containerEngine, executionError(result))
	}

	if p.stdout != nil {
		_, _ = p.stdout.Write(result.stdout)
	}

	if p.stderr != nil {
		_, _ = p.stderr.Write(result.stderr)
	}

	return nil
}

func executeCommand(name string, args ...string) executionResult {
	cmd := exec.Command(name, args...)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()

	return executionResult{
		stdout: stdout.Bytes(),
		stderr: stderr.Bytes(),
		err:    err,
	}
}

func executionError(result executionResult) error {
	details := bytes.TrimSpace(result.stderr)

	if len(details) == 0 {
		details = bytes.TrimSpace(result.stdout)
	}

	if len(details) == 0 {
		return result.err
	}

	detailText := strings.ReplaceAll(string(details), "\r\n", "\n")
	detailText = strings.ReplaceAll(detailText, "\r", "\n")
	detailText = strings.ReplaceAll(detailText, "\n", "; ")

	return fmt.Errorf("%w: %s", result.err, detailText)
}

func classifyContainerList(output string) error {
	output = strings.TrimSpace(output)

	if output == "" {
		return nil
	}

	installedRows := make([][]string, 0, 1)

	for line := range strings.SplitSeq(output, "\n") {
		parts := strings.SplitN(strings.TrimSuffix(line, "\r"), "\t", 3)

		if len(parts) != 3 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" || strings.TrimSpace(parts[2]) == "" {
			return fmt.Errorf("malformed container list output: %q", line)
		}

		if strings.TrimSpace(parts[0]) == containerName {
			installedRows = append(installedRows, parts)
		}
	}

	if len(installedRows) > 1 {
		return fmt.Errorf("ambiguous container list output: found %d %s rows", len(installedRows), containerName)
	}

	if len(installedRows) == 1 {
		return fmt.Errorf(
			"%w. Current state: %s; status: %s",
			ErrContainerAlreadyInstalled,
			strings.TrimSpace(installedRows[0][1]),
			strings.TrimSpace(installedRows[0][2]),
		)
	}

	return nil
}
