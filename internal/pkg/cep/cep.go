package cep

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"
)

type Address struct {
	Neighborhood string `json:"bairro"`
	Cep          string `json:"cep"`
	Complemento  string `json:"complemento"`
	DDD          string `json:"ddd"`
	GIA          string `json:"gia"`
	IBGE         string `json:"ibge"`
	City         string `json:"localidade"`
	PublicPlace  string `json:"logradouro"`
	Siafi        string `json:"siafi"`
	UF           string `json:"uf"`
}

func cepDigitsValid(cep string) bool {
	return regexp.MustCompile(`^\d{8}$`).MatchString(cep)
}

func cepFormatValid(cep string) bool {
	return regexp.MustCompile(`^\d{5}-\d{3}$`).MatchString(cep)
}

func CepValid(cep string) bool {
	return cepDigitsValid(cep) || cepFormatValid(cep)
}

func SearchAddress(ctx context.Context, cep string) (bool, int, string, Address) {
	if !CepValid(cep) {
		return false, http.StatusUnprocessableEntity, "invalid zipcode", Address{}
	}

	url := fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep)
	client := http.Client{Timeout: time.Second}
	resp, err := client.Get(url)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return false, http.StatusRequestTimeout, "request timeout", Address{}
		}
		return false, http.StatusInternalServerError, err.Error(), Address{}
	}
	defer resp.Body.Close()

	var address Address
	if err := json.NewDecoder(resp.Body).Decode(&address); err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return false, http.StatusNotFound, "can not find zipcode", Address{}
		} else {
			return false, http.StatusInternalServerError, "invalid response format", Address{}
		}
	}

	if len(address.Cep) == 0 {
		return false, http.StatusNotFound, "can not find zipcode", Address{}
	}

	return true, http.StatusOK, "", address
}
