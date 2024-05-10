package caching_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"

	"github.com/omalloc/contrib/kratos/caching"
)

func fakeRefresh() map[int64]string {
	key := time.Now().Unix()

	return map[int64]string{
		key - 1: "new-value1",
		key:     "new-value2",
	}
}

func TestBaseCache(t *testing.T) {

	//var c = 1

	cc := caching.New(
		caching.WithTracing[int64, string](otel.GetTracerProvider()),
		caching.WithSize[int64, string](100),
		caching.WithExpiration[int64, string](2*time.Second), // 每2秒刷新一次缓存
		caching.WithRefreshAfterWrite(fakeRefresh),
	)

	// 每秒取一次缓存
	ticker := time.NewTicker(time.Microsecond * 100)
	stop := time.After(5 * time.Second)

	for {
		select {
		case <-stop:
			t.Logf("stop")
			return
		case <-ticker.C:
			// do something
			kv := cc.GetALL(context.Background())
			for k, v := range kv {
				println("k: %d\t v: %s", k, v)
			}
			println("--------------")
			if len(kv) <= 0 {
				t.Logf("kv is empty")
			}
			if len(kv) != 2 {
				panic("kv is not two-size.")
			}
		}
	}
}

func TestCacheChanged(t *testing.T) {
	cc := caching.New(
		caching.WithTracing[int64, string](otel.GetTracerProvider()),
		caching.WithSize[int64, string](100),
		caching.WithExpiration[int64, string](2*time.Second), // 每2秒刷新一次缓存
		caching.WithRefreshAfterWrite(fakeRefresh),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Microsecond)
	defer cancel()

	// get all
	kvs := cc.GetALL(ctx)
	assert.Equal(t, 0, len(kvs))

	// set key 1
	cc.Set(ctx, 1, "value1")

	v, _ := cc.Get(ctx, 1)
	assert.Equal(t, "value1", v)

	// auto refresh
	time.Sleep(time.Second * 3)

	// get all
	// fakeRefresh = 2
	kvs = cc.GetALL(ctx)
	assert.Equal(t, 2, len(kvs))
}

func TestBlockIniting(t *testing.T) {
	cc := caching.New(
		caching.WithTracing[int64, string](otel.GetTracerProvider()),
		caching.WithSize[int64, string](100),
		caching.WithExpiration[int64, string](2*time.Second), // 每2秒刷新一次缓存
		caching.WithRefreshAfterWrite(fakeRefresh),
		caching.WithBlock[int64, string](),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Microsecond)
	defer cancel()

	// get all
	cc.GetALL(ctx)
	assert.Equal(t, 2, len(cc.GetALL(ctx)))
}

func TestGetValues(t *testing.T) {
	cc := caching.New(
		caching.WithSize[int64, string](100),
		caching.WithExpiration[int64, string](2*time.Second), // 每2秒刷新一次缓存
		caching.WithRefreshAfterWrite(fakeRefresh),
		caching.WithBlock[int64, string](),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Microsecond)
	defer cancel()

	t.Logf("get values %v", cc.Values(ctx))

	assert.Contains(t, cc.Values(ctx), "new-value1")
}

func TestAutoRefresh(t *testing.T) {
	cc := caching.New(
		caching.WithSize[int64, string](100),
		caching.WithExpiration[int64, string](2*time.Second), // 每2秒刷新一次缓存
		caching.WithRefreshAfterWrite(fakeRefresh),           // 每次请求刷出 2 个 kv
		caching.WithBlock[int64, string](),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Microsecond)
	defer cancel()

	assert.Equal(t, 2, len(cc.Values(ctx)))

	// 清空kv
	cc.Purge(ctx)
	assert.Equal(t, 0, len(cc.Values(ctx)))

	// 清空后缓存内部 2s 会自动刷新缓存 --> WithRefreshAfterWrite 的动作
	time.Sleep(time.Second * 3)

	assert.Equal(t, 2, len(cc.Values(ctx)))
	assert.Contains(t, cc.Values(ctx), "new-value1")
	assert.Contains(t, cc.Values(ctx), "new-value2")

	// 停掉刷新任务
	cc.Stop(ctx)
	// 再次清空kv
	cc.Purge(ctx)

	time.Sleep(time.Second * 3)
	assert.Equal(t, 0, len(cc.Values(ctx)))

	// 重启刷新任务
	cc.Restart(ctx)

	// 等待3秒 刷新后应该是2个kv
	time.Sleep(time.Second * 3)
	assert.Equal(t, 2, len(cc.Values(ctx)))
}
