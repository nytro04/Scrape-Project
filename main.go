package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/PuerkitoBio/goquery"
	"encoding/json"
	"io/ioutil"
	"strconv"
)

type Details struct {
	ReleaseDate string `json:"release_date"`
	Genre       string `json:"genre"`
	Language    string `json:"language"`
}

type Movie struct {
	ID		int	`json:"id"`
	Title    string `json:"title"`
	Duration string `json:"duration"`
	Details  Details `json:"details"`
	Genre    string `json:"genre"`
	ShowTime string `json:"showtime"`
	Votes    string `json:"votes"`
}

type MovieStore struct {
	Films []*Movie
}

var movies []Movie

func GetMovies(w http.ResponseWriter, r *http.Request) {
	var source = "https://silverbirdcinemas.com/cinema/accra/"

	// Request the HTML page.
	res, err := http.Get(source)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Find the review items
	doc.Find("#cinema-m .entry-item").Each(func(i int, s *goquery.Selection) {
		// For each item found
		movie := Movie{
			i + 1,
			s.Find(".entry-title a").Text(),
			s.Find(".entry-date").Text(),
			Details{
				s.Find(".desc-mv div:nth-child(1)").Text(),
				s.Find(".desc-mv div.note").Text(),
				s.Find(".desc-mv div:nth-child(3)").Text(),
			},
			s.Find(".note a").Text(),
			s.Find(".cinema_page_showtime strong").Text(),
			s.Find(".entry-rating .rate").Text(),
		}

		// Any additional filters.
		movie.Details.ReleaseDate = strings.Replace(movie.Details.ReleaseDate, "Release:", "", -1)
		movie.Details.Genre = strings.Replace(movie.Details.Genre, "Genre:", "", -1)
		movie.Details.Language = strings.Replace(movie.Details.Language, "Language:", "", -1)

		movies = append(movies, movie)
	})

	// Convert the desired data into a json format.
	payload, err := json.Marshal(movies)
	if err != nil {
		log.Println(err)
	}

	// Return the payload.
	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)
}

func RenderMovieList(w http.ResponseWriter, r *http.Request) {
	url := "http://localhost:8000/api/v1/movies"
	res, err := http.Get(url)

	if err != nil {
		panic(err.Error())
	}

	body, err := ioutil.ReadAll(res.Body)

	var data []Movie
	json.Unmarshal(body, &data)

	templ := template.Must(template.ParseFiles("template/index.gohtml"))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templ.Execute(w, data)
}

func GetSingleMovie(w http.ResponseWriter, r *http.Request) {
	url := "http://localhost:8000/api/v1/movies"
	res, err := http.Get(url)

	if err != nil {
		panic(err.Error())
	}

	body, err := ioutil.ReadAll(res.Body)

	var data []Movie
	json.Unmarshal(body, &data)

	params := mux.Vars(r)
	id, err := strconv.ParseInt(params["id"], 10, 64)
	if err != nil {
		log.Println(err)
	}

	var result Movie
	for _, movie := range data {
		if int64(movie.ID) == id {
			result = movie
			break
		}
	}

	fmt.Println(result)

	templ := template.Must(template.ParseFiles("template/details.gohtml"))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templ.Execute(w, result)
}


func main() {
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/movies", GetMovies).Methods("GET")
	router.HandleFunc("/movies", RenderMovieList).Methods("GET")
	router.HandleFunc("/movies/{id}", GetSingleMovie).Methods("GET")


	log.Println("Server started listening on port...")
	log.Fatal(http.ListenAndServe(":8000", router))
}

