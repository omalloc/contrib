package filter

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/selector"
)

const HangKey = "hang"
const HangValue = "true"

func HasHang(metadata map[string]string) bool {
	if value, ok := metadata[HangKey]; ok && HangValue == strings.TrimSpace(value) {
		return true
	}
	return false
}

// HangState is hang state filter.
func HangState() selector.NodeFilter {
	return func(_ context.Context, nodes []selector.Node) []selector.Node {
		newNodes := make([]selector.Node, 0)
		for _, n := range nodes {
			if HasHang(n.Metadata()) {
				continue
			}
			newNodes = append(newNodes, n)
		}
		return newNodes
	}
}
