package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/osniantonio/lab2-tracing-distribuido-span/internal/serviceb/temp"
	"github.com/osniantonio/lab2-tracing-distribuido-span/internal/serviceb/viacep"
	"github.com/osniantonio/lab2-tracing-distribuido-span/internal/serviceb/weather"
	"go.opentelemetry.io/otel/trace"
)

type input struct {
	CEP string `json:"cep"`
}

type output struct {
	City string  `json:"city"`
	C    float64 `json:"temp_C"`
	F    float64 `json:"temp_F"`
	K    float64 `json:"temp_K"`
}

type DefaultHandler struct {
	ViaCepApi  *viacep.ViaCepApi
	WeatherApi *weather.WeatherApi
	Tracer     trace.Tracer
}

func (h *DefaultHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	i := &input{}
	if err := json.NewDecoder(r.Body).Decode(i); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	match, err := regexp.MatchString("\\b\\d{5}-?\\d{3}\\b", i.CEP)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if !match {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}
	city, err := h.searchCEP(i.CEP, ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	temp, err := h.searchTemp(city, ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	o := &output{
		City: city,
		C:    temp.C,
		F:    temp.F,
		K:    temp.K,
	}
	body, err := json.Marshal(o)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func (h *DefaultHandler) searchCEP(cep string, ctx context.Context) (string, error) {
	_, span := h.Tracer.Start(ctx, "cep")
	defer span.End()
	return h.ViaCepApi.Search(cep)
}

func (h *DefaultHandler) searchTemp(city string, ctx context.Context) (*temp.Temp, error) {
	_, span := h.Tracer.Start(ctx, "temperature")
	defer span.End()
	return h.WeatherApi.Search(city)
}
