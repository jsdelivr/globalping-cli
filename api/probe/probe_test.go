package probe

import (
	"bytes"
	"errors"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type expectedExecution struct {
	name   string
	args   []string
	result executionResult
}

func fakeExecutor(t *testing.T, expected []expectedExecution) executor {
	t.Helper()

	next := 0
	t.Cleanup(func() {
		assert.Equal(t, len(expected), next, "not all expected commands were executed")
	})

	return func(name string, args ...string) executionResult {
		t.Helper()
		require.Less(t, next, len(expected), "unexpected command: %s %v", name, args)
		call := expected[next]
		next++
		assert.Equal(t, call.name, name)
		assert.Equal(t, call.args, args)

		return call.result
	}
}

func TestDetectContainerEngine(t *testing.T) {
	t.Run("docker", func(t *testing.T) {
		p := &probe{execute: fakeExecutor(t, []expectedExecution{
			{name: "docker", args: []string{"info"}},
			{name: "type", args: []string{"docker"}, result: executionResult{stdout: []byte("docker is /usr/local/bin/docker\n")}},
		})}

		engine, err := p.DetectContainerEngine()

		require.NoError(t, err)
		assert.Equal(t, ContainerEngineDocker, engine)
	})

	t.Run("docker alias for podman", func(t *testing.T) {
		p := &probe{execute: fakeExecutor(t, []expectedExecution{
			{name: "docker", args: []string{"info"}},
			{name: "type", args: []string{"docker"}, result: executionResult{stdout: []byte("docker is aliased to podman")}},
		})}

		engine, err := p.DetectContainerEngine()

		require.NoError(t, err)
		assert.Equal(t, ContainerEnginePodman, engine)
	})

	t.Run("podman fallback", func(t *testing.T) {
		dockerErr := errors.New("docker unavailable")
		p := &probe{execute: fakeExecutor(t, []expectedExecution{
			{name: "docker", args: []string{"info"}, result: executionResult{err: dockerErr}},
			{name: "podman", args: []string{"info"}},
		})}

		engine, err := p.DetectContainerEngine()

		require.NoError(t, err)
		assert.Equal(t, ContainerEnginePodman, engine)
	})

	t.Run("both fail with diagnostics", func(t *testing.T) {
		dockerErr := errors.New("docker exit")
		podmanErr := &exec.Error{Name: "podman", Err: exec.ErrNotFound}
		p := &probe{execute: fakeExecutor(t, []expectedExecution{
			{name: "docker", args: []string{"info"}, result: executionResult{stderr: []byte("cannot connect to docker\nerrors pretty printing info\n"), err: dockerErr}},
			{name: "podman", args: []string{"info"}, result: executionResult{err: podmanErr}},
		})}

		engine, err := p.DetectContainerEngine()

		assert.Equal(t, ContainerEngineUnknown, engine)
		assert.ErrorIs(t, err, dockerErr)
		assert.ErrorIs(t, err, podmanErr)
		assert.EqualError(t, err, "  Docker: installed but unavailable: docker exit: cannot connect to docker; errors pretty printing info\n"+
			"  Podman: not installed or not available in PATH: exec: \"podman\": executable file not found in $PATH")
	})
}

func TestProbeInspectContainer(t *testing.T) {
	queryErr := errors.New("exit status 1")
	tests := []struct {
		name        string
		engine      ContainerEngine
		command     string
		args        []string
		result      executionResult
		errorIs     error
		errorText   string
		wantNoError bool
	}{
		{
			name:        "docker absent",
			engine:      ContainerEngineDocker,
			command:     "docker",
			args:        []string{"ps", "--all", "--format", containerListFormat},
			result:      executionResult{stdout: []byte("another-container\trunning\tUp 3 hours\n")},
			wantNoError: true,
		},
		{
			name:      "podman already installed",
			engine:    ContainerEnginePodman,
			command:   "sudo",
			args:      []string{"podman", "ps", "--all", "--format", containerListFormat},
			result:    executionResult{stdout: []byte("globalping-probe\texited\tExited (1) 2 minutes ago\n")},
			errorIs:   ErrContainerAlreadyInstalled,
			errorText: "Current state: exited; status: Exited (1) 2 minutes ago",
		},
		{
			name:      "query failure",
			engine:    ContainerEngineDocker,
			command:   "docker",
			args:      []string{"ps", "--all", "--format", containerListFormat},
			result:    executionResult{stderr: []byte("permission denied\n"), err: queryErr},
			errorIs:   queryErr,
			errorText: "permission denied",
		},
		{
			name:      "malformed output",
			engine:    ContainerEngineDocker,
			command:   "docker",
			args:      []string{"ps", "--all", "--format", containerListFormat},
			result:    executionResult{stdout: []byte("globalping-probe running Up 3 hours\n")},
			errorText: "malformed container list output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &probe{execute: fakeExecutor(t, []expectedExecution{{name: tt.command, args: tt.args, result: tt.result}})}

			err := p.InspectContainer(tt.engine)

			if tt.wantNoError {
				require.NoError(t, err)

				return
			}

			require.Error(t, err)

			if tt.errorIs != nil {
				assert.ErrorIs(t, err, tt.errorIs)
			}

			assert.ErrorContains(t, err, tt.errorText)
		})
	}
}

func TestClassifyContainerListRequiresExactUnambiguousRows(t *testing.T) {
	assert.NoError(t, classifyContainerList("globalping-probe-helper\trunning\tUp 2 hours\n"))
	assert.ErrorContains(t, classifyContainerList("globalping-probe\trunning\tUp 2 hours\nglobalping-probe\texited\tExited (0)\n"), "ambiguous")
	assert.ErrorContains(t, classifyContainerList("globalping-probe\t\tUp 2 hours\n"), "malformed")
}

func TestProbeRunContainerOutput(t *testing.T) {
	t.Run("success forwards captured streams", func(t *testing.T) {
		stdout := new(bytes.Buffer)
		stderr := new(bytes.Buffer)
		p := &probe{
			execute: fakeExecutor(t, []expectedExecution{{
				name:   "docker",
				args:   []string{"run", "-d", "--log-driver", "local", "--network", "host", "--restart", "always", "--name", containerName, "globalping/globalping-probe"},
				result: executionResult{stdout: []byte("container-id\n"), stderr: []byte("warning\n")},
			}}),
			stdout: stdout,
			stderr: stderr,
		}

		require.NoError(t, p.RunContainer(ContainerEngineDocker))
		assert.Equal(t, "container-id\n", stdout.String())
		assert.Equal(t, "warning\n", stderr.String())
	})

	t.Run("failure returns stderr without forwarding streams", func(t *testing.T) {
		runErr := errors.New("exit status 125")
		stdout := new(bytes.Buffer)
		stderr := new(bytes.Buffer)
		p := &probe{
			execute: fakeExecutor(t, []expectedExecution{{
				name:   "sudo",
				args:   []string{"podman", "run", "--cap-add=NET_RAW", "-d", "--network", "host", "--restart=always", "--name", containerName, "globalping/globalping-probe"},
				result: executionResult{stdout: []byte("partial-id\n"), stderr: []byte("image pull failed\n"), err: runErr},
			}}),
			stdout: stdout,
			stderr: stderr,
		}

		err := p.RunContainer(ContainerEnginePodman)

		assert.ErrorIs(t, err, runErr)
		assert.ErrorContains(t, err, "image pull failed")
		assert.Empty(t, stdout.String())
		assert.Empty(t, stderr.String())
	})

	t.Run("failure falls back to stdout diagnostics", func(t *testing.T) {
		runErr := errors.New("exit status 125")
		p := &probe{execute: fakeExecutor(t, []expectedExecution{{
			name:   "docker",
			args:   []string{"run", "-d", "--log-driver", "local", "--network", "host", "--restart", "always", "--name", containerName, "globalping/globalping-probe"},
			result: executionResult{stdout: []byte("daemon unavailable\n"), err: runErr},
		}})}

		err := p.RunContainer(ContainerEngineDocker)

		assert.ErrorIs(t, err, runErr)
		assert.ErrorContains(t, err, "daemon unavailable")
	})
}

func TestProbeRejectsUnknownContainerEngine(t *testing.T) {
	p := &probe{execute: func(string, ...string) executionResult {
		panic("executor must not be called")
	}}

	assert.EqualError(t, p.InspectContainer(ContainerEngineUnknown), "unknown container engine Unknown")
	assert.EqualError(t, p.RunContainer(ContainerEngineUnknown), "unknown container engine Unknown")
}
