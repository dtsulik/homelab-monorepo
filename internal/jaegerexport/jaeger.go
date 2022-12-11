package jaegerexport

import (
	"context"

	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.6.1"
)

func JaegerTraceProvider(name, url string) (*tracesdk.TracerProvider, error) {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}
	resources, err := resource.New(context.Background(),
		resource.WithFromEnv(),
		resource.WithProcess(),
		semconv.ServiceNameKey.String(name))
	if err != nil {
		return nil, err
	}

	// TODO need to pass our logger somewhere.
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resources),
	)
	return tp, nil
}
