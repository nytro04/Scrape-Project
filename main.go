package main

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/PuerkitoBio/goquery"
	"encoding/json"
	"strconv"
)

// Details is a struct which represents the the three unstructured
//  pieces of data from the Details field of a Movie object.
type Details struct {
	ReleaseDate string `json:"release_date"`
	Genre       string `json:"genre"`
	Language    string `json:"language"`
}

// Movie represents the metadata for one Movie.
type Movie struct {
	ID		 int	`json:"id"`
	Title    string `json:"title"`
	Duration string `json:"duration"`
	Details  Details `json:"details"`
	Genre    string `json:"genre"`
	ShowTime string `json:"showtime"`
	Votes    string `json:"votes"`
}

// getMovies is a common utility API which returns
// all available movies compatible with the source.
func getMovies() []Movie {
	var movies []Movie
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
	return movies
}

// getMovie will use the getMovies API in
// order to retrieve a single Movie object.
func getMovie(id int) Movie {
	movies := getMovies()
	var r Movie
	for _, movie := range movies {
		if movie.ID == id {
			r = movie
		}
	}
	return r
}

// GetMovies is a handler for the http mux.
// It will get a JSON payload containing
// all compatible Movie objects.
func GetMovies(w http.ResponseWriter, r *http.Request) {
	// Get the movies.
	movies := getMovies()
	// Convert the desired data into a json format.
	payload, err := json.Marshal(movies)
	if err != nil {
		log.Println(err)
	}
	// Return the payload.
	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)
}

// GetMovie is a handler for the http mux.
// It will get a single JSON payload containing
// one specified compatible Movie object.
func GetMovie(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.ParseInt(params["id"], 10, 64)
	if err != nil {
		log.Println(err)
	}
	movies := getMovies()
	var movie Movie
	for _, m := range movies {
		if string(m.ID) == string(id) {
			movie = m
		}
	}
	// Convert the desired data into a json format.
	payload, err := json.Marshal(movie)
	if err != nil {
		log.Println(err)
	}
	// Return the payload.
	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)
}

// RenderMovies will render the complete set of []Movies.
func RenderMovies(w http.ResponseWriter, r *http.Request) {
	data := getMovies()
	templ := template.Must(template.ParseFiles("template/index.gohtml"))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templ.Execute(w, data)
}

// RenderMovie will render one Movies.
func RenderMovie(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.ParseInt(params["id"], 10, 64)
	if err != nil {
		log.Println(err)
	}

	movie := getMovie(int(id))

	templ := template.Must(template.ParseFiles("template/details.gohtml"))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templ.Execute(w, movie)
}

// main is the program entry point.
func main() {
	router := mux.NewRouter()

	router.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("templates/assets"))))

	router.HandleFunc("/api/v1/movies", GetMovies).Methods("GET")
	router.HandleFunc("/api/v1/movies/{id}", GetMovie).Methods("GET")
	router.HandleFunc("/movies", RenderMovies).Methods("GET")
	router.HandleFunc("/movies/{id}", RenderMovie).Methods("GET")


	log.Println("Server started listening on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", router))
}
