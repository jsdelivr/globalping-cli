package client

import (
	"testing"

	"github.com/jsdelivr/globalping-cli/model"
	"github.com/stretchr/testify/assert"
)

func TestGenerateHeaders(t *testing.T) {
	for scenario, fn := range map[string]func(t *testing.T){
		"base": testHeadersBase,
		"tags": testHeadersTags,
	} {
		t.Run(scenario, func(t *testing.T) {
			fn(t)
		})
	}
}

var (
	testContext = model.Context{
		From:   "New York",
		Target: "1.1.1.1",
		CI:     true,
	}
	testResult = model.MeasurementResponse{
		Probe: model.ProbeData{
			Continent: "Continent",
			Country:   "Country",
			State:     "State",
			City:      "City",
			ASN:       12345,
			Network:   "Network",
			Tags:      []string{"tag"},
		},
	}
)

func testHeadersBase(t *testing.T) {
	assert.Equal(t, "> Continent, Country, (State), City, ASN:12345, Network", generateHeader(testResult, testContext))
}

func testHeadersTags(t *testing.T) {
	newResult := testResult
	newResult.Probe.Tags = []string{"tag1", "tag2"}

	assert.Equal(t, "> Continent, Country, (State), City, ASN:12345, Network (tag1)", generateHeader(newResult, testContext))

	newResult.Probe.Tags = []string{"tag", "tag2"}
	assert.Equal(t, "> Continent, Country, (State), City, ASN:12345, Network (tag2)", generateHeader(newResult, testContext))
}
