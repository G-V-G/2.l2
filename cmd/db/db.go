package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/G-V-G/2.l2/datastore"
)

type InData struct {
	Value string
}

type OutData struct {
	Key   string
	Value string
}

const port string = "8091"
const path string = "./out/storage/"

func handler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[len("/db/"):]
	const sizeBytes int64 = datastore.MaxFileSizeMb * 1024 * 1024
	db, err := datastore.NewDb("out/storage/", sizeBytes)
	if err != nil {
		panic(err)
	}
	var c InData
	if r.Method == "POST" {
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(body, &c)
		if err != nil {
			panic(err)
		}
		err = db.Put(key, c.Value)
		if err != nil {
			if err.Error() == "ErrHashSums" {
				http.Error(w, "{}", http.StatusUnprocessableEntity)
				return
			}
			panic(err)
		}
	}
	if r.Method == "GET" {
		value, err := db.Get(key)
		if err != nil {
			if err.Error() == "record does not exist" {
				http.Error(w, "{}", http.StatusNotFound)
				return
			}
			panic(err)
		}
		c.Value = value
		res, err := json.Marshal(&c)
		if err != nil {
			log.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	}
}

func main() {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		e := os.MkdirAll(path, os.ModePerm)
		if e != nil {
			panic(e)
		}
	}
	http.HandleFunc("/db/", handler)
	log.Printf("Starting server on " + port + " port...")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
