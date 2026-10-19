package db

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestQuerySpanMarksPostgres(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)
	t.Cleanup(func() { otel.SetTracerProvider(previous) })

	ctx := queryTracer{}.TraceQueryStart(context.Background(), nil, pgx.TraceQueryStartData{SQL: "select 1"})
	queryTracer{}.TraceQueryEnd(ctx, nil, pgx.TraceQueryEndData{Err: errors.New("db down")})

	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("spans = %d", len(spans))
	}
	system := spans[0].Attributes()
	found := false
	for _, attr := range system {
		if string(attr.Key) == "db.system" && attr.Value.AsString() == "postgresql" {
			found = true
		}
	}
	if !found {
		t.Fatalf("db.system missing: %v", spans[0].Attributes())
	}
}
