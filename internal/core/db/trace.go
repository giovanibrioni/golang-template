package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type queryTracer struct{}

func (queryTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, _ pgx.TraceQueryStartData) context.Context {
	attrs := []attribute.KeyValue{attribute.String("db.system", "postgresql")}
	if conn != nil && conn.Config() != nil {
		if name := conn.Config().Database; name != "" {
			attrs = append(attrs, attribute.String("db.name", name))
		}
		if host := conn.Config().Host; host != "" {
			attrs = append(attrs, attribute.String("server.address", host))
		}
	}
	ctx, _ = otel.Tracer("db").Start(ctx, "postgresql.query",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attrs...),
	)
	return ctx
}

func (queryTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	span := trace.SpanFromContext(ctx)
	if data.Err != nil {
		span.RecordError(data.Err)
		span.SetStatus(codes.Error, data.Err.Error())
	}
	span.End()
}
