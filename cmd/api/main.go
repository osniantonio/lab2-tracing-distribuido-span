package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	cepPkg "github.com/osniantonio/lab2-tracing-distribuido-span/internal/pkg/cep"
	climatePkg "github.com/osniantonio/lab2-tracing-distribuido-span/internal/pkg/climate"
)

const (
	API_PORT = 8081
)

func main() {
	fmt.Println("Start running app...")
	router := mux.NewRouter()
	router.HandleFunc("/temperatures/{cep}", handleTemperatureRequest)
	fmt.Println("endpoint /temperatures/{cep} was created.")
	http.ListenAndServe(fmt.Sprintf(":%d", API_PORT), router)
}

func handleTemperatureRequest(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	cep := mux.Vars(r)["cep"]
	if len(cep) == 0 {
		cep = r.FormValue("cep")
	}

	isValid, StatusCode, InvalidMessage, Address := cepPkg.SearchAddress(ctx, cep)
	if !isValid {
		http.Error(w, InvalidMessage, StatusCode)
		return
	}

	isValid, StatusCode, InvalidMessage, Temperature := climatePkg.SearchTemperature(ctx, Address.City)
	if !isValid {
		http.Error(w, InvalidMessage, StatusCode)
		return
	}

	jsonBytes, err := json.Marshal(Temperature)
	if err != nil {
		http.Error(w, "Fail to generate the JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}
