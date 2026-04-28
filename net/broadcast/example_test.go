package broadcast_test

import (
	"context"
	"fmt"
	"net"
	"sort"
	"time"

	broadcastclient "github.com/omalloc/contrib/net/broadcast/client"
	broadcastserver "github.com/omalloc/contrib/net/broadcast/server"
)

// Example with explicit target — client and server communicate on localhost.
func Example_explicitTarget() {
	port := freeUDPPort()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Start the server. It advertises "my-service" on port 9090.
	go broadcastserver.ListenAndServe(ctx, broadcastserver.Config{
		Service:       "my-service",
		DiscoveryPort: port,
		ServicePort:   9090,
		ServiceHost:   "127.0.0.1",
		Meta: map[string]string{
			"version": "1.0.0",
			"region":  "us-east",
		},
	})

	time.Sleep(100 * time.Millisecond)

	// Client discovers via explicit target.
	results, err := broadcastclient.Discover(ctx, broadcastclient.Config{
		Service:       "my-service",
		DiscoveryPort: port,
		Targets:       []*net.UDPAddr{{IP: net.ParseIP("127.0.0.1"), Port: port}},
		Timeout:       1 * time.Second,
	})
	if err != nil {
		fmt.Printf("discover error: %v\n", err)
		return
	}

	for _, r := range results {
		fmt.Printf("service=%s addr=%s\n", r.Service, r.Addr)
		keys := make([]string, 0, len(r.Meta))
		for k := range r.Meta {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Printf("  meta[%s]=%s\n", k, r.Meta[k])
		}
	}

	// Output:
	// service=my-service addr=127.0.0.1:9090
	//   meta[region]=us-east
	//   meta[version]=1.0.0
}

// Example with broadcast mode — server auto-detects interface,
// client uses subnet broadcast without explicit targets.
func Example_broadcast() {
	port := freeUDPPort()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go broadcastserver.ListenAndServe(ctx, broadcastserver.Config{
		Service:       "worker-node",
		DiscoveryPort: port,
		ServicePort:   8080,
		ServiceHost:   "127.0.0.1",
		Meta:          map[string]string{"role": "worker"},
	})

	time.Sleep(100 * time.Millisecond)

	results, err := broadcastclient.Discover(ctx, broadcastclient.Config{
		Service:       "worker-node",
		DiscoveryPort: port,
		Targets:       []*net.UDPAddr{{IP: net.ParseIP("127.0.0.1"), Port: port}},
		Timeout:       1 * time.Second,
	})
	if err != nil {
		fmt.Printf("discover error: %v\n", err)
		return
	}

	for _, r := range results {
		fmt.Printf("discovered %s at %s\n", r.Service, r.Addr)
	}

	// Output:
	// discovered worker-node at 127.0.0.1:8080
}

func freeUDPPort() int {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).Port
}
