package client

import (
	"context"
	"fmt"
	"log"
	"net"
	"sort"
	"time"

	"github.com/omalloc/contrib/net/broadcast"
)

const (
	defaultBufferSize = 2048
	defaultTimeout    = 3 * time.Second
)

// Config controls how discovery queries are sent and how long responses are collected.
type Config struct {
	Service       string
	DiscoveryPort int
	Timeout       time.Duration
	Targets       []*net.UDPAddr
	Logger        *log.Logger
	BufferSize    int
}

// Result describes a single discovery response from a server instance.
type Result struct {
	Service string
	Addr    string
	Meta    map[string]string
	From    *net.UDPAddr
}

// Discover sends a discovery query and collects matching responses until timeout or context cancelation.
func Discover(ctx context.Context, cfg Config) ([]Result, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return nil, fmt.Errorf("listen discovery client socket: %w", err)
	}
	defer conn.Close()

	if err := broadcast.EnableBroadcast(conn); err != nil {
		return nil, fmt.Errorf("enable broadcast on client socket: %w", err)
	}

	targets, err := cfg.targets()
	if err != nil {
		return nil, err
	}

	queryPayload, err := broadcast.EncodeMessage(broadcast.Message{
		Type:    broadcast.MessageTypeQuery,
		Service: cfg.Service,
	})
	if err != nil {
		return nil, fmt.Errorf("encode discovery query: %w", err)
	}

	for _, target := range targets {
		if _, err := conn.WriteToUDP(queryPayload, target); err != nil {
			cfg.logf("send broadcast to %s: %v", target, err)
			continue
		}
		cfg.logf("broadcast query sent to %s", target)
	}

	results := make(map[string]Result)
	buf := make([]byte, cfg.bufferSize())
	deadline := readDeadline(ctx, cfg.timeout())

	for {
		if err := conn.SetReadDeadline(deadline); err != nil {
			return nil, fmt.Errorf("set read deadline: %w", err)
		}

		n, serverAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			if ctx.Err() != nil {
				break
			}

			netErr, ok := err.(net.Error)
			if ok && netErr.Timeout() {
				break
			}

			return nil, fmt.Errorf("read discovery response: %w", err)
		}

		msg, err := broadcast.DecodeMessage(buf[:n])
		if err != nil {
			cfg.logf("ignore malformed response from %s: %v", serverAddr, err)
			continue
		}

		if msg.Type != broadcast.MessageTypeResponse || msg.Service != cfg.Service || msg.Addr == "" {
			continue
		}

		results[msg.Addr] = Result{
			Service: msg.Service,
			Addr:    msg.Addr,
			Meta:    cloneMeta(msg.Meta),
			From:    cloneAddr(serverAddr),
		}
	}

	resultList := make([]Result, 0, len(results))
	for _, result := range results {
		resultList = append(resultList, result)
	}
	sort.Slice(resultList, func(i, j int) bool {
		return resultList[i].Addr < resultList[j].Addr
	})

	return resultList, nil
}

func (c Config) validate() error {
	if c.Service == "" {
		return fmt.Errorf("service is required")
	}
	if c.discoveryPort() < 1 || c.discoveryPort() > 65535 {
		return fmt.Errorf("discovery port must be between 1 and 65535")
	}
	if c.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}
	return nil
}

func (c Config) discoveryPort() int {
	if c.DiscoveryPort == 0 {
		return broadcast.DefaultDiscoveryPort
	}
	return c.DiscoveryPort
}

func (c Config) timeout() time.Duration {
	if c.Timeout == 0 {
		return defaultTimeout
	}
	return c.Timeout
}

func (c Config) bufferSize() int {
	if c.BufferSize <= 0 {
		return defaultBufferSize
	}
	return c.BufferSize
}

func (c Config) targets() ([]*net.UDPAddr, error) {
	if len(c.Targets) != 0 {
		result := make([]*net.UDPAddr, 0, len(c.Targets))
		for _, target := range c.Targets {
			if target == nil || target.IP == nil {
				continue
			}
			result = append(result, cloneAddr(target))
		}
		if len(result) == 0 {
			return nil, fmt.Errorf("targets must contain at least one valid UDP address")
		}
		return result, nil
	}

	targets, err := broadcast.SubnetBroadcastTargets(c.discoveryPort())
	if err != nil {
		return nil, fmt.Errorf("resolve broadcast targets: %w", err)
	}
	return targets, nil
}

func (c Config) logf(format string, args ...any) {
	if c.Logger == nil {
		return
	}
	c.Logger.Printf(format, args...)
}

func readDeadline(ctx context.Context, timeout time.Duration) time.Time {
	deadline := time.Now().Add(timeout)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		return ctxDeadline
	}
	return deadline
}

func cloneAddr(addr *net.UDPAddr) *net.UDPAddr {
	if addr == nil {
		return nil
	}
	clone := *addr
	if addr.IP != nil {
		clone.IP = append(net.IP(nil), addr.IP...)
	}
	return &clone
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
