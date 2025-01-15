package gin_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/go-resty/resty/v2"

	kgin "github.com/omalloc/contrib/kratos/gin"
)

func ExampleMiddlewares() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(kgin.Middlewares(
		recovery.Recovery(),
		tracing.Server(),
		metadata.Server(),
	))

	generator := func(value string) string {
		return fmt.Sprintf("hello %s", value)
	}

	r.GET("/hellworld", func(ctx *gin.Context) {
		v := ctx.Query("name")
		ctx.JSON(200, gin.H{
			"message": generator(v),
		})
	})

	r.GET("/panic", func(ctx *gin.Context) {
		kgin.Error(ctx, errors.New(500, "want panic", "panic"))
	})

	r.GET("/nilerr", func(ctx *gin.Context) {
		kgin.Error(ctx, nil)
	})
	r.GET("/err", func(ctx *gin.Context) {
		ctx.Request.Header.Set("Accept", "application/xml")
		kgin.Error(ctx, errors.New(500, "want err xml", "json xml"))
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	addr := listener.Addr().String()

	httpSrv := http.NewServer(http.Listener(listener))
	httpSrv.HandlePrefix("/", r)

	app := kratos.New(
		kratos.Logger(log.NewStdLogger(io.Discard)),
		kratos.Name("gin-test"),
		kratos.Server(
			httpSrv,
		),
	)

	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		<-ticker.C

		client := resty.New()

		resp, err := client.NewRequest().Get(fmt.Sprintf("http://%s/hellworld?name=gin", addr))
		if err != nil {
			panic(err)
		}

		fmt.Println(resp.String())

		resp, err = client.NewRequest().Get(fmt.Sprintf("http://%s/panic", addr))
		if err != nil {
			panic(err)
		}

		fmt.Println(resp.String())

		resp, err = client.NewRequest().Get(fmt.Sprintf("http://%s/nilerr", addr))
		if err != nil {
			panic(err)
		}

		fmt.Println(resp.String())

		resp, err = client.NewRequest().Get(fmt.Sprintf("http://%s/err", addr))
		if err != nil {
			panic(err)
		}
		fmt.Println(resp.String())

		if err := app.Stop(); err != nil {
			fmt.Println(err)
		}
	}()

	if err := app.Run(); err != nil {
		fmt.Println(err)
	}

	// Output:
	// {"message":"hello gin"}
	// {"code":500, "reason":"want panic", "message":"panic", "metadata":{}}
}

func TestGinContext(t *testing.T) {
	ctx := &gin.Context{Keys: map[string]interface{}{
		"k": "v",
	}}

	newCtx := kgin.NewGinContext(context.Background(), ctx)

	if n, ok := kgin.FromGinContext(newCtx); ok {
		if n.Keys["k"] != "v" {
			t.Errorf("expected k=v, got %v", n.Keys["k"])
		}
	} else {
		t.Errorf("expected gin.Context, got %v", n)
	}
}

func TestContentType(t *testing.T) {
	type args struct {
		subtype string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{"json", args{"json"}, "application/json"},
		{"xml", args{"xml"}, "application/xml"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := kgin.ContentType(tt.args.subtype); got != tt.want {
				t.Errorf("ContentType() = %v, want %v", got, tt.want)
			}
		})
	}
}
