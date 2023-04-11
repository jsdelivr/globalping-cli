package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/jsdelivr/globalping-cli/model"
	"github.com/stretchr/testify/assert"
)

func TestFetchWithEtag(t *testing.T) {
	id1 := "123abc"
	id2 := "567xyz"

	cacheMissCount := 0
	cacheHitCount := 0

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		id := parts[len(parts)-1]

		etag := func(id string) string {
			return "etag-" + id
		}

		if r.Header.Get("If-None-Match") == etag(id) {
			// cache hit
			cacheHitCount++
			w.Header().Set("ETag", etag(id))
			w.WriteHeader(http.StatusNotModified)

			return
		}

		// cache miss, return full response
		cacheMissCount++
		m := &model.GetMeasurement{
			ID: id,
		}

		w.Header().Set("ETag", etag(id))

		err := json.NewEncoder(w).Encode(m)
		assert.NoError(t, err)
	}))

	f := NewMeasurementsFetcher(s.URL)

	// first request for id1
	m, err := f.GetMeasurement(id1)
	assert.NoError(t, err)

	assert.Equal(t, id1, m.ID)

	// first request for id1
	m, err = f.GetMeasurement(id2)
	assert.NoError(t, err)

	assert.Equal(t, id2, m.ID)

	// second request for id1
	m, err = f.GetMeasurement(id2)
	assert.NoError(t, err)

	assert.Equal(t, id2, m.ID)

	assert.Equal(t, 1, cacheHitCount)
	assert.Equal(t, 2, cacheMissCount)
}

func TestFetchWithBrotli(t *testing.T) {
	id := "123abc"

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		id := parts[len(parts)-1]

		assert.Equal(t, "br", r.Header.Get("Accept-Encoding"))

		m := &model.GetMeasurement{
			ID: id,
		}

		w.Header().Set("Content-Encoding", "br")

		rW := brotli.NewWriter(w)
		defer rW.Close()

		err := json.NewEncoder(rW).Encode(m)
		assert.NoError(t, err)
	}))

	f := NewMeasurementsFetcher(s.URL)

	m, err := f.GetMeasurement(id)
	assert.NoError(t, err)

	assert.Equal(t, id, m.ID)
}
