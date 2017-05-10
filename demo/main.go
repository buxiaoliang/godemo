package main

import (
	"encoding/json"
	"net/http"
	"fmt"
	"io/ioutil"
	"log"
	"github.com/gorilla/mux"
)

type Location struct {
	Name    string `json:"name"`
	Weather []Weather `json:"weather"`
}

type Weather struct {
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Id          int `json:"id"`
	Main        string `json:"main"`
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/location/{name}", handleLocation).Methods("GET", "DELETE");
	log.Fatal(http.ListenAndServe(":8081", router))
}

func handleLocation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	name := vars["name"]

	fmt.Println("Request for:", name)
	// get data from rest api
	response, err := http.Get("http://api.openweathermap.org/data/2.5/weather?q=" + name + "&APPID=3a730068fddcec295e6ea1e29b342167")
	if err != nil {
		fmt.Print(err.Error())
		//os.Exit(1)
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var location Location
	json.Unmarshal(responseData, &location)

	switch r.Method {
	case "GET":
		// Serve the resource.
		fmt.Println("Endpoint Hit: GET")
		fmt.Println(location.Name)
		fmt.Println(len(location.Weather))
	case "POST":
		// Create a new record.
		fmt.Println("Endpoint Hit: POST")
	case "PUT":
		// Update an existing record.
		fmt.Println("Endpoint Hit: PUT")
	case "DELETE":
		// Remove the record.
		fmt.Println("Endpoint Hit: DELETE")
	default:
		// Give an error message.
		fmt.Println("Endpoint Hit: DEFAULT")
	}
	// location to json string
	outgoingJSON, error := json.Marshal(location)

	if error != nil {
		log.Println(error.Error())
		http.Error(w, error.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, string(outgoingJSON))
	fmt.Println("Endpoint Hit: location")
}
