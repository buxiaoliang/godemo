package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/gorilla/mux"
	"log"
	"fmt"
	"net/http"
	"encoding/json"
)

type Bookmark struct {
	Url   string `json:"url"`
	Title string `json:"title"`
	Tags  string `json:"tags"`
}

// rowScanner is implemented by sql.Row and sql.Rows
type rowScanner interface {
	Scan(dest ...interface{}) error
}

func main() {
	// router
	router := mux.NewRouter()
	router.HandleFunc("/bookmark", handleBookmark).Methods("GET", "POST")
	//bookmarks, err := getBooks()
	//if err != nil {
	//	log.Println(err.Error())
	//}
	//for i, v := range bookmarks {
	//	fmt.Println(i, ":", v.Title)
	//	fmt.Println(i, ":", v.Url)
	//	fmt.Println(i, ":", v.Tags)
	//}
	// attention: the port may been used by other process
	log.Fatal(http.ListenAndServe(":8084", router))
}

func handleBookmark(w http.ResponseWriter, r *http.Request) {
	fmt.Println("handleBookmark:")
	w.Header().Set("Content-Type", "application/json")

	bookmarks, err := getBooks()
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	outgoingJSON, error := json.Marshal(bookmarks)
	if error != nil {
		log.Println(error.Error())
		http.Error(w, error.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, string(outgoingJSON))
}

func getBooks() ([]*Bookmark, error) {
	fmt.Println("getBookmarks:")
	db, err := sql.Open("postgres", "user=postgres dbname=mozy_bookmark password=password sslmode=disable")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	rows, err := db.Query("SELECT url, title, tags FROM bookmark")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	defer rows.Close()
	defer db.Close()

	var bookmarks []*Bookmark
	for rows.Next() {
		bookmark, err := scanBookmark(rows)
		if err != nil {
			return nil, fmt.Errorf("mysql: could not read row: %v", err)
		}

		bookmarks = append(bookmarks, bookmark)
	}

	return bookmarks, nil
}

func scanBookmark(s rowScanner) (*Bookmark, error) {
	fmt.Println("getBookmark:")
	var bookmark Bookmark
	if err := s.Scan(&bookmark.Url, &bookmark.Title, &bookmark.Tags); err != nil {
		return nil, err
	}
	return &bookmark, nil
}