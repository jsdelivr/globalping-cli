package cmd

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/jsdelivr/globalping-cli/version"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/stretchr/testify/assert"
)

func Test_Execute_Version_Default(t *testing.T) {
	version.Version = "1.0.0"
	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	root := NewRoot(printer, &view.Context{}, nil, nil, nil, nil)

	os.Args = []string{"globalping", "version"}
	err := root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	assert.Equal(t, "Globalping CLI v1.0.0\n", w.String())
}
