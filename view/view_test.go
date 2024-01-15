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
	assert.Equal(t, "> Continent, Country, (State), City, ASN:12345, Network", generateHeader(testResult, testContext))
}

func TestHeadersTags(t *testing.T) {
	newResult := testResult
	newResult.Probe.Tags = []string{"tag1", "tag2"}

	assert.Equal(t, "> Continent, Country, (State), City, ASN:12345, Network (tag1)", generateHeader(newResult, testContext))

	newResult.Probe.Tags = []string{"tag", "tag2"}
	assert.Equal(t, "> Continent, Country, (State), City, ASN:12345, Network (tag2)", generateHeader(newResult, testContext))
}

func TestPrintStandardResultsHTTPGet(t *testing.T) {
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

	ctx := model.Context{
		Cmd: "http",
		CI:  true,
	}

	m := model.PostMeasurement{
		Options: &model.MeasurementOptions{
			Request: &model.RequestOptions{
				Method: "GET",
			},
		},
	}

	id := "123abc"

	data := &model.GetMeasurement{
		Results: []model.MeasurementResponse{
			{
				Probe: model.ProbeData{
					Continent: "EU",
					Country:   "DE",
					City:      "Berlin",
					ASN:       123,
					Network:   "Network 1",
				},
				Result: model.ResultData{
					RawOutput:  "Headers 1\nBody 1",
					RawHeaders: "Headers 1",
					RawBody:    "Body 1",
				},
			},

			{
				Probe: model.ProbeData{
					Continent: "NA",
					Country:   "US",
					City:      "New York",
					State:     "NY",
					ASN:       567,
					Network:   "Network 2",
				},
				Result: model.ResultData{
					RawOutput:  "Headers 2\nBody 2",
					RawHeaders: "Headers 2",
					RawBody:    "Body 2",
				},
			},
		},
	}

	PrintStandardResults(id, data, ctx, m)
	myStdOut.Close()
	myStdErr.Close()

	errContent, err := io.ReadAll(rStdErr)
	assert.NoError(t, err)
	assert.Equal(t, "> EU, DE, Berlin, ASN:123, Network 1\n> NA, US, (NY), New York, ASN:567, Network 2\n", string(errContent))

	outContent, err := io.ReadAll(rStdOut)
	assert.NoError(t, err)
	assert.Equal(t, "Body 1\n\nBody 2\n", string(outContent))
}

func TestPrintStandardResultsHTTPGetShare(t *testing.T) {
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

	ctx := model.Context{
		Cmd:   "http",
		CI:    true,
		Share: true,
	}

	m := model.PostMeasurement{
		Options: &model.MeasurementOptions{
			Request: &model.RequestOptions{
				Method: "GET",
			},
		},
	}

	id := "123abc"

	data := &model.GetMeasurement{
		Results: []model.MeasurementResponse{
			{
				Probe: model.ProbeData{
					Continent: "EU",
					Country:   "DE",
					City:      "Berlin",
					ASN:       123,
					Network:   "Network 1",
				},
				Result: model.ResultData{
					RawOutput:  "Headers 1\nBody 1",
					RawHeaders: "Headers 1",
					RawBody:    "Body 1",
				},
			},

			{
				Probe: model.ProbeData{
					Continent: "NA",
					Country:   "US",
					City:      "New York",
					State:     "NY",
					ASN:       567,
					Network:   "Network 2",
				},
				Result: model.ResultData{
					RawOutput:  "Headers 2\nBody 2",
					RawHeaders: "Headers 2",
					RawBody:    "Body 2",
				},
			},
		},
	}

	PrintStandardResults(id, data, ctx, m)
	myStdOut.Close()
	myStdErr.Close()

	errContent, err := io.ReadAll(rStdErr)
	assert.NoError(t, err)
	assert.Equal(t, "> EU, DE, Berlin, ASN:123, Network 1\n> NA, US, (NY), New York, ASN:567, Network 2\n> View the results online: https://www.jsdelivr.com/globalping?measurement=123abc\n", string(errContent))

	outContent, err := io.ReadAll(rStdOut)
	assert.NoError(t, err)
	assert.Equal(t, "Body 1\n\nBody 2\n", string(outContent))
}

func TestPrintStandardResultsHTTPGetFull(t *testing.T) {
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

	ctx := model.Context{
		Cmd:  "http",
		CI:   true,
		Full: true,
	}

	m := model.PostMeasurement{
		Options: &model.MeasurementOptions{
			Request: &model.RequestOptions{
				Method: "GET",
			},
		},
	}

	id := "123abc"

	data := &model.GetMeasurement{
		Results: []model.MeasurementResponse{
			{
				Probe: model.ProbeData{
					Continent: "EU",
					Country:   "DE",
					City:      "Berlin",
					ASN:       123,
					Network:   "Network 1",
				},
				Result: model.ResultData{
					RawOutput:  "Headers 1\nBody 1",
					RawHeaders: "Headers 1",
					RawBody:    "Body 1",
				},
			},

			{
				Probe: model.ProbeData{
					Continent: "NA",
					Country:   "US",
					City:      "New York",
					State:     "NY",
					ASN:       567,
					Network:   "Network 2",
				},
				Result: model.ResultData{
					RawOutput:  "Headers 2\nBody 2",
					RawHeaders: "Headers 2",
					RawBody:    "Body 2",
				},
			},
		},
	}

	PrintStandardResults(id, data, ctx, m)
	myStdOut.Close()
	myStdErr.Close()

	errContent, err := io.ReadAll(rStdErr)
	assert.NoError(t, err)
	assert.Equal(t, "> EU, DE, Berlin, ASN:123, Network 1\n> NA, US, (NY), New York, ASN:567, Network 2\n", string(errContent))

	outContent, err := io.ReadAll(rStdOut)
	assert.NoError(t, err)
	assert.Equal(t, "Headers 1\nBody 1\n\nHeaders 2\nBody 2\n", string(outContent))
}

func TestPrintStandardResultsHTTPHead(t *testing.T) {
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

	ctx := model.Context{
		Cmd: "http",
		CI:  true,
	}

	m := model.PostMeasurement{
		Options: &model.MeasurementOptions{
			Request: &model.RequestOptions{
				Method: "HEAD",
			},
		},
	}

	id := "123abc"

	data := &model.GetMeasurement{
		Results: []model.MeasurementResponse{
			{
				Probe: model.ProbeData{
					Continent: "EU",
					Country:   "DE",
					City:      "Berlin",
					ASN:       123,
					Network:   "Network 1",
				},
				Result: model.ResultData{
					RawOutput:  "Headers 1",
					RawHeaders: "Headers 1",
				},
			},

			{
				Probe: model.ProbeData{
					Continent: "NA",
					Country:   "US",
					City:      "New York",
					State:     "NY",
					ASN:       567,
					Network:   "Network 2",
				},
				Result: model.ResultData{
					RawOutput:  "Headers 2",
					RawHeaders: "Headers 2",
				},
			},
		},
	}

	PrintStandardResults(id, data, ctx, m)
	myStdOut.Close()
	myStdErr.Close()

	errContent, err := io.ReadAll(rStdErr)
	assert.NoError(t, err)
	assert.Equal(t, "> EU, DE, Berlin, ASN:123, Network 1\n> NA, US, (NY), New York, ASN:567, Network 2\n", string(errContent))

	outContent, err := io.ReadAll(rStdOut)
	assert.NoError(t, err)
	assert.Equal(t, "Headers 1\n\nHeaders 2\n", string(outContent))
}

func TestPrintStandardResultsPing(t *testing.T) {
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

	ctx := model.Context{
		Cmd: "ping",
		CI:  true,
	}

	m := model.PostMeasurement{}

	id := "123abc"

	data := &model.GetMeasurement{
		Results: []model.MeasurementResponse{
			{
				Probe: model.ProbeData{
					Continent: "EU",
					Country:   "DE",
					City:      "Berlin",
					ASN:       123,
					Network:   "Network 1",
				},
				Result: model.ResultData{
					RawOutput: "Ping Results 1",
				},
			},

			{
				Probe: model.ProbeData{
					Continent: "NA",
					Country:   "US",
					City:      "New York",
					State:     "NY",
					ASN:       567,
					Network:   "Network 2",
				},
				Result: model.ResultData{
					RawOutput: "Ping Results 2",
				},
			},
		},
	}

	PrintStandardResults(id, data, ctx, m)
	myStdOut.Close()
	myStdErr.Close()

	errContent, err := io.ReadAll(rStdErr)
	assert.NoError(t, err)
	assert.Equal(t, "> EU, DE, Berlin, ASN:123, Network 1\n> NA, US, (NY), New York, ASN:567, Network 2\n", string(errContent))

	outContent, err := io.ReadAll(rStdOut)
	assert.NoError(t, err)
	assert.Equal(t, "Ping Results 1\n\nPing Results 2\n", string(outContent))
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

	OutputJson(id, fetcher, ctx)
	myStdOut.Close()
	myStdErr.Close()

	errContent, err := io.ReadAll(rStdErr)
	assert.NoError(t, err)
	assert.Equal(t, "> View the results online: https://www.jsdelivr.com/globalping?measurement=my-id\n", string(errContent))

	outContent, err := io.ReadAll(rStdOut)
	assert.NoError(t, err)
	assert.Equal(t, "{\"fake\": \"results\"}\n\n", string(outContent))
}
