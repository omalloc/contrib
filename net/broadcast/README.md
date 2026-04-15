# UDP Broadcast Discovery

基于 UDP Broadcast 的局域网服务发现库。

协议约定：

- 客户端向固定 UDP 端口广播 JSON 查询：`{"type":"query","service":"service-x"}`
- 服务端收到后向请求来源单播 JSON 响应：`{"type":"response","service":"service-x","addr":"IP:PORT"}`

## 作为依赖使用

服务端：

```go
package main

import (
	"context"
	"log"

	broadcastserver "github.com/omalloc/contrib/net/broadcast/server"
)

func main() {
	if err := broadcastserver.ListenAndServe(context.Background(), broadcastserver.Config{
		Service:     "service-x",
		ServicePort: 8080,
		Meta: map[string]string{
			"version": "1.0.0",
		},
		Logger: log.Default(),
	}); err != nil {
		log.Fatal(err)
	}
}
```

客户端：

```go
package main

import (
	"context"
	"fmt"
	"log"

	broadcastclient "github.com/omalloc/contrib/net/broadcast/client"
)

func main() {
	results, err := broadcastclient.Discover(context.Background(), broadcastclient.Config{
		Service: "service-x",
		Logger:  log.Default(),
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, result := range results {
		fmt.Println(result.Addr, result.Meta)
	}
}
```
