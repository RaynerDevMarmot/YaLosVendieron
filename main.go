package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type RateResponse struct {
	Value       float64 `json:"value"`
	LastUpdated string  `json:"last_updated"`
}

type ConversionResponse struct {
	Input     float64 `json:"input"`
	Converted float64 `json:"converted"`
	Rate      float64 `json:"rate"`
}

var cachedRate float64
var lastUpdated time.Time

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.HandleFunc("/api/rate", handleRate)
	http.HandleFunc("/api/to-bs", handleToBs)
	http.HandleFunc("/api/to-usd", handleToUsd)

	fmt.Println("Servidor iniciado en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleRate(w http.ResponseWriter, r *http.Request) {
	rate := getRate()
	resp := RateResponse{
		Value:       rate,
		LastUpdated: lastUpdated.Format("02/01/2006 15:04:05"),
	}
	jsonResponse(w, resp)
}

func handleToBs(w http.ResponseWriter, r *http.Request) {
	usdStr := r.URL.Query().Get("usd")
	usd, err := strconv.ParseFloat(usdStr, 64)
	if err != nil || usd < 0 {
		http.Error(w, "Cantidad inválida", http.StatusBadRequest)
		return
	}
	rate := getRate()
	resp := ConversionResponse{
		Input:     usd,
		Converted: usd * rate,
		Rate:      rate,
	}
	jsonResponse(w, resp)
}

func handleToUsd(w http.ResponseWriter, r *http.Request) {
	bsStr := r.URL.Query().Get("bs")
	bs, err := strconv.ParseFloat(bsStr, 64)
	if err != nil || bs < 0 {
		http.Error(w, "Cantidad inválida", http.StatusBadRequest)
		return
	}
	rate := getRate()
	resp := ConversionResponse{
		Input:     bs,
		Converted: bs / rate,
		Rate:      rate,
	}
	jsonResponse(w, resp)
}

func getRate() float64 {
	if time.Since(lastUpdated) < 15*time.Minute && cachedRate > 0 {
		return cachedRate
	}

	// Usa tu API key aquí
	apiKey := "2dbb23de20e25c56ebc1f143"
	url := "https://v6.exchangerate-api.com/v6/" + apiKey + "/latest/USD"
	res, err := http.Get(url)
	if err != nil {
		log.Println("Error al consultar exchangerate-api:", err)
		return cachedRate
	}
	defer res.Body.Close()

	var result struct {
		ConversionRates map[string]float64 `json:"conversion_rates"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		log.Println("Error al decodificar JSON:", err)
		return cachedRate
	}

	newRate := result.ConversionRates["VES"]
	if newRate > 1 {
		cachedRate = newRate
		lastUpdated = time.Now()
	}

	return cachedRate
}

func sanitize(s string) string {
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", ".")
	s = strings.TrimSpace(s)
	return s
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
