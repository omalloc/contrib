package gin_test

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/go-resty/resty/v2"

	kgin "github.com/omalloc/contrib/kratos/gin"
)

func ExampleF_Middlewarese() {
	r := gin.New()
	r.Use(kgin.Middlewares(
		recovery.Recovery(),
		tracing.Server(),
		metadata.Server(),
	))

	r.GET("/hellworld", func(ctx *gin.Context) {
		v := ctx.Query("name")
		ctx.JSON(200, gin.H{
			"message": fmt.Sprintf("hello %s", v),
		})
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	addr := listener.Addr().String()
	// _, port, err := net.SplitHostPort(addr)
	// if err != nil {
	// 	panic(err)
	// }

	httpSrv := http.NewServer(http.Listener(listener))
	httpSrv.HandlePrefix("/", r)

	app := kratos.New(
		kratos.Name("gin-test"),
		kratos.Server(
			httpSrv,
		),
	)

	go func() {
		ticker := time.NewTicker(2 * time.Second)

		<-ticker.C
		_ = app.Stop()
	}()

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		<-ticker.C

		resp, err := resty.New().R().
			Get(fmt.Sprintf("http://%s/hellworld?name=gin", addr))
		if err != nil {
			panic(err)
		}

		fmt.Println(resp.String())
	}()

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}

	// Output:
	// {"message":"hello gin"}
}
