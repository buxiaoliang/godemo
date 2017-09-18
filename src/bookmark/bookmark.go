package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"github.com/gorilla/mux"
	"log"
	"fmt"
	"net/http"
	"encoding/json"
	"io/ioutil"
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

const insertStatement = `
  INSERT INTO bookmark (
    url, title, tags
  ) VALUES ($1, $2, $3)`

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
	handler := cors.Default().Handler(router)
	log.Fatal(http.ListenAndServe(":8084", handler))
}

func handleBookmark(w http.ResponseWriter, r *http.Request) {
	fmt.Println("handleBookmark:")
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case "GET":
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
	case "POST":
		requestData, err := ioutil.ReadAll(r.Body)
		fmt.Println("POST of requestData: " + B2S(requestData))
		if err != nil {
			log.Fatal(err)
		}

		var bookmark Bookmark
		json.Unmarshal(requestData, &bookmark)

		err = AddBookmark(&bookmark);
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, string(requestData))
	default:
		// Give an error message.
		fmt.Println("handleBookmark: DEFAULT")
	}

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
			return nil, fmt.Errorf("postgres: could not read row: %v", err)
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

// AddBookmark saves a given bookmark, assigning it a new ID.
func AddBookmark(b *Bookmark) error {
	fmt.Println("AddBookmark:")
	db, err := sql.Open("postgres", "user=postgres dbname=mozy_bookmark password=password sslmode=disable")
	if err != nil {
		log.Fatal(err)
		return err
	}

	var insert *sql.Stmt
	if insert, err = db.Prepare(insertStatement); err != nil {
		return fmt.Errorf("postgres: prepare insert: %v", err)
	}

	err = execAffectingOneRow(insert, b.Url, b.Title, b.Tags)
	if err != nil {
		return err
	}
	return nil
}

// execAffectingOneRow executes a given statement, expecting one row to be affected.
func execAffectingOneRow(stmt *sql.Stmt, args ...interface{}) error {
	r, err := stmt.Exec(args...)
	if err != nil {
		return fmt.Errorf("postgres: could not execute statement: %v", err)
	}
	rowsAffected, err := r.RowsAffected()
	if err != nil {
		return fmt.Errorf("postgres: could not get rows affected: %v", err)
	} else if rowsAffected != 1 {
		return fmt.Errorf("postgres: expected 1 row affected, got %d", rowsAffected)
	}
	return nil
}

func B2S(bs []byte) string {
	b := make([]byte, len(bs))
	for i, v := range bs {
		b[i] = byte(v)
	}
	return string(b)
}