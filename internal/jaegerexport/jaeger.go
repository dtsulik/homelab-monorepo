package jaegerexport

import (
	"context"

	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

func JaegerTraceProvider(url string) (*tracesdk.TracerProvider, error) {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}
	resources, err := resource.New(context.Background(),
		resource.WithFromEnv(),
		resource.WithProcess())
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
