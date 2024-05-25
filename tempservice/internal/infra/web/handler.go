package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type TemperatureOutputDTO struct {
	City       string `json:"city"`
	Celsius    string `json:"temp_C"`
	Fahrenheit string `json:"temp_F"`
	Kelvin     string `json:"temp_K"`
}

type ViaCEP struct {
	Localidade string `json:"localidade"`
}

type WeatherAPI struct {
	Current Current `json:"current"`
}

type Current struct {
	CelsiusTemperature    float64 `json:"temp_c"`
	FarhenheitTemperature float64 `json:"temp_f"`
}

var viaCepApiURL = "http://viacep.com.br/ws/"
var weatherApiURL = "http://api.weatherapi.com/v1/current.json"

const apiKey = "bdb5695826b246b5a47230638241405"

var (
	tracer = otel.Tracer("temp-service")
)

func HandleTemperature(w http.ResponseWriter, r *http.Request) {

	var span trace.Span
	ctx, span := tracer.Start(r.Context(), "HandleTemperature")
	defer span.End()

	zipCode := strings.Trim(r.URL.Path, "/temperature/")

	location, err := fetchLocationByZipCode(ctx, zipCode)
	if err != nil {
		slog.Error("failed to fetch location by zipCode", "input:", zipCode, "error", err)
		http.Error(w, "cannot find zipCode", http.StatusNotFound)
		return
	}

	temperature, err := fetchTemperature(ctx, location)
	if err != nil {
		slog.Error("failed to fetch temperature by location", "input:", location.Localidade, "error", err)
		http.Error(w, "could not get weather", http.StatusInternalServerError)
		return
	}

	formatCelcius := fmt.Sprintf("%.1f", temperature.Current.CelsiusTemperature)

	var dto TemperatureOutputDTO = TemperatureOutputDTO{
		City:       location.Localidade,
		Celsius:    formatCelcius,
		Fahrenheit: ConvertCelsiusToFahrenheit(temperature.Current.CelsiusTemperature),
		Kelvin:     ConvertCelsiusToKelvin(temperature.Current.CelsiusTemperature),
	}

	byteJson, err := json.Marshal(dto)
	if err != nil {
		slog.Error("failed to marshal dto", dto, "error", err)
		http.Error(w, "could not get temperature", http.StatusInternalServerError)
		return
	}

	w.Write(byteJson)
}

func fetchLocationByZipCode(ctx context.Context, zipCode string) (ViaCEP, error) {

	var span trace.Span
	ctx, span = tracer.Start(ctx, "fetchLocationByZipCode")
	defer span.End()

	req, err := http.NewRequestWithContext(ctx, "GET", viaCepApiURL+zipCode+"/json/", nil)
	if err != nil {
		slog.Error("unable to make new request with context", "ctx", ctx, "error", err)
		return ViaCEP{}, err
	}

	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("unable to do request", "req:", req.URL.Path, "error", err)
		return ViaCEP{}, err
	}

	defer resp.Body.Close()

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("error reading response", "response:", resp.Body, "error", err)
		return ViaCEP{}, err
	}

	var viaCepData ViaCEP

	err = json.Unmarshal(result, &viaCepData)
	if err != nil {
		slog.Error("error unmarshal result", "result:", string(result), "error", err)
		return ViaCEP{}, err
	}

	if viaCepData.Localidade == "" {
		err = fmt.Errorf("error validating location: %s", viaCepData.Localidade)
		slog.Error("location is empty", "error", err)
		return ViaCEP{}, err
	}

	return viaCepData, nil
}

func fetchTemperature(ctx context.Context, v ViaCEP) (WeatherAPI, error) {

	var span trace.Span
	ctx, span = tracer.Start(ctx, "fetchTemperature")
	defer span.End()

	params := map[string]string{
		"key": apiKey,
		"q":   v.Localidade,
		"aqi": "no",
	}

	u, err := url.Parse(weatherApiURL)
	if err != nil {
		slog.Error("error parsing URL", "url", weatherApiURL, "error", err)
		return WeatherAPI{}, err
	}

	q := u.Query()
	for key, value := range params {
		q.Set(key, value)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		slog.Error("unable to make new request with context", "ctx", ctx, "error", err)
		return WeatherAPI{}, err
	}

	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("error sending request", "query", u.RawQuery, "error", err)
		return WeatherAPI{}, err
	}

	defer resp.Body.Close()

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("error reading response", "response:", resp.Body, "error", err)
		return WeatherAPI{}, err
	}

	var weather WeatherAPI

	err = json.Unmarshal(result, &weather)
	if err != nil {
		slog.Error("error unmarshal result", "result:", string(result), "error", err)
		return WeatherAPI{}, err
	}

	return weather, nil
}

func ConvertCelsiusToFahrenheit(celsius float64) string {
	var f = celsius*1.8 + 32
	return fmt.Sprintf("%.1f", f)

}

func ConvertCelsiusToKelvin(celsius float64) string {
	var k = celsius + 273
	return fmt.Sprintf("%.1f", k)
}
