package client_test

import (
	"context"
	"net"
	"testing"
	"time"

	broadcastclient "github.com/omalloc/contrib/net/broadcast/client"
	broadcastserver "github.com/omalloc/contrib/net/broadcast/server"
)

func TestDiscoverWithExplicitTarget(t *testing.T) {
	t.Parallel()

	discoveryPort := freeUDPPort(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- broadcastserver.ListenAndServe(ctx, broadcastserver.Config{
			Service:       "service-x",
			DiscoveryPort: discoveryPort,
			ServicePort:   8428,
			ServiceHost:   "127.0.0.1",
			Meta: map[string]string{
				"version": "test",
			},
		})
	}()

	target := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: discoveryPort}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		results, err := broadcastclient.Discover(context.Background(), broadcastclient.Config{
			Service:       "service-x",
			DiscoveryPort: discoveryPort,
			Targets:       []*net.UDPAddr{target},
			Timeout:       150 * time.Millisecond,
		})
		if err != nil {
			t.Fatalf("discover: %v", err)
		}

		if len(results) == 1 {
			if results[0].Addr != "127.0.0.1:8428" {
				t.Fatalf("unexpected address: %s", results[0].Addr)
			}
			if results[0].Meta["version"] != "test" {
				t.Fatalf("unexpected meta: %#v", results[0].Meta)
			}
			cancel()
			select {
			case err := <-errCh:
				if err != nil {
					t.Fatalf("server exited with error: %v", err)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("server did not stop after cancel")
			}
			return
		}

		select {
		case err := <-errCh:
			if err != nil {
				t.Fatalf("server exited early: %v", err)
			}
			t.Fatal("server exited before discovery succeeded")
		default:
		}

		time.Sleep(20 * time.Millisecond)
	}

	cancel()
	t.Fatal("did not discover server before timeout")
}

func freeUDPPort(t *testing.T) int {
	t.Helper()

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		t.Fatalf("listen udp: %v", err)
	}
	defer conn.Close()

	addr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		t.Fatalf("unexpected local addr type: %T", conn.LocalAddr())
	}

	return addr.Port
}
