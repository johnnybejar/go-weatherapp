package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type PageData struct {
	City, Country, Weather, Icon string
	Temperature, Wind int
}

type Response struct {
	Base string `json:"base"`
	Clouds map[string]string `json:"clouds"`
	Cod int `json:"cod"`
	Coord map[string]float32 `json:"coord"`
	Dt int `json:"dt"`
	Id int `json:"id"`
	Main map[string]interface{} `json:"main"`
	Name string `json:"name"`
	Sys map[string]interface{} `json:"sys"`
	Timezone int `json:"timezone"`
	Visibility int `json:"visibility"`
	WeatherCondition []Weather `json:"weather"`
	Wind map[string]interface{} `json:"wind"`
}

type Weather struct {
	Id int `json:"id"`
	Main string `json:"main"`
	Description string `json:"description"`
	Icon string `json:"icon"`
}

func getWeatherData(req string, key string) (*http.Response, error) {
	reqUrl := fmt.Sprintf(
		"https://api.openweathermap.org/data/2.5/weather?q=%s&units=imperial&appid=%s",
		req,
		key,
	)

	res, err := http.Get(reqUrl)

	return res, err
}

func handler(w http.ResponseWriter, r *http.Request) {
	template, err := template.ParseFiles("index.html")
	if err != nil {
		log.Fatal(err)
	}

	godotenv.Load(".env")
	API_KEY := os.Getenv("API_KEY")

	res, err := getWeatherData("wilmington", API_KEY)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		http.Error(w, "Failed to read API response", http.StatusInternalServerError)
		log.Fatal(err)
	}

	var apiRes Response

	err = json.Unmarshal(body, &apiRes)
	if err != nil {
		http.Error(w, "Failed to parse API response", http.StatusInternalServerError)
		log.Fatal(err)
	}

	err = template.Execute(w, apiRes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatal(err)
	}
}

func main() {
	http.HandleFunc("/", handler)

	fmt.Printf("Listening on port 8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

