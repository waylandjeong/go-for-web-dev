package main

import (
	"fmt"
	"net/http"
	"html/template"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"encoding/json"
	"net/url"
	"io/ioutil"
	"encoding/xml"
)

type Page struct {
	Name     string
	DBStatus bool
}

type SearchResult struct {
	Title  string	`xml:"title,attr"`
	Author string	`xml:"author,attr"`
	Year   string	`xml:"hyr,attr"`
	ID     string	`xml:"owi,attr"`
}

func main() {

	fmt.Println("Starting web server ....")

	templates := template.Must(template.ParseFiles("templates/index.html"))

	db, _ := sql.Open("sqlite3", "dev.db")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		p := Page{Name: "Gopher", DBStatus: false}

		if name := r.FormValue("name"); name != "" {
			p.Name = name
		}

		p.DBStatus = db.Ping() == nil

		if err := templates.ExecuteTemplate(w, "index.html", p); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		db.Close()
	})

	// Adding comment here
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		//results := []SearchResult{
		//	SearchResult{"Moby-Dick", "Herman Melville", "1851", "22222"},
		//	SearchResult{"The Adventures of Huckleberry Fin", "Mark Twain", "1854", "44444"},
		//	SearchResult{"The Catcher in the Rye", "JD Salinger", "1951", "33333"},
		//}

		// call search function
		var results []SearchResult
		var err error
		if results, err = search(r.FormValue("search")); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		encoder := json.NewEncoder(w)
		if err := encoder.Encode(results); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/books/add", func(w http.ResponseWriter, r *http.Request) {
	})

	fmt.Println(http.ListenAndServe(":8080", nil))

}

type ClassifySearchResponse struct {
	Results []SearchResult	`xml:"works>work"`
}

// Search function
func search(query string) ([]SearchResult, error) {
	var resp *http.Response
	var err error

	if resp, err = http.Get("http://classify.oclc.org/classify2/Classify?&summary=true&title=" + url.QueryEscape(query)); err != nil {
		return []SearchResult{}, err
	}

	defer resp.Body.Close()
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return []SearchResult{}, err
	}

	var c ClassifySearchResponse
	err = xml.Unmarshal(body, &c)
	return c.Results, err
}


