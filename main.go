package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/pborman/uuid"
)

func createWorker(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	id := uuid.New()
	log.Printf("Creating worker... %s\n", id)
	w.Write([]byte(id))
}

func fetchData(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	log.Printf("Fetching data for id: %s\n", id)
	job := Job{
		id:    id,
		query: "http://www.example.com",
	}
	html, err := GetData(&job)
	if err != nil {
		log.Printf("error: failed to get data: %s", err)
	}
	w.Write(html)
}

func main() {
	router := httprouter.New()
	router.POST("/api", createWorker)
	router.GET("/api/:id", fetchData)

	log.Fatal(http.ListenAndServe(":8080", router))
}

// worker

type Job struct {
	id    string
	query string
}

func GetData(j *Job) ([]byte, error) {
	resp, err := http.Get(j.query) // ahue
	if err != nil {
		return nil, fmt.Errorf("warn: Failed to get data from query: %s", err)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("warn: Failed to read response body: %s", err)
	}
	resp.Body.Close()

	return b, nil
}
