package broadcast

import (
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"syscall"
)

const DefaultDiscoveryPort = 5353

const (
	MessageTypeQuery    = "query"
	MessageTypeResponse = "response"
)

type Message struct {
	Type    string            `json:"type"`
	Service string            `json:"service"`
	Addr    string            `json:"addr,omitempty"`
	Meta    map[string]string `json:"meta,omitempty"`
}

func EncodeMessage(msg Message) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("marshal discovery message: %w", err)
	}
	return data, nil
}

func DecodeMessage(data []byte) (Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return Message{}, fmt.Errorf("decode discovery message: %w", err)
	}
	return msg, nil
}

func EnableBroadcast(conn *net.UDPConn) error {
	rawConn, err := conn.SyscallConn()
	if err != nil {
		return fmt.Errorf("get raw udp connection: %w", err)
	}

	var sockErr error
	if err := rawConn.Control(func(fd uintptr) {
		sockErr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
	}); err != nil {
		return fmt.Errorf("enable broadcast control: %w", err)
	}
	if sockErr != nil {
		return fmt.Errorf("enable SO_BROADCAST: %w", sockErr)
	}

	return nil
}

func SubnetBroadcastTargets(port int) ([]*net.UDPAddr, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("list interfaces: %w", err)
	}

	targets := make(map[string]*net.UDPAddr)
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagBroadcast == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP == nil {
				continue
			}

			ip := ipNet.IP.To4()
			mask := net.IP(ipNet.Mask).To4()
			if ip == nil || mask == nil {
				continue
			}

			broadcastIP := make(net.IP, net.IPv4len)
			for i := 0; i < net.IPv4len; i++ {
				broadcastIP[i] = ip[i] | ^mask[i]
			}

			udpAddr := &net.UDPAddr{IP: broadcastIP, Port: port}
			targets[udpAddr.String()] = udpAddr
		}
	}

	if len(targets) == 0 {
		fallback := &net.UDPAddr{IP: net.IPv4bcast, Port: port}
		return []*net.UDPAddr{fallback}, nil
	}

	keys := make([]string, 0, len(targets))
	for key := range targets {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make([]*net.UDPAddr, 0, len(keys))
	for _, key := range keys {
		result = append(result, targets[key])
	}

	return result, nil
}

func IPv4ForPeer(peer net.IP) (net.IP, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("list interfaces: %w", err)
	}

	peerV4 := peer.To4()
	var fallback net.IP

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP == nil {
				continue
			}

			ip := ipNet.IP.To4()
			if ip == nil {
				continue
			}

			if fallback == nil {
				fallback = append(net.IP(nil), ip...)
			}

			if peerV4 != nil && ipNet.Contains(peerV4) {
				return append(net.IP(nil), ip...), nil
			}
		}
	}

	if fallback != nil {
		return fallback, nil
	}

	return nil, fmt.Errorf("no non-loopback IPv4 address found")
}

func AddrWithPort(ip net.IP, port int) string {
	return net.JoinHostPort(ip.String(), fmt.Sprintf("%d", port))
}

func ParseIPv4(addr *net.UDPAddr) net.IP {
	if addr == nil {
		return nil
	}
	return addr.IP.To4()
}
