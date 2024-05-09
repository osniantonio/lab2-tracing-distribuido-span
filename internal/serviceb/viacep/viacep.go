package viacep

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var (
	errUnabletoFindCep = errors.New("can not find zipcode")
)

type response struct {
	City    string `json:"localidade"`
	Errored string `json:"erro"`
}

type ViaCepApi struct{}

func (a *ViaCepApi) Search(cep string) (string, error) {
	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)
	res, err := http.Get(url)
	if err != nil {
		return "", errUnabletoFindCep
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", errUnabletoFindCep
	}
	r := &response{}
	if err := json.NewDecoder(res.Body).Decode(r); err != nil {
		return "", errUnabletoFindCep
	}
	if r.Errored == "true" {
		return "", errUnabletoFindCep
	}
	return r.City, nil
}
