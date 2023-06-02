package resty_test

import (
	"testing"
	"time"

	"github.com/omalloc/contrib/kratos/resty"
)

func TestClient(t *testing.T) {
	client := resty.New(
		resty.WithDebug(true),
		resty.WithTimeout(5*time.Second),
	)

	resp, err := client.R().Get("http://localhost:8080/api/app1/v1/hello")
	if err != nil {
		t.Error(err)
	}
	t.Log(resp)
}
