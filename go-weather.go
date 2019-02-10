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

			dataowm, datawu, err := queryWeatherData(city)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			json.NewEncoder(w).Encode(dataowm)
			json.NewEncoder(w).Encode(datawu)
		})

	http.ListenAndServe(":8081", nil)
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from Go!\n"))
}

func queryWeatherData(city string) (weatherDataJSON, weatherDataJSON, error) {
	apiConfig, err := loadAPIConfig("apiConfig")
	if err != nil {
		return weatherDataJSON{}, weatherDataJSON{}, err
	}

	owm := openWeatherMap{apiKey: apiConfig.OpenWeatherMapAPIKey}
	wu := weatherUnderground{apiKey: apiConfig.WUndergroundAPIKey}

	respOWM, err := owm.temperature(city)
	respWU, err := wu.temperature(city)

	// resp, err := http.Get("http://api.openweathermap.org/data/2.5/weather?APPID=" + apiConfig.OpenWeatherMapAPIKey + "&q=" + city)
	// if err != nil {
	// 	return weatherDataJSON{}, err
	// }
	//defer resp.Body.Close()

	// parse respOWM into this struct
	var weatherDataBlobOWM weatherDataJSON
	weatherDataBlobOWM.Name = city
	weatherDataBlobOWM.Main.Kelvin = respOWM

	// parse respWU into this struct
	var weatherDataBlobWU weatherDataJSON
	weatherDataBlobWU.Name = city
	weatherDataBlobWU.Main.Kelvin = respWU

	// if err := json.NewDecoder(resp.Body).Decode(&weatherDataBlobOWM); err != nil {
	// 	return weatherDataJSON{}, err
	// }

	//TODO will return 2 of weather data blobs
	return weatherDataBlobOWM, weatherDataBlobWU, nil
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
		Kelvin float64 `json:"temp"`
	} `json:"main"`
}

type apiConfigData struct {
	OpenWeatherMapAPIKey string `json:"OpenWeatherMapApiKey"`
	WUndergroundAPIKey   string `json:"WUndergroundApiKey"`
}

type weatherProvider interface {
	temperature(city string) (float64, error) // temperature in Kelvin!
}

type openWeatherMap struct {
	apiKey string
}

func (w openWeatherMap) temperature(city string) (float64, error) {

	resp, err := http.Get("http://api.openweathermap.org/data/2.5/weather?APPID=" + w.apiKey + "&q=" + city)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var d struct {
		Main struct {
			Kelvin float64 `json:"temp"`
		} `json:"main"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return 0, err
	}

	log.Printf("openWeatherMap: %s: %.2f", city, d.Main.Kelvin)
	return d.Main.Kelvin, nil
}

type weatherUnderground struct {
	apiKey string
}

func (w weatherUnderground) temperature(city string) (float64, error) {

	// PRETEND WEATHER UNDERGROUND STILL WORKS AND GIVES A TEMP 60 deg F
	return 777.777, nil
}
