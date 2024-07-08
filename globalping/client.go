package globalping

import (
	"net/http"
	"time"

	"github.com/jsdelivr/globalping-cli/utils"
)

type Client interface {
	// Creates a new measurement with parameters set in the request body. The measurement runs asynchronously and you can retrieve its current state at the URL returned in the Location header.
	//
	// https://www.jsdelivr.com/docs/api.globalping.io#post
	CreateMeasurement(measurement *MeasurementCreate) (*MeasurementCreateResponse, error)
	// Returns the status and results of an existing measurement. Measurements are typically available for up to 7 days after creation.
	//
	// https://www.jsdelivr.com/docs/api.globalping.io#get
	GetMeasurement(id string) (*Measurement, error)
	// Returns the status and results of an existing measurement. Measurements are typically available for up to 7 days after creation.
	//
	// https://www.jsdelivr.com/docs/api.globalping.io#get
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
