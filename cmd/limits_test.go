package cmd

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/jsdelivr/globalping-cli/mocks"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_Limits_User(t *testing.T) {
	t.Cleanup(sessionCleanup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gbMock := mocks.NewMockClient(ctrl)

	gbMock.EXPECT().TokenIntrospection("").Return(&globalping.IntrospectionResponse{
		Active:   true,
		Username: "test",
	}, nil)
	gbMock.EXPECT().Limits().Return(&globalping.LimitsResponse{
		RateLimits: globalping.RateLimits{
			Measurements: globalping.MeasurementsLimits{
				Create: globalping.MeasurementsCreateLimits{
					Type:      "user",
					Limit:     500,
					Remaining: 350,
					Reset:     600,
				},
			},
		},
		Credits: globalping.CreditLimits{
			Remaining: 1000,
		},
	}, nil)

	w := new(bytes.Buffer)
	r := new(bytes.Buffer)
	printer := view.NewPrinter(r, w, w)
	ctx := createDefaultContext("")

	root := NewRoot(printer, ctx, nil, nil, gbMock, nil, nil)

	os.Args = []string{"globalping", "limits"}
	err := root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	assert.Equal(t, `Authentication: token (test)

Creating measurements: 
 - 500 tests per hour
 - 150 consumed, 350 remaining
 - resets in 10 minutes

Credits:
 - 1000 credits remaining (may be used to create measurements above the hourly limits)
`, w.String())
}

func Test_Limits_IP(t *testing.T) {
	t.Cleanup(sessionCleanup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gbMock := mocks.NewMockClient(ctrl)

	gbMock.EXPECT().TokenIntrospection("").Return(nil, &globalping.AuthorizeError{Description: "client is not authorized"})
	gbMock.EXPECT().Limits().Return(&globalping.LimitsResponse{
		RateLimits: globalping.RateLimits{
			Measurements: globalping.MeasurementsLimits{
				Create: globalping.MeasurementsCreateLimits{
					Type:      "ip",
					Limit:     500,
					Remaining: 350,
					Reset:     600,
				},
			},
		},
	}, nil)

	w := new(bytes.Buffer)
	r := new(bytes.Buffer)
	printer := view.NewPrinter(r, w, w)
	ctx := createDefaultContext("")

	root := NewRoot(printer, ctx, nil, nil, gbMock, nil, nil)

	os.Args = []string{"globalping", "limits"}
	err := root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	assert.Equal(t, `Authentication: IP address

Creating measurements: 
 - 500 tests per hour
 - 150 consumed, 350 remaining
 - resets in 10 minutes
`, w.String())
}
