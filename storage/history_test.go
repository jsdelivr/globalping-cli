package storage

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_GetHistoryIndex(t *testing.T) {
	_storage := NewLocalStorage(nil)
	defer _storage.Remove()
	err := _storage.Init(".test_globalping-cli")
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(_storage.historyPath(), []byte("1|1|1|id|command\n"), 0644)
	index, err := _storage.GetHistoryIndex()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 2, index)
}

func Test_GetHistory(t *testing.T) {
	_storage := NewLocalStorage(nil)
	defer _storage.Remove()
	err := _storage.Init(".test_globalping-cli")
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(_storage.historyPath(), []byte(`1|1|1730310880|id1|command1
1|2|1730310890|id2|command2
`), 0644)

	items, err := _storage.GetHistory(0)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []string{
		"1 | 2024-10-30 19:54:40 | command1\n> https://globalping.io?measurement=id1",
		"2 | 2024-10-30 19:54:50 | command2\n> https://globalping.io?measurement=id2",
	}, items)

	items, err = _storage.GetHistory(1)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []string{
		"1 | 2024-10-30 19:54:40 | command1\n> https://globalping.io?measurement=id1",
	}, items)

	items, err = _storage.GetHistory(-1)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []string{
		"2 | 2024-10-30 19:54:50 | command2\n> https://globalping.io?measurement=id2",
	}, items)
}

func Test_SaveCommandToHistory(t *testing.T) {
	_storage := NewLocalStorage(nil)
	defer _storage.Remove()
	err := _storage.Init(".test_globalping-cli")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	err = _storage.SaveCommandToHistory(
		"1",
		now.Unix(),
		"id1",
		"command1",
	)
	if err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(_storage.historyPath())
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t,
		fmt.Sprintf(`1|1|%d|id1|command1
`, now.Unix()), string(b))
}
