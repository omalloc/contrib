package caching

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
)

type tracedWrapperLoadableCache[K comparable, V any] struct {
	*loadableCache[K, V]
}

// func traceWrapper[F any, R any](tracer trace.Tracer, f func(context.Context, F) R) func(context.Context, F) R {
// 	return func(parentCtx context.Context, arg F) R {
// 		ctx, span := tracer.Start(parentCtx, "loadableCache")
// 		defer span.End()

// 		span.SetAttributes(
// 			attribute.String("arg", fmt.Sprintf("%v", arg)),
// 		)

// 		return f(ctx, arg)
// 	}
// }

func (r *tracedWrapperLoadableCache[K, V]) Get(parentCtx context.Context, k K) (V, error) {
	ctx, span := r.tracer.Start(parentCtx, "loadableCache")
	defer span.End()

	attrs := make([]attribute.KeyValue, 0, 6)

	v, err := r.loadableCache.Get(ctx, k)
	if err != nil {
		attrs = append(attrs, attribute.String("cache.error", fmt.Sprintf("%v", err)))
	} else {
		attrs = append(attrs, attribute.String("cache.key_hit", fmt.Sprintf("%v", k)))
	}
	attrs = append(attrs, attribute.String("cache.key", fmt.Sprintf("%v", k)))

	span.SetAttributes(
		attrs...,
	)

	return v, err
}

func (r *tracedWrapperLoadableCache[K, V]) GetALL(parentCtx context.Context) map[K]V {
	ctx, span := r.tracer.Start(parentCtx, "loadableCache")
	defer span.End()

	ret := r.loadableCache.GetALL(ctx)
	// 为了避免在链路追踪中出现大量的 key-value, 这里只记录 key 的数量
	span.SetAttributes(
		attribute.Int("cache.key_size", len(ret)),
	)

	return ret
}

func (r *tracedWrapperLoadableCache[K, V]) Values(parentCtx context.Context) []V {
	ctx, span := r.tracer.Start(parentCtx, "loadableCache")
	defer span.End()

	ret := r.loadableCache.Values(ctx)

	span.SetAttributes(
		attribute.Int("cache.value_size", len(ret)),
	)

	return ret
}

func (r *tracedWrapperLoadableCache[K, V]) Set(parentCtx context.Context, k K, v V) error {
	ctx, span := r.tracer.Start(parentCtx, "loadableCache")
	defer span.End()

	err := r.loadableCache.Set(ctx, k, v)
	if err != nil {
		span.SetAttributes(
			attribute.String("cache.error", fmt.Sprintf("%v", err)),
		)
	} else {
		span.SetAttributes(
			attribute.String("cache.key", fmt.Sprintf("%v", k)),
		)
	}
	return err
}

func (r *tracedWrapperLoadableCache[K, V]) Purge(parentCtx context.Context) {
	ctx, span := r.tracer.Start(parentCtx, "loadableCache")
	defer span.End()

	r.loadableCache.Purge(ctx)
}

func (r *tracedWrapperLoadableCache[K, V]) TryPurgeAndReload(parentCtx context.Context) bool {
	ctx, span := r.tracer.Start(parentCtx, "loadableCache")
	defer span.End()

	ok := r.loadableCache.TryPurgeAndReload(ctx)

	span.SetAttributes(
		attribute.Bool("cache.has_reload_success", ok),
	)
	return ok
}

func (r *tracedWrapperLoadableCache[K, V]) Stop(ctx context.Context) {
	r.loadableCache.Stop(ctx)
}

func (r *tracedWrapperLoadableCache[K, V]) Restart(ctx context.Context) {
	r.loadableCache.Restart(ctx)
}
