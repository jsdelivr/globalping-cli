package cmd

import (
	"bytes"
	"os"
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
	ctx := createDefaultContext("install-probe")
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
