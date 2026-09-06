package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/jsdelivr/globalping-cli/api/probe"
	apiMocks "github.com/jsdelivr/globalping-cli/mocks/api"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_Execute_Install_Probe_Docker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	probeMock := apiMocks.NewMockProbe(ctrl)
	probeMock.EXPECT().DetectContainerEngine().Times(1).Return(probe.ContainerEngineDocker, nil)
	probeMock.EXPECT().InspectContainer(probe.ContainerEngineDocker).Times(1).Return(nil)
	probeMock.EXPECT().RunContainer(probe.ContainerEngineDocker).Times(1).Return(nil)

	reader := bytes.NewReader([]byte("Y\n"))
	w := new(bytes.Buffer)
	printer := view.NewPrinter(reader, w, w)
	ctx := createDefaultContext()
	root := NewRoot(printer, ctx, nil, nil, nil, probeMock, nil)
	os.Args = []string{"globalping", "install-probe"}

	err := root.Cmd.ExecuteContext(t.Context())
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, `Detected container engine: Docker

The Globalping platform is a community powered project and relies on individuals like yourself to host our probes and make them accessible to everyone else.
Please confirm to pull and run our Docker container (globalping/globalping-probe) [Y/n] The Globalping probe started successfully. Thank you for joining our community! 
`, w.String())

	expectedCtx := &view.Context{
		History:             view.NewHistoryBuffer(1),
		From:                "world",
		Limit:               1,
		RunSessionStartedAt: defaultCurrentTime,
	}
	assert.Equal(t, expectedCtx, ctx)
}

func Test_Execute_Install_Probe_DetectionFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	probeMock := apiMocks.NewMockProbe(ctrl)
	podmanErr := &exec.Error{Name: "podman", Err: exec.ErrNotFound}
	detectErr := errors.New("  Docker: installed but unavailable: exit status 1: request returned 500; errors pretty printing info\n" +
		"  Podman: not installed or not available in PATH: " + podmanErr.Error())
	probeMock.EXPECT().DetectContainerEngine().Return(probe.ContainerEngineUnknown, detectErr)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	root := NewRoot(view.NewPrinter(nil, stdout, stderr), createDefaultContext(), nil, nil, nil, probeMock, nil)
	os.Args = []string{"globalping", "install-probe"}

	err := root.Cmd.ExecuteContext(t.Context())

	assert.ErrorIs(t, err, detectErr)
	assert.True(t, root.Cmd.SilenceUsage)
	assert.Empty(t, stdout.String())
	assert.Equal(t, "Error: no working container engine was detected; ensure Docker Desktop or Podman is installed and running:\n"+
		detectErr.Error()+"\n", stderr.String())
	assert.Equal(t, 1, strings.Count(stderr.String(), "Error:"))
}

func Test_Execute_Install_Probe_AlreadyInstalled(t *testing.T) {
	ctrl := gomock.NewController(t)
	probeMock := apiMocks.NewMockProbe(ctrl)
	installedErr := fmt.Errorf("%w. Current state: running; status: Up 3 hours", probe.ErrContainerAlreadyInstalled)
	probeMock.EXPECT().DetectContainerEngine().Return(probe.ContainerEngineDocker, nil)
	probeMock.EXPECT().InspectContainer(probe.ContainerEngineDocker).Return(installedErr)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	root := NewRoot(view.NewPrinter(nil, stdout, stderr), createDefaultContext(), nil, nil, nil, probeMock, nil)
	os.Args = []string{"globalping", "install-probe"}

	err := root.Cmd.ExecuteContext(t.Context())

	assert.NoError(t, err)
	assert.Equal(t, "Detected container engine: Docker\n\n"+installedErr.Error()+"\n", stdout.String())
	assert.Empty(t, stderr.String())
	assert.False(t, root.Cmd.SilenceUsage)
}

func Test_Execute_Install_Probe_PresenceFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	probeMock := apiMocks.NewMockProbe(ctrl)
	presenceErr := errors.New("failed to query containers")
	probeMock.EXPECT().DetectContainerEngine().Return(probe.ContainerEngineDocker, nil)
	probeMock.EXPECT().InspectContainer(probe.ContainerEngineDocker).Return(presenceErr)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	root := NewRoot(view.NewPrinter(nil, stdout, stderr), createDefaultContext(), nil, nil, nil, probeMock, nil)
	os.Args = []string{"globalping", "install-probe"}

	err := root.Cmd.ExecuteContext(t.Context())

	assert.ErrorIs(t, err, presenceErr)
	assert.True(t, root.Cmd.SilenceUsage)
	assert.Equal(t, "Detected container engine: Docker\n\n", stdout.String())
	assert.Equal(t, "Error: failed to query containers\n", stderr.String())
	assert.Equal(t, 1, strings.Count(stderr.String(), "Error:"))
}

func Test_Execute_Install_Probe_PromptFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	probeMock := apiMocks.NewMockProbe(ctrl)
	probeMock.EXPECT().DetectContainerEngine().Return(probe.ContainerEngineDocker, nil)
	probeMock.EXPECT().InspectContainer(probe.ContainerEngineDocker).Return(nil)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	root := NewRoot(view.NewPrinter(bytes.NewReader(nil), stdout, stderr), createDefaultContext(), nil, nil, nil, probeMock, nil)
	os.Args = []string{"globalping", "install-probe"}

	err := root.Cmd.ExecuteContext(t.Context())

	assert.ErrorIs(t, err, io.EOF)
	assert.True(t, root.Cmd.SilenceUsage)
	assert.Contains(t, stdout.String(), "[Y/n] ")
	assert.Equal(t, "Error: failed to read character: EOF\n", stderr.String())
	assert.Equal(t, 1, strings.Count(stderr.String(), "Error:"))
}

func Test_Execute_Install_Probe_Declined(t *testing.T) {
	ctrl := gomock.NewController(t)
	probeMock := apiMocks.NewMockProbe(ctrl)
	probeMock.EXPECT().DetectContainerEngine().Return(probe.ContainerEngineDocker, nil)
	probeMock.EXPECT().InspectContainer(probe.ContainerEngineDocker).Return(nil)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	root := NewRoot(view.NewPrinter(bytes.NewBufferString("n\n"), stdout, stderr), createDefaultContext(), nil, nil, nil, probeMock, nil)
	os.Args = []string{"globalping", "install-probe"}

	err := root.Cmd.ExecuteContext(t.Context())

	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "Exited without changes.\n")
	assert.Empty(t, stderr.String())
}

func Test_Execute_Install_Probe_RunFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	probeMock := apiMocks.NewMockProbe(ctrl)
	runErr := errors.New("container run failed: image pull denied")
	probeMock.EXPECT().DetectContainerEngine().Return(probe.ContainerEngineDocker, nil)
	probeMock.EXPECT().InspectContainer(probe.ContainerEngineDocker).Return(nil)
	probeMock.EXPECT().RunContainer(probe.ContainerEngineDocker).Return(runErr)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	root := NewRoot(view.NewPrinter(bytes.NewBufferString("Y\n"), stdout, stderr), createDefaultContext(), nil, nil, nil, probeMock, nil)
	os.Args = []string{"globalping", "install-probe"}

	err := root.Cmd.ExecuteContext(t.Context())

	assert.ErrorIs(t, err, runErr)
	assert.True(t, root.Cmd.SilenceUsage)
	assert.NotContains(t, stdout.String(), "started successfully")
	assert.Equal(t, "Error: container run failed: image pull denied\n", stderr.String())
	assert.Equal(t, 1, strings.Count(stderr.String(), "Error:"))
}

func Test_Execute_Install_Probe_PodmanSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	probeMock := apiMocks.NewMockProbe(ctrl)
	probeMock.EXPECT().DetectContainerEngine().Return(probe.ContainerEnginePodman, nil)
	probeMock.EXPECT().InspectContainer(probe.ContainerEnginePodman).Return(nil)
	probeMock.EXPECT().RunContainer(probe.ContainerEnginePodman).Return(nil)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	root := NewRoot(view.NewPrinter(bytes.NewBufferString("\n"), stdout, stderr), createDefaultContext(), nil, nil, nil, probeMock, nil)
	os.Args = []string{"globalping", "install-probe"}

	err := root.Cmd.ExecuteContext(t.Context())

	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "Detected container engine: Podman")
	assert.Contains(t, stdout.String(), "sudo podman")
	assert.Contains(t, stdout.String(), "The Globalping probe started successfully.")
	assert.Contains(t, stdout.String(), "install a service")
	assert.Empty(t, stderr.String())
}
