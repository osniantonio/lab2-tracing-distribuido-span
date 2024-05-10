package viacep

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const (
	VIACEP_URL = "http://viacep.com.br/ws/%s/json/"
)

var (
	errCepNotFound = errors.New("can not find zipcode")
)

type response struct {
	City    string `json:"localidade"`
	Errored string `json:"erro"`
}

type ViaCepApi struct{}

func (a *ViaCepApi) Search(cep string) (string, error) {
	url := fmt.Sprintf(VIACEP_URL, cep)
	res, err := http.Get(url)
	if err != nil {
		return "", errCepNotFound
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", errCepNotFound
	}
	r := &response{}
	if err := json.NewDecoder(res.Body).Decode(r); err != nil {
		return "", errCepNotFound
	}
	if r.Errored == "true" {
		return "", errCepNotFound
	}
	return r.City, nil
}
