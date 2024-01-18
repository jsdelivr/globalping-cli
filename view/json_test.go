package view

import (
	"io"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jsdelivr/globalping-cli/mocks"
	"github.com/jsdelivr/globalping-cli/model"
	"github.com/stretchr/testify/assert"
)

func TestOutputJson(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	id := "my-id"

	b := []byte(`{"fake": "results"}`)

	fetcher := mocks.NewMockMeasurementsFetcher(ctrl)
	fetcher.EXPECT().GetRawMeasurement(id).Times(1).Return(b, nil)

	ctx := model.Context{
		JsonOutput: true,
		Share:      true,
	}
	osStdErr := os.Stderr
	osStdOut := os.Stdout

	rStdErr, myStdErr, err := os.Pipe()
	assert.NoError(t, err)
	defer rStdErr.Close()

	rStdOut, myStdOut, err := os.Pipe()
	assert.NoError(t, err)
	defer rStdOut.Close()

	os.Stderr = myStdErr
	os.Stdout = myStdOut

	defer func() {
		os.Stderr = osStdErr
		os.Stdout = osStdOut
	}()

	err = OutputJson(id, fetcher, ctx)
	assert.NoError(t, err)
	myStdOut.Close()
	myStdErr.Close()

	errContent, err := io.ReadAll(rStdErr)
	assert.NoError(t, err)
	assert.Equal(t, "> View the results online: https://www.jsdelivr.com/globalping?measurement=my-id\n", string(errContent))

	outContent, err := io.ReadAll(rStdOut)
	assert.NoError(t, err)
	assert.Equal(t, "{\"fake\": \"results\"}\n\n", string(outContent))
}
