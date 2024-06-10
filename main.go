/*
Simple website that gets weather data from OpenWeatherMap API in Go
*/
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Data that will be used by the template in index.html
// WeatherData is a pointer so we can set it to
// nil when page is initialized or bad request
type WeatherPageData struct {
	WeatherData *Response
	IconURL string
	Error bool
}

// Setter for any errors encountered
func (wpd *WeatherPageData) setError(err bool) {
	wpd.Error = err
}

// The full response from OpenWeatherMap API
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

// Round the temps from main
func (m *Main) roundTemps() {
	m.FeelsLike = math.Round(m.FeelsLike*100/100)
	m.Temp = math.Round(m.Temp*100/100)
	m.TempMax = math.Round(m.TempMax*100/100)
	m.TempMin = math.Round(m.TempMin*100/100)
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
	Degree int 					`json:"deg"`
	Speed float64 				`json:"speed"`
}

// Calls the OpenWeatherMap API
func getWeatherData(search string, key string) (*http.Response, error) {
	// Replaces all spaces with "+", as that is how OpenWeatherMap represents spaces
	// Ex. "Los Angeles" -> "Los+Angeles"
	search = strings.Replace(search, " ", "+", -1)

	// Build the request url
	reqUrl := fmt.Sprintf(
		"https://api.openweathermap.org/data/2.5/weather?q=%s&units=imperial&appid=%s",
		search,
		key,
	)

	// Note that the response can be a 400 and err be nil, but is handled in the handler below
	res, err := http.Get(reqUrl)

	return res, err
}

// Handles requests to the server
func handler(w http.ResponseWriter, r *http.Request) {
	// Any path that isn't the root will be a not found
	if r.URL.Path != "/" && r.URL.Path != "" {
		 http.NotFound(w, r)
		 fmt.Fprintf(w, "Go back to the root directory/path")
		 return
	}
	
	// Creates the template from index.html
	template, err := template.ParseFiles("index.html")
	if err != nil {
		log.Fatal(err)
	}

	// Load environment variables
	godotenv.Load(".env")
	API_KEY := os.Getenv("API_KEY")

	var search string
	var pageData WeatherPageData = WeatherPageData{nil, "", false}

	// Check if it was a search
	if r.Method == http.MethodPost {
		search = r.FormValue("search")

		// Get weather data and handle any errors, execute template and return if >400 or err
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
		
		// Read the response body, should be JSON
		body, err := io.ReadAll(res.Body)
		if err != nil {
			http.Error(w, "Failed to read API response", http.StatusInternalServerError)
			log.Fatal(err)
		}

		var apiRes *Response

		// Parse data from the api and store into apiRes
		err = json.Unmarshal(body, &apiRes)
		if err != nil {
			http.Error(w, "Failed to parse API response", http.StatusInternalServerError)
			log.Fatal(err)
		}
		
		// Final WeatherPageData that will be given to the template
		pageData = WeatherPageData{apiRes, fmt.Sprintf("http://openweathermap.org/img/w/%s.png", apiRes.WeatherCondition[0].Icon), false}

		// Temperatures will likely have decimal values, so round them
		pageData.WeatherData.Main.roundTemps()
	}

	// Execute the template with the WeatherPageData
	err = template.Execute(w, pageData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Fatal(err)
	}
}

func main() {
	http.HandleFunc("/", handler)
	
	// Serves css files
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))

	fmt.Printf("Listening on port 8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

