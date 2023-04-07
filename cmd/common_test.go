package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInProgressUpdates_CI(t *testing.T) {
	ci := true
	assert.Equal(t, false, inProgressUpdates(ci))
}

func TestInProgressUpdates_NotCI(t *testing.T) {
	ci := false
	assert.Equal(t, true, inProgressUpdates(ci))
}
