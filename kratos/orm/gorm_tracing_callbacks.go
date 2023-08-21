package orm

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

type contextTraceKey string

const (
	dbTableKey        = attribute.Key("db.sql.table")
	dbRowsAffectedKey = attribute.Key("db.rows_affected")
	dbOperationKey    = semconv.DBOperationKey
	dbStatementKey    = semconv.DBStatementKey
	omitVarsKey       = contextTraceKey("omit_vars")
)

const (
	eventMaxSize = 500
	maxChunks    = 4
)

func dbTable(name string) attribute.KeyValue {
	return dbTableKey.String(name)
}

func dbStatement(stmt string) attribute.KeyValue {
	return dbStatementKey.String(stmt)
}

func dbCount(n int64) attribute.KeyValue {
	return dbRowsAffectedKey.Int64(n)
}

func dbOperation(op string) attribute.KeyValue {
	return dbOperationKey.String(op)
}

func (op *GormOpenTelemetryPlugin) spanName(tx *gorm.DB, operation string) string {
	query := op.extractQuery(tx)

	operation = operationForQuery(query, operation)

	target := op.c.dbName
	if target == "" {
		target = tx.Dialector.Name()
	}

	if tx.Statement != nil && tx.Statement.Table != "" {
		target += "." + tx.Statement.Table
	}

	return fmt.Sprintf("%s %s", operation, target)
}

func operationForQuery(query, op string) string {
	if op != "" {
		return op
	}

	return strings.ToUpper(strings.Split(query, " ")[0])
}

func (op *GormOpenTelemetryPlugin) before(operation string) traceHookFunc {
	return func(tx *gorm.DB) {
		// skip the reporting if not recording
		if !tx.Statement.SkipHooks {
			tx.Statement.Context, _ = op.tracer.
				Start(tx.Statement.Context, op.spanName(tx, operation), trace.WithSpanKind(trace.SpanKindClient))
		}
	}
}

func (op *GormOpenTelemetryPlugin) extractQuery(tx *gorm.DB) string {
	shouldOmit, ok := tx.Statement.Context.Value(omitVarsKey).(bool)
	if !ok {
		shouldOmit = op.c.alwaysOmitVars
	}

	if shouldOmit {
		return tx.Statement.SQL.String()
	}
	return tx.Dialector.Explain(tx.Statement.SQL.String(), tx.Statement.Vars...)
}

func chunkBy(val string, size int, callback func(string, ...trace.EventOption)) {
	if len(val) > maxChunks*size {
		return
	}

	for i := 0; i < maxChunks*size; i += size {
		end := len(val)
		if end > size {
			end = size
		}
		callback(val[0:end])
		if end > len(val)-1 {
			break
		}
		val = val[end:]
	}
}

func (op *GormOpenTelemetryPlugin) after(operation string) traceHookFunc {
	return func(tx *gorm.DB) {
		// skip the reporting if not recording
		if tx.Statement.SkipHooks {
			return
		}

		span := trace.SpanFromContext(tx.Statement.Context)
		if !span.IsRecording() {
			// skip the reporting if not recording
			return
		}
		defer span.End()

		span.SetName(op.spanName(tx, operation))
		// Error
		if tx.Error != nil {
			span.SetStatus(codes.Error, tx.Error.Error())
		}

		// extract the db operation
		query := strings.ToValidUTF8(op.extractQuery(tx), "")

		// If query is longer then max size log it as chunked event, otherwise log it in attribute
		if len(query) > eventMaxSize {
			chunkBy(query, eventMaxSize, span.AddEvent)
		} else {
			span.SetAttributes(dbStatement(query))
		}

		operation = operationForQuery(query, operation)
		if tx.Statement.Table != "" {
			span.SetAttributes(dbTable(tx.Statement.Table))
		}

		span.SetAttributes(
			dbOperation(operation),
			dbCount(tx.Statement.RowsAffected),
		)
	}
}

func WithOmitVariablesFromTrace(ctx context.Context) context.Context {
	return context.WithValue(ctx, omitVarsKey, true)
}
