package clima

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

const weatherAPIKey = "8a140ace6b42089b39b3450624b0f495"

type Temperatura struct {
	City   string
	Temp_F string
	Temp_C string
	Temp_K string
}

func ToFahrenheit(celsius float64) float64 {
	return (celsius * 1.8) + 32
}

func ToKelvin(celsius float64) float64 {
	return celsius + 273.15
}

func NewTemperature(city string, tempC float64) Temperatura {
	temp := Temperatura{}
	temp.City = city
	temp.Temp_C = fmt.Sprintf("%.1f", tempC)
	temp.Temp_F = fmt.Sprintf("%.1f", ToFahrenheit(tempC))
	temp.Temp_K = fmt.Sprintf("%.1f", ToKelvin(tempC))
	return temp
}

func SearchTemperature(ctx context.Context, city string) (bool, int, string, Temperatura) {
	urlStr := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric", url.QueryEscape(city), weatherAPIKey)

	resp, err := http.Get(urlStr)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return false, http.StatusRequestTimeout, "request timeout", Temperatura{}
		}
		return false, http.StatusInternalServerError, err.Error(), Temperatura{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorMessage := fmt.Sprintf("API temperature request failed: code %d, status: %s, city: %s", resp.StatusCode, resp.Status, city)
		return false, resp.StatusCode, errorMessage, Temperatura{}
	}

	decoder := json.NewDecoder(resp.Body)
	var temperatureData map[string]interface{}
	if err := decoder.Decode(&temperatureData); err != nil {
		return false, http.StatusBadRequest, "fail to decode resp.Body", Temperatura{}
	}

	if mainData, ok := temperatureData["main"].(map[string]interface{}); ok {
		if celsiusTemperature, ok := mainData["temp"].(float64); ok {
			return true, http.StatusOK, "", NewTemperature(city, celsiusTemperature)
		}
	}

	return false, http.StatusNotFound, "error: temperature not found for the city", Temperatura{}
}
