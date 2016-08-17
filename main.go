package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"encoding/json"

	"github.com/julienschmidt/httprouter"
	"github.com/pborman/uuid"
	"github.com/streadway/amqp"
)

func main() {
	router := httprouter.New()
	router.POST("/api", createWorker) // ?query=http://www.google.com
	router.GET("/api/:id", fetchData)

	log.Fatal(http.ListenAndServe(":8080", router))
}

func createWorker(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	id := uuid.New()
	log.Printf("Creating worker... %s\n", id)
	query := r.FormValue("query")

	// TODO: validation, for now assume scheme is passed in as well

	unescaped, err := url.QueryUnescape(query)
	if err != nil {
		log.Printf("Failed to unescape query: %s\n", unescaped)
		return
	}
	// put stuff into rabbitmq
	j := Job{
		Id:    id,
		Query: unescaped,
	}
	err = CreateJob(j)
	if err != nil {
		log.Printf("Failed to create job %s\n", err)
		w.WriteHeader(500)
		return
	}
	w.Write([]byte(id))
}

func fetchData(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	// get stuff out of redis
	log.Printf("Fetching data for id: %s\n", id)
	job := Job{
		Id:    id,
		Query: "http://www.example.com",
	}
	html, err := GetData(&job)
	if err != nil {
		log.Printf("error: failed to get data: %s", err)
	}
	w.Write(html)
}

type Job struct {
	Id    string `json:"id"`
	Query string `json:"query"`
}

func GetData(j *Job) ([]byte, error) {
	resp, err := http.Get(j.Query) // ahue
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

func CreateJob(j Job) error {
	// marshal into bytes
	b, err := json.Marshal(j)
	if err != nil {
		return fmt.Errorf("warn: Failed to marshal job into bytes %s", err)
	}
	fmt.Printf("converted job to bytes: %s\n", string(b))

	conn, err := amqp.Dial("amqp://" + os.Getenv("RABBIT"))
	if err != nil {
		return fmt.Errorf("error: Failed to create a connection to rabbit %s", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("error: Failed to open a channel")
	}
	defer ch.Close()
	q, err := ch.QueueDeclare(
		"jobs", // name
		true,   // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(b),
		})
	if err != nil {
		return fmt.Errorf("Failed to publish a message: %s", err)
	}
	log.Printf(" [x] Sent %s", b)
	return nil
}
