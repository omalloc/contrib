package machineid

import (
	"crypto/sha1"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/samber/lo"
	"github.com/shirou/gopsutil/v4/cpu"
)

type MachineCPUInfo struct {
	VendorID string
	ModeName string
	Cores    uint32
	CpuMHz   float64
}

type MachineInfo struct {
	arch                string
	archMark            string // AMD64 = 22, AARCH64 = 24, UNKNOWN = 20
	cores               uint32 // cpu cores
	vendorID            string
	cpuMHz              float64
	modelName           string
	productUUID         string
	defaultGatewayIface string
	defaultGatewayMAC   string
}

func Load() MachineInfo {
	m := MachineInfo{}

	cpus := getCPUs()
	m.arch = runtime.GOARCH
	m.archMark = archMark(m.arch)
	m.cores = cpus.Cores
	m.vendorID = cpus.VendorID
	m.cpuMHz = cpus.CpuMHz
	m.modelName = cpus.ModeName
	m.productUUID = getProductUUID()
	m.defaultGatewayIface = getDefaultRoute()
	m.defaultGatewayMAC = getIfaceMAC(m.defaultGatewayIface)
	return m
}

func (m MachineInfo) SN() string {
	sb := strings.Builder{}

	sb.WriteString(m.arch)              // CPU ARCH
	sb.WriteString(m.modelName)         // CPU名称
	sb.WriteRune(rune(m.cores))         // 核心数
	sb.WriteString(m.productUUID)       // 主板序列号
	sb.WriteString(m.defaultGatewayMAC) // 默认路由的MAC地址

	h := sha1.New()
	h.Write([]byte(sb.String()))
	hash := h.Sum(nil)

	mark := m.archMark
	digit := hashToFixedDigits(hash, 26)
	return fmt.Sprintf("%s00%s", mark, digit)
}

func getCPUs() MachineCPUInfo {
	m := MachineCPUInfo{}

	infos, err := cpu.Info()
	if err != nil {
		return m
	}

	cores := len(infos)
	m.Cores = uint32(cores)
	m.VendorID = infos[0].VendorID
	m.ModeName = infos[0].ModelName
	m.CpuMHz = infos[0].Mhz
	return m
}

// Get /sys/class/dmi/id/product_uuid
func getProductUUID() string {
	data, err := os.ReadFile(filepath.Clean("/sys/class/dmi/id/product_uuid"))
	if err != nil {
		// backport
		return getMachineID()
	}
	return string(data)
}

// Get /etc/machine-id
func getMachineID() string {
	data, err := os.ReadFile(filepath.Clean("/etc/machine-id"))
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(data))
}

func getDefaultRoute() string {
	data, err := os.ReadFile(filepath.Clean("/proc/net/route"))
	if err != nil {
		return ""
	}

	routes := make([]lo.Tuple2[string, string], 0)

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		routes = append(routes, lo.Tuple2[string, string]{
			A: fields[0],
			B: fields[1],
		})
	}

	defaultRoute := ""
	for _, route := range routes {
		if route.B == "00000000" {
			defaultRoute = route.A
			break
		}
	}
	return defaultRoute
}

func getIfaceMAC(ifName string) string {
	iface, err := net.InterfaceByName(ifName)
	if err != nil {
		return ""
	}

	return iface.HardwareAddr.String()
}

func archMark(arch string) string {
	switch strings.ToUpper(arch) {
	case "AMD64", "X86_64":
		return "22"
	case "ARM64", "AARCH64":
		return "24"
	default:
		return "20"
	}
}

func hashToFixedDigits(data []byte, digits int) string {
	h := sha1.Sum(data)

	n := new(big.Int).SetBytes(h[:])

	mod := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits)), nil)

	n.Mod(n, mod)

	return fmt.Sprintf("%0*d", digits, n)
}
