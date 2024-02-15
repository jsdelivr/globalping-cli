package globalping

import (
	"net/http"
	"time"
)

type Client interface {
	CreateMeasurement(measurement *MeasurementCreate) (*MeasurementCreateResponse, bool, error)
	GetMeasurement(id string) (*Measurement, error)
	GetMeasurementRaw(id string) ([]byte, error)
}

type client struct {
	http   *http.Client
	apiUrl string // The api url endpoint

	etags        map[string]string // caches Etags by measurement id
	measurements map[string][]byte // caches Measurements by ETag
}

func NewClient(url string) Client {
	return &client{
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiUrl:       url,
		etags:        map[string]string{},
		measurements: map[string][]byte{},
	}
}
