package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/omalloc/contrib/net/broadcast"
)

const defaultBufferSize = 2048

// Config controls how the discovery server listens and what it advertises.
type Config struct {
	Service       string
	DiscoveryPort int
	ServicePort   int
	ServiceHost   string
	Meta          map[string]string
	Logger        *log.Logger
	BufferSize    int
}

// ListenAndServe starts a UDP discovery server and blocks until the context is canceled.
func ListenAndServe(ctx context.Context, cfg Config) error {
	if err := cfg.validate(); err != nil {
		return err
	}

	listenAddr := &net.UDPAddr{IP: net.IPv4zero, Port: cfg.discoveryPort()}
	conn, err := net.ListenUDP("udp4", listenAddr)
	if err != nil {
		return fmt.Errorf("listen discovery socket: %w", err)
	}
	defer conn.Close()

	cfg.logf("udp discovery server listening on %s for service=%s service-port=%d", conn.LocalAddr(), cfg.Service, cfg.servicePort())

	buf := make([]byte, cfg.bufferSize())
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := conn.SetReadDeadline(readDeadline(context.Background())); err != nil {
				return fmt.Errorf("set read deadline: %w", err)
			}

			n, clientAddr, err := conn.ReadFromUDP(buf)
			if err != nil {
				if ctx.Err() != nil {
					return nil
				}

				netErr, ok := err.(net.Error)
				if ok && netErr.Timeout() {
					continue
				}

				cfg.logf("read discovery packet: %v", err)
				continue
			}

			msg, err := broadcast.DecodeMessage(buf[:n])
			if err != nil {
				cfg.logf("ignore malformed discovery packet from %s: %v", clientAddr, err)
				continue
			}

			if msg.Type != broadcast.MessageTypeQuery || msg.Service != cfg.Service {
				continue
			}

			advertiseIP, err := advertisedIP(cfg.ServiceHost, clientAddr)
			if err != nil {
				cfg.logf("resolve advertised IP for %s: %v", clientAddr, err)
				continue
			}

			resp := broadcast.Message{
				Type:    broadcast.MessageTypeResponse,
				Service: cfg.Service,
				Addr:    broadcast.AddrWithPort(advertiseIP, cfg.servicePort()),
				Meta:    cloneMeta(cfg.Meta),
			}

			payload, err := broadcast.EncodeMessage(resp)
			if err != nil {
				cfg.logf("encode discovery response: %v", err)
				continue
			}

			if _, err := conn.WriteToUDP(payload, clientAddr); err != nil {
				cfg.logf("reply to %s: %v", clientAddr, err)
				continue
			}

			cfg.logf("discovery hit from %s -> %s", clientAddr, resp.Addr)
		}
	}
}

func (c Config) validate() error {
	if c.Service == "" {
		return fmt.Errorf("service is required")
	}
	if c.servicePort() < 1 || c.servicePort() > 65535 {
		return fmt.Errorf("service port must be between 1 and 65535")
	}
	if c.discoveryPort() < 1 || c.discoveryPort() > 65535 {
		return fmt.Errorf("discovery port must be between 1 and 65535")
	}
	if c.ServiceHost == "" {
		return nil
	}
	if ip := net.ParseIP(c.ServiceHost); ip == nil || ip.To4() == nil {
		return fmt.Errorf("service host must be a valid IPv4 address")
	}
	return nil
}

func (c Config) discoveryPort() int {
	if c.DiscoveryPort == 0 {
		return broadcast.DefaultDiscoveryPort
	}
	return c.DiscoveryPort
}

func (c Config) servicePort() int {
	if c.ServicePort == 0 {
		return 8080
	}
	return c.ServicePort
}

func (c Config) bufferSize() int {
	if c.BufferSize <= 0 {
		return defaultBufferSize
	}
	return c.BufferSize
}

func (c Config) logf(format string, args ...any) {
	if c.Logger == nil {
		return
	}
	c.Logger.Printf(format, args...)
}

func advertisedIP(serviceHost string, clientAddr *net.UDPAddr) (net.IP, error) {
	if serviceHost != "" {
		ip := net.ParseIP(serviceHost)
		if ip == nil {
			return nil, &net.ParseError{Type: "IP address", Text: serviceHost}
		}
		ipv4 := ip.To4()
		if ipv4 == nil {
			return nil, &net.ParseError{Type: "IPv4 address", Text: serviceHost}
		}
		return ipv4, nil
	}

	return broadcast.IPv4ForPeer(broadcast.ParseIPv4(clientAddr))
}

func readDeadline(ctx context.Context) time.Time {
	if ctx.Err() != nil {
		return time.Now()
	}
	return time.Now().Add(time.Second)
}

func cloneMeta(meta map[string]string) map[string]string {
	if len(meta) == 0 {
		return nil
	}
	result := make(map[string]string, len(meta))
	for key, value := range meta {
		result[key] = value
	}
	return result
}
