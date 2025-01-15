package resty_test

import (
	"fmt"
	"time"

	"github.com/omalloc/contrib/kratos/resty"
)

type TestAnythingBody struct {
	Args struct {
		Example string `json:"example,omitempty"`
	} `json:"args"`
	Headers map[string]string `json:"headers"`
}

func ExampleNew() {
	client := resty.New().SetTimeout(time.Second * 10)

	var body TestAnythingBody
	resp, err := client.NewRequest().SetResult(&body).Get("https://httpbin.org/anything?example=new")
	if err != nil {
		panic(err)
	}

	// Output: 200-new
	fmt.Printf("%d-%s", resp.StatusCode(), body.Args.Example)
}
