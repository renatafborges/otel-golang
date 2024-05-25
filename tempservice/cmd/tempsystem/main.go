package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/renatafborges/tempservice/internal/infra/web"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

const portNum string = ":9090"

var Host string
var URL_ZIPKIN string

func init() {
	Host = os.Getenv("URL_ZIPKIN")
	if Host == "" {
		Host = "localhost"
	}

	URL_ZIPKIN = "http://" + Host + ":9411/api/v2/spans"
}

func initTracer() func() {

	exporter, err := zipkin.New(
		URL_ZIPKIN,
	)
	if err != nil {
		log.Fatalf("failed to create Zipkin exporter: %v", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("service-B"),
		)),
	)

	otel.SetTracerProvider(tp)

	return func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatalf("failed to shutdown TracerProvider: %v", err)
		}
	}
}

func main() {

	log.Println("Starting http server.")
	initTracer()

	http.HandleFunc("/temperature/{zipcode}", web.HandleTemperature)

	err := http.ListenAndServe(portNum, otelhttp.NewHandler(http.DefaultServeMux, "http-server"))
	if err != nil {
		log.Fatal("error listen and serve", "error:", err)
	}
	log.Println("Started on port", portNum)
}
