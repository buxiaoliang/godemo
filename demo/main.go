package main

import (
	"encoding/json"
	"net/http"
	"fmt"
	"io/ioutil"
	"log"
	"github.com/gorilla/mux"
	"github.com/go-redis/redis"
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
	// router
	router := mux.NewRouter()
	router.HandleFunc("/location", handleLocationPost).Methods("GET", "POST")
	router.HandleFunc("/location/{name}", handleLocation).Methods("GET", "DELETE")
	log.Fatal(http.ListenAndServe(":8081", router))
}

func handleLocationPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	responseData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	var location Location
	json.Unmarshal(responseData, &location)

	switch r.Method {
	case "GET":
		// Serve the resource.
		fmt.Println("Endpoint Hit: GET Locations")
		// location to json string
		outgoingJSON, error := json.Marshal(DBClientGet())
		if error != nil {
			log.Println(error.Error())
			http.Error(w, error.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, string(outgoingJSON))
	case "POST":
		// Create a new record.
		fmt.Println("Endpoint Hit: POST Location By " + location.Name)
		// insert into database
		DBClientPost(location.Name)
		// location to json string
		outgoingJSON, error := json.Marshal(location)
		if error != nil {
			log.Println(error.Error())
			http.Error(w, error.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, string(outgoingJSON))
	default:
		// Give an error message.
		fmt.Println("Endpoint Hit: DEFAULT")
	}

	fmt.Println("Endpoint Hit: handleLocationPost")
}

func handleLocation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	name := vars["name"]

	fmt.Println("Request for:", name)

	switch r.Method {
	case "GET":
		if (DBClientCheck(name)) {
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
			// Serve the resource.
			fmt.Println("Endpoint Hit: GET Location By " + name)
			//fmt.Println(location.Name)
			//fmt.Println(len(location.Weather))
			// location to json string
			outgoingJSON, error := json.Marshal(location)

			if error != nil {
				log.Println(error.Error())
				http.Error(w, error.Error(), http.StatusInternalServerError)
				return
			}
			DBClient(name, string(outgoingJSON))
			fmt.Fprintf(w, string(outgoingJSON))
		} else {
			fmt.Fprintf(w, DBClientWeather(name))
		}
	case "DELETE":
		// Remove the record.
		fmt.Println("Endpoint Hit: DELETE Location By " + name)
		DBClientDelete(name)
	default:
		// Give an error message.
		fmt.Println("Endpoint Hit: DEFAULT")
	}
	fmt.Println("Endpoint Hit: handleLocation")
}

func DBClientGet() []string {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0, // use default DB
	})

	val, err := client.SMembers("locations").Result()
	if err != nil {
		panic(err)
	}
	//fmt.Println(val)
	return val
}

func DBClientPost(name string) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0, // use default DB
	})

	val, err := client.SIsMember("locations", name).Result()
	if err != nil {
		panic(err)
	}
	if (!val) {
		err2 := client.SAdd("locations", name).Err()
		if err2 != nil {
			panic(err2)
		}
	}
}

func DBClientDelete(name string) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0, // use default DB
	})

	val, err := client.SIsMember("locations", name).Result()
	if err != nil {
		panic(err)
	}
	if (val) {
		err2 := client.SRem("locations", name).Err()
		if err2 != nil {
			panic(err2)
		}
	}
}

// check if need to create/update location:name's weather
func DBClientCheck(name string) bool {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0, // use default DB
	})

	//expired, err := client.Expire("location:" + name, 3600).Result()
	//if err != nil {
	//	panic(err)
	//}
	//if (expired) {
	//	return false
	//}
	existed, err2 := client.Exists("location:" + name).Result()
	if err2 != nil {
		panic(err2)
	}
	return existed == 0
}

func DBClientWeather(name string) string {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0, // use default DB
	})

	weather, err := client.Get("location:" + name).Result()
	if err != nil {
		panic(err)
	}
	return weather
}

func DBClient(name string, json string) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0, // use default DB
	})

	err := client.Set("location:" + name, json, 0).Err()
	if err != nil {
		panic(err)
	}
}
