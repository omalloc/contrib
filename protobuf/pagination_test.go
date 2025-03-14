package protobuf_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/omalloc/contrib/protobuf"
)

func TestPaginationWrap(t *testing.T) {
	p := protobuf.PageWrap(nil)
	assert.Equal(t, p.Current, int32(1))
	assert.Equal(t, p.PageSize, int32(20))

	assert.Equal(t, p.Limit(), 20)
	assert.Equal(t, p.Offset(), 0)

	p = protobuf.PageWrap(&protobuf.Pagination{Current: 0, PageSize: 100})
	assert.Equal(t, p.Current, int32(1))
	assert.Equal(t, p.PageSize, int32(100))

	assert.Equal(t, p.Limit(), 100)
	assert.Equal(t, p.Offset(), 0)
}
