package clima_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	climatePkg "github.com/osniantonio/lab2-tracing-distribuido-span/internal/pkg/climate"
	"github.com/stretchr/testify/assert"
)

func TestSearchTemperature(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"main": {"temp": 25.5}}`))
	}))
	defer server.Close()

	ctx := context.Background()

	// Teste caso de busca de temperatura válida para uma cidade
	city := "London"
	isValid, statusCode, InvalidMessage, temperature := climatePkg.SearchTemperature(ctx, city)
	assert.Equal(t, isValid, true)
	assert.Equal(t, statusCode, http.StatusOK)
	assert.Equal(t, InvalidMessage, "")
	assert.NotNil(t, temperature)

	// Validando as conversões
	tempC, err := strconv.ParseFloat(temperature.Temp_C, 64)
	if err != nil {
		fmt.Println("Error converting temperature to float64:", err)
		return
	}

	tempCStr := fmt.Sprintf("%.1f", tempC)
	tempFStr := fmt.Sprintf("%.1f", climatePkg.ToFahrenheit(tempC))
	tempKStr := fmt.Sprintf("%.1f", climatePkg.ToKelvin(tempC))

	temperatureFromCelsius := climatePkg.NewTemperature(city, tempC)

	assert.Equal(t, tempCStr, temperatureFromCelsius.Temp_C)
	assert.Equal(t, tempFStr, temperatureFromCelsius.Temp_F)
	assert.Equal(t, tempKStr, temperatureFromCelsius.Temp_K)
}
