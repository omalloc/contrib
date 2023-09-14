package filter_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/selector"

	"github.com/omalloc/contrib/kratos/selector/filter"
)

func TestHangState(t *testing.T) {
	tests := []struct {
		name string
		got  []selector.Node
		want []selector.Node
	}{
		{
			name: "TestHangState",
			got: []selector.Node{
				selector.NewNode("http", "127.0.0.1:8000", &registry.ServiceInstance{
					Metadata: map[string]string{
						"hang": "true",
					},
				}),
				selector.NewNode("http", "127.0.0.1:8001", &registry.ServiceInstance{
					Metadata: map[string]string{
						"hang": "false",
					},
				}),
				selector.NewNode("http", "127.0.0.1:8002", &registry.ServiceInstance{
					Metadata: map[string]string{},
				}),
			},
			want: []selector.Node{
				selector.NewNode("http", "127.0.0.1:8001", &registry.ServiceInstance{
					Metadata: map[string]string{
						"hang": "false",
					},
				}),
				selector.NewNode("http", "127.0.0.1:8002", &registry.ServiceInstance{
					Metadata: map[string]string{},
				}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := filter.HangState()

			if got := filter(context.Background(), tt.got); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HangState() = %v, want %v", got, tt.want)
			}
		})
	}
}
