package main

import (
	// "fmt"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func main() {

	// TO TELL IF SERVER ALIVE
	// curl http://localhost:8081/hello
	http.HandleFunc("/hello", hello)

	// GET WEATHER OF A CITY
	// curl http://localhost:8081/weather/phoenix
	http.HandleFunc("/weather/",
		func(w http.ResponseWriter, r *http.Request) {
			city := strings.SplitN(r.URL.Path, "/", 3)[2]

			data, err := query(city)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			json.NewEncoder(w).Encode(data)
		})

	http.ListenAndServe(":8081", nil)
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from Go!\n"))
}

func query(city string) (weatherData, error) {
	apiConfig, err := loadAPIConfig(".apiConfig")
	if err != nil {
		return weatherData{}, err
	}

	resp, err := http.Get("http://api.openweathermap.org/data/2.5/weather?APPID=" + apiConfig.OpenWeatherMapAPIKey + "&q=" + city)
	if err != nil {
		return weatherData{}, err
	}

	defer resp.Body.Close()

	var d weatherData

	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return weatherData{}, err
	}

	return d, nil
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

type weatherData struct {
	Name string `json:"name"`
	Main struct {
		Kelvin float64 `json:"temp"`
	} `json:"main"`
}

type apiConfigData struct {
	OpenWeatherMapAPIKey string `json:"OpenWeatherMapApiKey"`
}

type weatherProvider interface {
	temperature(city string) (float64, error) // temperature in Kelvin!
}

type openWeatherMap struct{}

func (w openWeatherMap) temperature(city string) (float64, error) {
	apiConfig, err := loadAPIConfig(".apiConfig")
	if err != nil {
		return 0, err
	}

	resp, err := http.Get("http://api.openweathermap.org/data/2.5/weather?APPID=" + apiConfig.OpenWeatherMapAPIKey + "&q=" + city)
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

	// PRETEND WEATHER UNDERGROUND STILL WORKS AND GIVES A TEMP
	return 288.706, nil

	// apiConfig, err := loadAPIConfig(".apiConfig")
	// if err != nil {
	// 	return 0, err
	// }
	// myKey := apiConfig.WundergroundAPIKey

	// resp, err := http.Get("http://api.wunderground.com/api/" + myKey + "/conditions/q/" + city + ".json")
	// if err != nil {
	// 	return 0, err
	// }

	// defer resp.Body.Close()

	// var d struct {
	// 	Observation struct {
	// 		Celsius float64 `json:"temp_c"`
	// 	} `json:"current_observation"`
	// }

	// if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
	// 	return 0, err
	// }

	// kelvin := d.Observation.Celsius + 273.15
	// log.Printf("wunderGround: %s: %.2f", city, kelvin)
	// return kelvin, nil
}
