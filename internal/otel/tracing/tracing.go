package tracing

import (
	"context"
	"log"
	"os"

	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

var logger = log.New(os.Stderr, "zipkin", log.Ldate|log.Ltime|log.Llongfile)

func Start() (trace.Tracer, func(context.Context) error) {
	url := os.Getenv("ZIPKIN_URL")
	exp, err := newExporter(url)
	if err != nil {
		log.Fatalln(err)
	}
	tp := newTraceProvider(exp)
	return tp.Tracer(os.Getenv("SERVICE_NAME")), tp.Shutdown
}

func newExporter(url string) (*zipkin.Exporter, error) {
	return zipkin.New(url, zipkin.WithLogger(logger))
}

func newTraceProvider(exp sdktrace.SpanExporter) *sdktrace.TracerProvider {
	// Ensure default SDK resources and the required service name are set.
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(os.Getenv("SERVICE_NAME")),
		),
	)

	if err != nil {
		log.Fatalln(err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(r),
	)
}
