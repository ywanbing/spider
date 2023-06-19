package xtrace

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

const (
	OTEL_TRACER_NAME = "spider"
)

// Init initializes the global TextMapPropagator.
func Init() {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}))
}
