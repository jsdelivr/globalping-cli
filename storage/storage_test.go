package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_truncateFile(t *testing.T) {
	file := "globalping_truncate_test"
	os.WriteFile(file, []byte(`CrkT2oK70XgKQRPT
io57ICA41VN5DPhh
JDHFAGabKrAYvp7i
rInFNLFr3Tzj43FO
EAQ8LpKfXkfBPdUG
`), 0644)
	defer os.Remove(file)

	err := truncateFile(file, 17*4+1)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(file)
	assert.Nil(t, err)
	assert.Equal(t, `io57ICA41VN5DPhh
JDHFAGabKrAYvp7i
rInFNLFr3Tzj43FO
EAQ8LpKfXkfBPdUG
`, string(b))

	err = truncateFile(file, 17*3)
	if err != nil {
		t.Fatal(err)
	}
	b, err = os.ReadFile(file)
	assert.Nil(t, err)
	assert.Equal(t, `rInFNLFr3Tzj43FO
EAQ8LpKfXkfBPdUG
`, string(b))

	err = truncateFile(file, 17*2)
	if err != nil {
		t.Fatal(err)
	}
	b, err = os.ReadFile(file)
	assert.Nil(t, err)
	assert.Equal(t, `rInFNLFr3Tzj43FO
EAQ8LpKfXkfBPdUG
`, string(b))

	err = truncateFile(file, 4)
	if err != nil {
		t.Fatal(err)
	}
	b, err = os.ReadFile(file)
	assert.Nil(t, err)
	assert.Equal(t, ``, string(b))
}
