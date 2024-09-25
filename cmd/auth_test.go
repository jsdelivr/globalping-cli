package cmd

import (
	"bytes"
	"context"
	"math"
	"os"
	"syscall"
	"testing"

	"github.com/jsdelivr/globalping-cli/globalping"
	"github.com/jsdelivr/globalping-cli/mocks"
	"github.com/jsdelivr/globalping-cli/storage"
	"github.com/jsdelivr/globalping-cli/view"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_Auth_Login_WithToken(t *testing.T) {
	t.Cleanup(sessionCleanup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gbMock := mocks.NewMockClient(ctrl)

	utilsMock := mocks.NewMockUtils(ctrl)
	utilsMock.EXPECT().Now().Return(defaultCurrentTime).AnyTimes()

	w := new(bytes.Buffer)
	r := new(bytes.Buffer)
	r.WriteString("token\n")
	printer := view.NewPrinter(r, w, w)
	ctx := createDefaultContext("")
	_storage := storage.NewLocalStorage(".test_globalping-cli")
	defer _storage.Remove()
	err := _storage.Init()
	if err != nil {
		t.Fatal(err)
	}
	_storage.GetProfile().Token = &globalping.Token{
		AccessToken:  "oldToken",
		RefreshToken: "oldRefreshToken",
	}

	root := NewRoot(printer, ctx, nil, utilsMock, gbMock, nil, _storage)

	gbMock.EXPECT().TokenIntrospection("token").Return(&globalping.IntrospectionResponse{
		Active:   true,
		Username: "test",
	}, nil)
	gbMock.EXPECT().RevokeToken("oldRefreshToken").Return(nil)

	os.Args = []string{"globalping", "auth", "login", "--with-token"}
	err = root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	assert.Equal(t, `Please enter your token:
Logged in as test.
`, w.String())

	profile := _storage.GetProfile()
	assert.Equal(t, &storage.Profile{
		Token: &globalping.Token{
			AccessToken: "token",
			Expiry:      defaultCurrentTime.Add(math.MaxInt64),
		},
	}, profile)
}

func Test_Auth_Login(t *testing.T) {
	t.Cleanup(sessionCleanup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gbMock := mocks.NewMockClient(ctrl)
	utilsMock := mocks.NewMockUtils(ctrl)

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext("")
	_storage := storage.NewLocalStorage(".test_globalping-cli")
	defer _storage.Remove()
	err := _storage.Init()
	if err != nil {
		t.Fatal(err)
	}
	_storage.GetProfile().Token = &globalping.Token{
		AccessToken:  "oldToken",
		RefreshToken: "oldRefreshToken",
	}

	root := NewRoot(printer, ctx, nil, utilsMock, gbMock, nil, _storage)

	gbMock.EXPECT().Authorize(gomock.Any()).Do(func(_ any) {
		root.cancel <- syscall.SIGINT
	}).Return(&globalping.AuthorizeResponse{
		AuthorizeURL: "http://localhost",
	}, nil)
	utilsMock.EXPECT().OpenBrowser("http://localhost").Return(nil)

	os.Args = []string{"globalping", "auth", "login"}
	err = root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	assert.Equal(t, `Please visit the following URL to authenticate:
http://localhost

Can't use the browser-based flow? Use "globalping auth login --with-token" to read a token from stdin instead.
`, w.String())
}

func Test_AuthStatus(t *testing.T) {
	t.Cleanup(sessionCleanup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gbMock := mocks.NewMockClient(ctrl)

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext("")

	root := NewRoot(printer, ctx, nil, nil, gbMock, nil, nil)

	gbMock.EXPECT().TokenIntrospection("").Return(&globalping.IntrospectionResponse{
		Active:   true,
		Username: "test",
	}, nil)

	os.Args = []string{"globalping", "auth", "status"}
	err := root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	assert.Equal(t, `Logged in as test.
`, w.String())
}

func Test_Logout(t *testing.T) {
	t.Cleanup(sessionCleanup)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	gbMock := mocks.NewMockClient(ctrl)

	w := new(bytes.Buffer)
	printer := view.NewPrinter(nil, w, w)
	ctx := createDefaultContext("")

	root := NewRoot(printer, ctx, nil, nil, gbMock, nil, nil)

	gbMock.EXPECT().Logout().Return(nil)

	os.Args = []string{"globalping", "auth", "logout"}
	err := root.Cmd.ExecuteContext(context.TODO())
	assert.NoError(t, err)

	assert.Equal(t, "You are now logged out.\n", w.String())
}
