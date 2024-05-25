package web

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type TemperatureInputDTO struct {
	Zipcode string `json:"cep"`
}

type TemperatureOutputDTO struct {
	City       string `json:"city"`
	Celsius    string `json:"temp_C"`
	Fahrenheit string `json:"temp_F"`
	Kelvin     string `json:"temp_K"`
}

var (
	tracer = otel.Tracer("cep-service")
)

var Host string
var URL string

func init() {
	Host = os.Getenv("URL_TEMP")
	if Host == "" {
		Host = "localhost"
	}

	URL = "http://" + Host + ":9090/temperature/"
}

func HandleZipcode(w http.ResponseWriter, r *http.Request) {

	var span trace.Span
	ctx, span := tracer.Start(r.Context(), "handleZipcode")
	defer span.End()

	var inputDto TemperatureInputDTO

	err := json.NewDecoder(r.Body).Decode(&inputDto)
	if err != nil {
		slog.Error("unable to decode", inputDto.Zipcode, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	isValidZipcode := isValidZipcode(inputDto.Zipcode)

	if !isValidZipcode {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	response, err := fetchTemperatureByZipCode(ctx, inputDto.Zipcode)
	if err != nil {
		http.Error(w, `"unable to fetch temperature by zipcode"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func fetchTemperatureByZipCode(ctx context.Context, zipCode string) ([]byte, error) {

	var span trace.Span
	ctx, span = tracer.Start(ctx, "fetchTemperatureByZipCode")
	defer span.End()

	req, err := http.NewRequestWithContext(ctx, "GET", URL+zipCode, nil)
	if err != nil {
		slog.Error("unable to make new request with context", "ctx", ctx, "error", err)
		return nil, err
	}

	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("unable to do request", "req:", req.URL.Path, "error", err)
		return nil, err
	}

	defer resp.Body.Close()
	

	return io.ReadAll(resp.Body)
}

func isValidZipcode(zipCode string) bool {
	re := regexp.MustCompile(`^\d{8}$`)
	return re.MatchString(zipCode)
}
