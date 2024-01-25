package view

import (
	"testing"

	"github.com/jsdelivr/globalping-cli/model"
	"github.com/stretchr/testify/assert"
)

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

func TestHeadersBase(t *testing.T) {
	assert.Equal(t, "> Continent, Country, (State), City, ASN:12345, Network", generateProbeInfo(&testResult, !testContext.CI))
}

func TestHeadersTags(t *testing.T) {
	newResult := testResult
	newResult.Probe.Tags = []string{"tag1", "tag2"}

	assert.Equal(t, "> Continent, Country, (State), City, ASN:12345, Network (tag1)", generateProbeInfo(&newResult, !testContext.CI))

	newResult.Probe.Tags = []string{"tag", "tag2"}
	assert.Equal(t, "> Continent, Country, (State), City, ASN:12345, Network (tag2)", generateProbeInfo(&newResult, !testContext.CI))
}

func TestTrimOutput(t *testing.T) {
	output := `> EU, GB, London, ASN:12345
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
LOREM IPSUM LOREM IPSUM LOREM IPSUM`

	res := trimOutput(output, 84, 11)

	expectedRes := `LOREM  IPSUM ♥ LOREM IPSUM LOREM IPSUM LOREM IPSUM LOREM IPSUM LOREM IPSUM LOREM
TEST OUTPUT 123456
IOPU
GHJKL
IOPU
GHJKL
LOREM IPSUM LOREM IPSUM LOREM IPSUM`

	assert.Equal(t, expectedRes, res)
}

func TestTrimOutput_CN(t *testing.T) {
	output := `> EU, GB, London, ASN:12345
some text a
中文互联文互联网高质量的问答社区和创 作者聚集的原创内容平台于201 1年1月正式上线让人们更 好的分享 知识经验和见解到自己的解答」中文互联网高质量的问答社区和创作者聚集的原创内容平台中文互联网高质量的问答社区和创作者聚集的原创内容平台于2011年1月正式上线让人们更好的分享知识经验和见解到自己的解答」中文互联网高质量的问答社区和创作者聚集的原创内容平台于
some text e
some text f`

	res := trimOutput(output, 84, 10)

	expectedRes := `> EU, GB, London, ASN:12345
some text a
中文互联文互联网高质量的问答社区和创 作者聚集的原创内容平台于201 1年1月正式上线让人们更 好的分享 知识经验和见解到自己的解答」中文互联网高质量的问答社区和创作者聚集的原
some text e
some text f`

	assert.Equal(t, expectedRes, res)
}
