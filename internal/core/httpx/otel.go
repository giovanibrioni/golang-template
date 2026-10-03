package httpx

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/propagation"
)

// OTelMiddleware wraps the handler with OpenTelemetry HTTP instrumentation.
func OTelMiddleware(next http.Handler) http.Handler {
	return otelhttp.NewHandler(next, "http",
		otelhttp.WithPropagators(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)),
	)
}
