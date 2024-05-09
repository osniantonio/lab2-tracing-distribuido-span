package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"regexp"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type input struct {
	CEP string `json:"cep"`
}

type DefaultHandler struct {
	Tracer trace.Tracer
}

func (h *DefaultHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := h.Tracer.Start(ctx, "input", trace.WithNewRoot())
	defer span.End()
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
	b, err := json.Marshal(i)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, os.Getenv("API_URL"), bytes.NewBuffer(b))
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	carrier := propagation.HeaderCarrier(req.Header)
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()
	buf := &bytes.Buffer{}
	buf.ReadFrom(res.Body)
	if res.StatusCode == http.StatusOK {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(res.StatusCode)
	w.Write(buf.Bytes())
}
