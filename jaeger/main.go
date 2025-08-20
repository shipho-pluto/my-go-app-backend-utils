package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

func initTracer() (*sdktrace.TracerProvider, error) {
	// Создаем Jaeger экспортер
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://localhost:14268/api/traces")))
	if err != nil {
		return nil, err
	}

	// Создаем TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("my-go-app"),
		)),
	)

	// Устанавливаем глобальный tracer
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}))

	tracer = tp.Tracer("my-go-app")
	return tp, nil
}

func main() {
	// Инициализируем tracer
	tp, err := initTracer()
	if err != nil {
		log.Fatal("Failed to initialize tracer:", err)
	}
	defer tp.Shutdown(context.Background())

	fmt.Println("ok Tracer initialized! Sending traces to Jaeger...")
	fmt.Println("📊 Jaeger UI: http://localhost:16686")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем контекст из заголовков
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		// Создаем span
		_, span := tracer.Start(ctx, "HTTP_GET /")
		defer span.End()

		// Имитируем работу
		time.Sleep(50 * time.Millisecond)

		// Добавляем атрибуты в span
		span.SetAttributes(
			semconv.HTTPMethodKey.String(r.Method),
			semconv.HTTPRouteKey.String("/"),
		)

		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("🚀 Hello with Jaeger tracing!"))

		log.Printf("Request processed - TraceID: %s", span.SpanContext().TraceID())
	})

	http.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "HTTP_GET /api/data")
		defer span.End()

		// Имитируем сложную работу
		time.Sleep(100 * time.Millisecond)
		processData(ctx)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "success", "data": "processed"}`))
	})

	log.Println("Starting server on :8008")
	log.Fatal(http.ListenAndServe(":8008", nil))
}

func processData(ctx context.Context) {
	_, span := tracer.Start(ctx, "processData")
	defer span.End()

	time.Sleep(30 * time.Millisecond)

	// Дополнительная "работа"
	_, childSpan := tracer.Start(ctx, "complexCalculation")
	time.Sleep(20 * time.Millisecond)
	childSpan.End()
}
