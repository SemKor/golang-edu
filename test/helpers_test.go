package test

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
)

func getTestToken(t *testing.T) string {
	t.Helper()

	e := httpexpect.Default(t, testServerURL)

	return e.POST("/v1/token").
		WithHeader("Content-Type", "application/x-www-form-urlencoded").
		WithBytes([]byte("grant_type=password&scope=read+write&username=tst&password=tst")).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("access_token").
		String().
		Raw()
}