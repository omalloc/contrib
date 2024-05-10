package health

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"

	"github.com/omalloc/contrib/kratos/resty"
)

func TestToSnake(t *testing.T) {
	got := "GreeterService"
	want := "greeter_service"

	if toSnake(got) != want {
		t.Errorf("toSnake(%q) = %q; want %q", got, toSnake(got), want)
	}
}

type errChecker struct {
	val uint8
}

func (m *errChecker) Check(ctx context.Context) error {
	if m.val <= 0 {
		m.val++
		return errors.New("check error.")
	}
	return nil
}

type okChecker struct{}

func (m *okChecker) Check(ctx context.Context) error {
	return nil
}

func TestHealthService(t *testing.T) {
	checkers := []Checker{
		// ok checker
		&okChecker{},
		// error checker
		&errChecker{},
	}

	httpSrv := http.NewServer(http.Address(":60180"))
	s := NewServer(checkers, log.NewStdLogger(io.Discard), httpSrv)
	s.Start(context.Background())

	go func() {
		if err := httpSrv.Start(context.Background()); err != nil {
			panic(err)
		}
	}()

	time.Sleep(1 * time.Second)
	resp, err := resty.New().NewRequest().Get("http://127.0.0.1:60180/health")
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.String())

	if resp.StatusCode() != 503 {
		t.Errorf("expect status code 503, got %d", resp.StatusCode())
	}

	time.Sleep(1 * time.Second)
	resp, err = resty.New().NewRequest().Get("http://127.0.0.1:60180/health")
	if err != nil {
		panic(err)
	}

	fmt.Println(resp.String())

	if resp.StatusCode() != 200 {
		t.Errorf("expect status code 200, got %d", resp.StatusCode())
	}

	if err := s.Stop(context.Background()); err != nil {
		panic(err)
	}
}
