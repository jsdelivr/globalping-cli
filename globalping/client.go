package globalping

import (
	"net/http"
	"time"

	"github.com/jsdelivr/globalping-cli/utils"
)

type Client interface {
	CreateMeasurement(measurement *MeasurementCreate) (*MeasurementCreateResponse, bool, error)
	GetMeasurement(id string) (*Measurement, error)
	GetMeasurementRaw(id string) ([]byte, error)
}

type client struct {
	http   *http.Client
	config *utils.Config

	etags        map[string]string // caches Etags by measurement id
	measurements map[string][]byte // caches Measurements by ETag
}

func NewClient(config *utils.Config) Client {
	return &client{
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
		config:       config,
		etags:        map[string]string{},
		measurements: map[string][]byte{},
	}
}
