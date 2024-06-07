package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type WeatherPageData struct {
	WeatherData *Response
	IconURL string
	Error bool
}

func (wpd *WeatherPageData) setError(err bool) {
	wpd.Error = err
}

type Response struct {
	Base string 				`json:"base"`
	Clouds map[string]int 		`json:"clouds"`
	Cod int 					`json:"cod"`
	Coord map[string]float32 	`json:"coord"`
	Dt int 						`json:"dt"`
	Id int 						`json:"id"`
	Main Main 					`json:"main"`
	Name string 				`json:"name"`
	Sys Sys 					`json:"sys"`
	Timezone int 				`json:"timezone"`
	Visibility int 				`json:"visibility"`
	WeatherCondition []Weather 	`json:"weather"`
	Wind Wind 					`json:"wind"`
}

type Main struct {
	FeelsLike float64 			`json:"feels_like"`
	Humidity int 				`json:"humidity"`
	Pressure int 				`json:"pressure"`
	Temp float64 				`json:"temp"`
	TempMax float64 			`json:"temp_max"`
	TempMin float64 			`json:"temp_min"`
}

type Sys struct {
	Country string 				`json:"country"`
	Id int 						`json:"id"`
	Sunrise int 				`json:"sunrise"`
	Sunset int 					`json:"sunset"`
	// Type int `json:"type"` (not needed)
}

type Weather struct {
	Id int 						`json:"id"`
	Main string 				`json:"main"`
	Description string 			`json:"description"`
	Icon string 				`json:"icon"`
}

type Wind struct {
	Degree int 					`json:"def"`
	Speed float64 				`json:"speed"`
}

func getWeatherData(search string, key string) (*http.Response, error) {
	search = strings.Replace(search, " ", "+", -1)

	reqUrl := fmt.Sprintf(
		"https://api.openweathermap.org/data/2.5/weather?q=%s&units=imperial&appid=%s",
		search,
		key,
	)

	res, err := http.Get(reqUrl)

	return res, err
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "" {
		 http.NotFound(w, r)
		 fmt.Fprintf(w, "Go back to the root directory/path")
		 return
	}

	template, err := template.ParseFiles("index.html")
	if err != nil {
		log.Fatal(err)
	}

	godotenv.Load(".env")
	API_KEY := os.Getenv("API_KEY")

	var search string
	var pageData *WeatherPageData = &WeatherPageData{nil, "", false}

	if r.Method == http.MethodPost {
		search = r.FormValue("search")

		res, err := getWeatherData(search, API_KEY)
		if res.StatusCode >= 400 || err != nil {
			pageData.setError(true)
			err = template.Execute(w, pageData)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Fatal(err)
			}
			return
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

		pageData = &WeatherPageData{&apiRes, fmt.Sprintf("http://openweathermap.org/img/w/%s.png", apiRes.WeatherCondition[0].Icon), false}
	}

	err = template.Execute(w, pageData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatal(err)
	}
}

func main() {
	http.HandleFunc("/", handler)
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css/"))))

	fmt.Printf("Listening on port 8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

