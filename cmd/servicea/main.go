package main

import (
	"context"
	"log"

	"github.com/osniantonio/lab2-tracing-distribuido-span/internal/otel/setup"
	"github.com/osniantonio/lab2-tracing-distribuido-span/internal/otel/tracing"
	"github.com/osniantonio/lab2-tracing-distribuido-span/internal/servicea/handler"
)

func main() {
	ctx := context.Background()
	tracer, shutdow := tracing.Start()
	defer func() {
		_ = shutdow(ctx)
	}()
	h := &handler.DefaultHandler{
		Tracer: tracer,
	}
	if err := setup.Run(h); err != nil {
		log.Fatalln(err)
	}
}
