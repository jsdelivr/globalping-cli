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
	t.Cleanup(func() {
		_storage.Remove()
	})
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
	t.Cleanup(func() {
		_storage.Remove()
	})
	err := _storage.Init(".test_globalping-cli")
	if err != nil {
		t.Fatal(err)
	}
	time1 := time.Unix(1730310880, 0)
	time2 := time.Unix(1730310890, 0)
	os.WriteFile(_storage.historyPath(), []byte(fmt.Sprintf(`1|1|%d|id1|command1
1|2|%d|id2|command2
`, time1.Unix(), time2.Unix())), 0644)

	items, err := _storage.GetHistory(0)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []string{
		fmt.Sprintf("1 | %s | command1\n> https://globalping.io?measurement=id1", time1.Format("2006-01-02 15:04:05")),
		fmt.Sprintf("2 | %s | command2\n> https://globalping.io?measurement=id2", time2.Format("2006-01-02 15:04:05")),
	}, items)

	items, err = _storage.GetHistory(1)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []string{
		fmt.Sprintf("1 | %s | command1\n> https://globalping.io?measurement=id1", time1.Format("2006-01-02 15:04:05")),
	}, items)

	items, err = _storage.GetHistory(-1)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []string{
		fmt.Sprintf("2 | %s | command2\n> https://globalping.io?measurement=id2", time2.Format("2006-01-02 15:04:05")),
	}, items)
}

func Test_SaveCommandToHistory(t *testing.T) {
	_storage := NewLocalStorage(nil)
	t.Cleanup(func() {
		_storage.Remove()
	})
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
