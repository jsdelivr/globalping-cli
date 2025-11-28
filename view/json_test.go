package view

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Output_Json(t *testing.T) {
	b := []byte(`{"fake": "results"}`)

	w := new(bytes.Buffer)
	printer := NewPrinter(nil, w, w)
	printer.DisableStyling()
	viewer := NewViewer(
		&Context{
			ToJSON: true,
			Share:  true,
		},
		printer,
		nil,
	)

	viewer.OutputJSON(measurementID1, b)

	assert.Equal(t, fmt.Sprintf(`{"fake": "results"}
> View the results online: https://globalping.io?measurement=%s

`, measurementID1), w.String())
}
