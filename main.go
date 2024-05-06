package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// go:embed template/*.html
var templateContent embed.FS

type Webserver struct {
	TemplateData *TemplateData
}

const (
	API_PORT             = 8080
	TIMEOUT_API          = 100 * time.Millisecond
	ORCHESTRATOR_API_URL = "http://localhost:8081/weather"
)

type CEP struct {
	CEP string
}

type TemplateData struct {
	Title              string
	BackgroundColor    string
	ResponseTime       time.Duration
	ExternalCallURL    string
	ExternalCallMethod string
	Content            string
	RequestNameOTEL    string
	OTELTracer         trace.Tracer
}

func initProvider(serviceName, collectorURL string) (func(context.Context) error, error) {
	ctx := context.Background()

	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	conn, err := grpc.DialContext(
		ctx,
		collectorURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	fmt.Println("aqui")
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tracerProvider.Shutdown, nil
}

func NewServer(templateData *TemplateData) *Webserver {
	return &Webserver{
		TemplateData: templateData,
	}
}

func (we *Webserver) CreateServer() *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)
	router.Use(middleware.Timeout(60 * time.Second))
	router.Handle("/metrics", promhttp.Handler())
	router.Post("/weather", we.HandleRequest)
	return router
}

func (h *Webserver) HandleRequest(w http.ResponseWriter, r *http.Request) {
	carrier := propagation.HeaderCarrier(r.Header)
	ctx := r.Context()
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

	ctx, span := h.TemplateData.OTELTracer.Start(ctx, "SPAN_EXTERNAL_REQUEST")
	if h.TemplateData.ExternalCallURL != "" {
		var req *http.Request
		var err error

		if h.TemplateData.ExternalCallMethod == "GET" {
			req, err = http.NewRequestWithContext(ctx, "GET", h.TemplateData.ExternalCallURL, nil)
		} else if h.TemplateData.ExternalCallMethod == "POST" {
			req, err = http.NewRequestWithContext(ctx, "POST", h.TemplateData.ExternalCallURL, nil)
		} else {
			http.Error(w, "Invalid ExternalCallMethod", http.StatusInternalServerError)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		h.TemplateData.Content = string(bodyBytes)
	}
	span.End()

	var cep CEP
	_, span_read_response := h.TemplateData.OTELTracer.Start(ctx, "SPAN_READ_RESPONSE")
	body_data, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	span_read_response.End()

	_, span_parse_response := h.TemplateData.OTELTracer.Start(ctx, "SPAN_PARSE_RESPONSE")
	err = json.Unmarshal(body_data, &cep)
	if err != nil {
		panic(err)
	}
	span_parse_response.End()

	_, span_cep_validation := h.TemplateData.OTELTracer.Start(ctx, "SPAN_CEP_VALIDATION")
	if len(cep.CEP) != 8 || reflect.TypeOf(cep.CEP).String() != "string" {
		log.Printf("invalid zipcode")
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}
	span_cep_validation.End()

	// _, span_prepare_orchestrator_api_request := h.TemplateData.OTELTracer.Start(ctx, "SPAN_PREPARE_ORCHESTRATOR_API_REQUEST")
	fmt.Println("input: call the service-b with cep " + cep.CEP)
	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT_API)
	defer cancel()
	endpoint := fmt.Sprintf("%s?cep=%s", ORCHESTRATOR_API_URL, cep.CEP)
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, nil)
	if err != nil {
		log.Fatalf("Fail to create the request: %v", err)
		return
	}
	// span_prepare_orchestrator_api_request.End()

	_, span_call_orchestrator_api := h.TemplateData.OTELTracer.Start(ctx, "SPAN_CALL_ORCHESTRATOR_API")
	defer span_call_orchestrator_api.End()
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Fail to make the request: %v", err)
		return
	}
	defer res.Body.Close()
	ctx_err := ctx.Err()
	if ctx_err != nil {
		select {
		case <-ctx.Done():
			err := ctx.Err()
			log.Fatalf("Max timeout reached: %v", err)
			return
		}
	}
}

func init() {
	viper.AutomaticEnv()
}

func main() {
	// ---------- gracefull shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// ---------- provider
	shutdown, err := initProvider(viper.GetString("OTEL_SERVICE_NAME"), viper.GetString("OTEL_EXPORTER_OTLP_ENDPOINT"))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Fatal("failed to shutdown TracerProvider: %w", err)
		}
	}()
	tracer := otel.Tracer("intput-api-tracer")
	templateData := &TemplateData{
		Title:              viper.GetString("TITLE"),
		BackgroundColor:    viper.GetString("BACKGROUND_COLOR"),
		ResponseTime:       viper.GetDuration("RESPONSE_TIME"),
		ExternalCallURL:    viper.GetString("EXTERNAL_CALL_URL"),
		ExternalCallMethod: viper.GetString("EXTERNAL_CALL_METHOD"),
		RequestNameOTEL:    viper.GetString("REQUEST_NAME_OTEL"),
		OTELTracer:         tracer,
	}
	server := NewServer(templateData)
	router := server.CreateServer()

	go func() {
		log.Println("Starting server on port", viper.GetString("HTTP_PORT"))
		if err := http.ListenAndServe(viper.GetString("HTTP_PORT"), router); err != nil {
			log.Fatal(err)
		}
	}()

	select {
	case <-sigCh:
		log.Println("Shutting down gracefully, CTRL+c pressed...")
	case <-ctx.Done():
		log.Println("Shutting down due other reason...")
	}

	_, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
}
