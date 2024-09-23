package globalping

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Limits(t *testing.T) {
	expectedResponse := &LimitsResponse{
		RateLimits: RateLimits{
			Measurements: MeasurementsLimits{
				Create: MeasurementsCreateLimits{
					Type:      CreateLimitTypeUser,
					Limit:     1000,
					Remaining: 999,
					Reset:     600,
				},
			},
		},
		Credits: CreditLimits{
			Remaining: 1000,
		},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/limits" && r.Method == http.MethodGet {
			assert.Equal(t, "Bearer tok3n", r.Header.Get("Authorization"))
			w.Header().Set("Content-Type", "application/json")
			b, _ := json.Marshal(expectedResponse)
			_, err := w.Write(b)
			if err != nil {
				t.Fatal(err)
			}
			return
		}
		t.Fatalf("unexpected request to %s", r.URL.Path)
	}))
	defer server.Close()

	client := NewClient(Config{
		AuthClientID:     "<client_id>",
		AuthClientSecret: "<client_secret>",
		AuthURL:          server.URL,
		DashboardURL:     server.URL,
		APIURL:           server.URL,
		AuthToken: &Token{
			AccessToken: "tok3n",
			Expiry:      time.Now().Add(time.Hour),
		},
	})
	res, err := client.Limits()
	assert.Nil(t, err)
	assert.Equal(t, expectedResponse, res)
}
