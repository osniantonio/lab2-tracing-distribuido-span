package cep_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	cepPkg "github.com/osniantonio/lab2-tracing-distribuido-span/internal/pkg/cep"
)

func TestCEPValidFormat(t *testing.T) {
	assert.Equal(t, cepPkg.CepValid("12345-678"), true)
	assert.Equal(t, cepPkg.CepValid("09876-543"), true)
}

func TestCEPInvalidFormat(t *testing.T) {
	assert.Equal(t, cepPkg.CepValid("123456789"), false)
	assert.Equal(t, cepPkg.CepValid("123456"), false)
	assert.Equal(t, cepPkg.CepValid("12345-abc"), false)
	assert.Equal(t, cepPkg.CepValid("invalid-cep"), false)
	assert.Equal(t, cepPkg.CepValid(""), false)
}

func TestSearchAddress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"bairro": "Centro", "cep": "89198-000", "complemento": "", "ddd": "47", "gia": "", "ibge": "4217300", "localidade": "Rio do Sul", "logradouro": "", "siafi": "8159", "uf": "SC"}`))
	}))
	defer server.Close()

	ctx := context.Background()

	// Teste caso de busca de um CEP válido
	isValid, statusCode, _, _ := cepPkg.SearchAddress(ctx, "89198-000")
	assert.Equal(t, isValid, true)
	assert.Equal(t, statusCode, http.StatusOK)

	// Teste caso de busca de um CEP inválido
	isValid, statusCode, _, _ = cepPkg.SearchAddress(ctx, "invalid-cep")
	assert.Equal(t, isValid, false)
	assert.Equal(t, statusCode, http.StatusUnprocessableEntity)

	// Teste caso de busca de um CEP não encontrado
	isValid, statusCode, _, _ = cepPkg.SearchAddress(ctx, "12345678")
	assert.Equal(t, isValid, false)
	assert.Equal(t, statusCode, http.StatusNotFound)
}
