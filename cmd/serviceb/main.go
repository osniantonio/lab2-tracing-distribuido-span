package main

import (
	"context"
	"log"
	"os"

	"github.com/osniantonio/lab2-tracing-distribuido-span/internal/otel/setup"
	"github.com/osniantonio/lab2-tracing-distribuido-span/internal/otel/tracing"
	"github.com/osniantonio/lab2-tracing-distribuido-span/internal/serviceb/handler"
	"github.com/osniantonio/lab2-tracing-distribuido-span/internal/serviceb/viacep"
	"github.com/osniantonio/lab2-tracing-distribuido-span/internal/serviceb/weather"
)

func main() {
	ctx := context.Background()
	tracer, shutdow := tracing.Start()
	defer func() {
		_ = shutdow(ctx)
	}()
	viaCepApi := &viacep.ViaCepApi{}
	weatherApi := &weather.WeatherApi{
		Key: os.Getenv("API_KEY"),
	}
	h := &handler.DefaultHandler{
		ViaCepApi:  viaCepApi,
		WeatherApi: weatherApi,
		Tracer:     tracer,
	}
	if err := setup.Run(h); err != nil {
		log.Fatalln(err)
	}
}
