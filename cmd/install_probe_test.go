package cmd

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/jsdelivr/globalping-cli/globalping/probe"
	"github.com/jsdelivr/globalping-cli/mocks"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_Execute_Install_Probe_Docker(t *testing.T) {
	t.Cleanup(sessionCleanup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	probeMock := mocks.NewMockProbe(ctrl)
	probeMock.EXPECT().DetectContainerEngine().Times(1).Return(probe.ContainerEngineDocker, nil)
	probeMock.EXPECT().InspectContainer(probe.ContainerEngineDocker).Times(1).Return(nil)
	probeMock.EXPECT().RunContainer(probe.ContainerEngineDocker).Times(1).Return(nil)

	reader := bytes.NewReader([]byte("Y\n"))
	w := new(bytes.Buffer)
	printer := view.NewPrinter(reader, w, w)
	ctx := &view.Context{}
	root := NewRoot(printer, ctx, nil, nil, nil, probeMock)
	os.Args = []string{"globalping", "install-probe"}
	err := root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, `Detected container engine: Docker

The Globalping platform is a community powered project and relies on individuals like yourself to host our probes and make them accessible to everyone else.
Please confirm to pull and run our Docker container (ghcr.io/jsdelivr/globalping-probe) [Y/n] The Globalping probe started successfully. Thank you for joining our community! 
`, w.String())

	expectedCtx := &view.Context{
		From:  "world",
		Limit: 1,
	}
	assert.Equal(t, expectedCtx, ctx)
}
