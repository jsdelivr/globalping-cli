package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetIdFromSession(t *testing.T) {
	_storage := createDefaultTestStorage(t, nil)
	os.WriteFile(_storage.measurementsPath(), []byte("id1\nid2\nid3\n"), 0644)
	id, err := _storage.GetIdFromSession(1)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "id1", id)

	id, err = _storage.GetIdFromSession(2)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "id2", id)

	id, err = _storage.GetIdFromSession(-1)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "id3", id)
}

func Test_SaveIdToSession(t *testing.T) {
	_storage := createDefaultTestStorage(t, nil)
	err := _storage.SaveIdToSession("id")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(_storage.measurementsPath())
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "id\n", string(data))
}
