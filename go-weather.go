package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func main() {

	// CLIENT EXAMPLE: TO TELL IF SERVER ALIVE
	// curl http://localhost:8081/hello
	http.HandleFunc("/hello", hello)

	// CLIENT EXAMPLE: GET WEATHER OF A CITY
	// curl http://localhost:8081/weather/phoenix
	http.HandleFunc("/weather/",
		func(w http.ResponseWriter, r *http.Request) {
			city := strings.SplitN(r.URL.Path, "/", 3)[2]

			providerResults, err := queryWeatherData(city)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			for _, providerBlob := range providerResults {
				// send back a list of results
				json.NewEncoder(w).Encode(providerBlob)
			}
		})

	http.ListenAndServe(":8081", nil)
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from Go!\n"))
}

func queryWeatherData(city string) ([]weatherDataJSON, error) {
	apiConfig, err := loadAPIConfig("apiConfig")
	if err != nil {
		return []weatherDataJSON{}, err
	}

	mwp := multiWeatherProvider{
		openWeatherMap{apiKey: apiConfig.OpenWeatherMapAPIKey},
		weatherUnderground{apiKey: apiConfig.WUndergroundAPIKey},
	}

	return mwp.temperature(city)

}

func loadAPIConfig(filename string) (apiConfigData, error) {
	bytes, err := ioutil.ReadFile(filename)

	if err != nil {
		return apiConfigData{}, err
	}

	var c apiConfigData
	err = json.Unmarshal(bytes, &c)
	if err != nil {
		return apiConfigData{}, err
	}

	return c, nil
}

type weatherDataJSON struct {
	Name string `json:"name"`
	Main struct {
		ProviderName string  `json:"providerName"`
		Kelvin       float64 `json:"temp"`
	} `json:"main"`
}

type apiConfigData struct {
	OpenWeatherMapAPIKey string `json:"OpenWeatherMapApiKey"`
	WUndergroundAPIKey   string `json:"WUndergroundApiKey"`
}

type multiWeatherProvider []weatherProvider

type weatherProvider interface {
	temperature(city string) (weatherData, error) // temperature in Kelvin!
}

type openWeatherMap struct {
	apiKey string
}

type weatherData struct {
	temp         float64
	providerName string
}

func (mwp multiWeatherProvider) temperature(city string) ([]weatherDataJSON, error) {
	// Make a channel for temperatures, and a channel for errors.
	// Each provider will push a value into only one.
	var weatherDataBlobs []weatherDataJSON
	temps := make(chan weatherData, len(mwp))
	errs := make(chan error, len(mwp))

	// For each provider, spawn a goroutine with an anonymous function.
	// That function will invoke the temperature method, and forward the response.
	for _, provider := range mwp {
		go func(p weatherProvider) {
			wd, err := p.temperature(city)
			if err != nil {
				errs <- err
				return
			}
			temps <- wd
		}(provider)
	}

	// Collect a temperature or an error from each provider.
	for i := 0; i < len(mwp); i++ {
		select {
		case temp := <-temps:
			// new weather data json based on channel passed blob from provider
			var blob weatherDataJSON
			blob.Name = city
			blob.Main.ProviderName = temp.providerName
			blob.Main.Kelvin = temp.temp
			// combine slices
			weatherDataBlobs = append(weatherDataBlobs, blob)
		case err := <-errs:
			if err != nil {
				return weatherDataBlobs, err
			}
		}
	}

	// Return the average, same as before.
	return weatherDataBlobs, nil
}

func (w openWeatherMap) temperature(city string) (weatherData, error) {

	resp, err := http.Get("http://api.openweathermap.org/data/2.5/weather?APPID=" + w.apiKey + "&q=" + city)
	if err != nil {
		return weatherData{0, "openWeatherMap"}, err
	}
	defer resp.Body.Close()

	var d struct {
		Main struct {
			Kelvin float64 `json:"temp"`
		} `json:"main"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return weatherData{0, "openWeatherMap"}, err
	}

	log.Printf("openWeatherMap: %s: %.2f", city, d.Main.Kelvin)
	return weatherData{d.Main.Kelvin, "openWeatherMap"}, nil
}

type weatherUnderground struct {
	apiKey string
}

func (w weatherUnderground) temperature(city string) (weatherData, error) {

	// PRETEND WEATHER UNDERGROUND STILL WORKS AND GIVES A TEMP 60 deg F
	return weatherData{777.777, "weatherUnderground"}, nil
}
