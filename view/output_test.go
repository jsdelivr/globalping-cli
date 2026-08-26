package view

import (
	"strings"
	"testing"

	"github.com/jsdelivr/globalping-go"
	"github.com/stretchr/testify/assert"
)

var (
	testResult = globalping.ProbeMeasurement{
		Probe: globalping.ProbeDetails{
			Continent: "Continent",
			Country:   "Country",
			State:     pointerTo("State"),
			City:      "City",
			ASN:       12345,
			Network:   "Network",
			Tags:      []string{"tag"},
		},
	}
)

func Test_HeadersBase(t *testing.T) {
	v := viewer{
		ctx:     &Context{},
		printer: NewPrinter(nil, nil, nil),
	}
	assert.Equal(t, "\033[1;38;5;43m> City (State), Country, Continent, Network (AS12345)\033[0m", v.getProbeInfo(&testResult))
}

func Test_HeadersTags(t *testing.T) {
	newResult := testResult
	newResult.Probe.Tags = []string{"tag1", "tag2"}
	printer := NewPrinter(nil, nil, nil)
	printer.DisableStyling()
	v := viewer{
		ctx:     &Context{},
		printer: printer,
	}

	assert.Equal(t, "> City (State), Country, Continent, Network (AS12345) (tag1)", v.getProbeInfo(&newResult))

	newResult.Probe.Tags = []string{"u-JohnDoe1", "tag2", "u-JohnDoe2"}
	assert.Equal(t, "> City (State), Country, Continent, Network (AS12345), u-JohnDoe (tag2)", v.getProbeInfo(&newResult))
}

func Test_OutputLive_NilHTTPBodiesAndProbeStates(t *testing.T) {
	w := new(strings.Builder)
	printer := NewPrinter(nil, w, w)
	printer.DisableStyling()
	v := viewer{
		ctx:     &Context{Cmd: "http"},
		printer: printer,
	}
	measurement := &globalping.Measurement{
		Results: []globalping.ProbeMeasurement{
			{
				Probe: globalping.ProbeDetails{
					Continent: "EU",
					Country:   "DE",
					City:      "Berlin",
					ASN:       123,
					Network:   "Network",
				},
			},
			{
				Probe: globalping.ProbeDetails{
					Continent: "NA",
					Country:   "US",
					City:      "New York",
					ASN:       456,
					Network:   "Other Network",
				},
			},
		},
	}
	opts := &globalping.MeasurementCreate{
		Options: &globalping.MeasurementOptions{
			Request: &globalping.RequestOptions{Method: "GET"},
		},
	}

	v.OutputLive(measurement, opts, 80, 24)

	assert.Equal(t, "> Berlin, DE, EU, Network (AS123)\n\n\n> New York, US, NA, Other Network (AS456)\n\n\n", w.String())
	assert.NotContains(t, w.String(), "%!")
}

func Test_TrimOutput(t *testing.T) {
	output := &strings.Builder{}
	output.WriteString(`> London, GB, EU, Network (AS12345)
TEST CONTENT
ABCD
EDF
XYZ
LOREM	IPSUM ♥ LOREM IPSUM LOREM IPSUM LOREM IPSUM LOREM IPSUM LOREM IPSUM LOREM IPSUM LOREM IPSUM
TEST OUTPUT 123456
IOPU
GHJKL
IOPU
GHJKL
LOREM IPSUM LOREM IPSUM LOREM IPSUM`)

	res := trimOutput(output, 84, 11)

	expectedRes := `LOREM  IPSUM ♥ LOREM IPSUM LOREM IPSUM LOREM IPSUM LOREM IPSUM LOREM IPSUM LOREM
TEST OUTPUT 123456
IOPU
GHJKL
IOPU
GHJKL
LOREM IPSUM LOREM IPSUM LOREM IPSUM`

	assert.Equal(t, expectedRes, *res)
}

func Test_TrimOutput_CN(t *testing.T) {
	output := &strings.Builder{}
	output.WriteString(`> London, GB, EU, Network (AS12345)
some text a
中文互联文互联网高质量的问答社区和创 作者聚集的原创内容平台于201 1年1月正式上线让人们更 好的分享 知识经验和见解到自己的解答」中文互联网高质量的问答社区和创作者聚集的原创内容平台中文互联网高质量的问答社区和创作者聚集的原创内容平台于2011年1月正式上线让人们更好的分享知识经验和见解到自己的解答」中文互联网高质量的问答社区和创作者聚集的原创内容平台于
some text e
some text f`)

	res := trimOutput(output, 84, 10)

	expectedRes := `> London, GB, EU, Network (AS12345)
some text a
中文互联文互联网高质量的问答社区和创 作者聚集的原创内容平台于201 1年1月正式上线让人们更 好的分享 知识经验和见解到自己的解答」中文互联网高质量的问答社区和创作者聚集的原
some text e
some text f`

	assert.Equal(t, expectedRes, *res)
}
