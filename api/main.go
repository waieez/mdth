package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/julienschmidt/httprouter"
	"github.com/pborman/uuid"
	"github.com/streadway/amqp"
	redis "gopkg.in/redis.v4"
)

func main() {
	router := httprouter.New()

	// Handlers
	router.POST("/api", createWorker) // ?query=http://www.google.com
	router.GET("/api/:id", fetchData)

	log.Fatal(http.ListenAndServe(":8080", router))
}

// API Handlers
func createWorker(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// TODO: pass a client around instead of recreating it every time
	cache, err := NewRedisClient()
	if err != nil {
		log.Printf("error: Failed to create a connection to Redis")
		w.WriteHeader(500)
		return
	}

	// Create Job
	id := uuid.New()
	log.Printf("info: Creating worker... %s\n", id)
	query := r.FormValue("query")
	// TODO: validation, for now assume scheme is passed in as well
	unescaped, err := url.QueryUnescape(query)
	if err != nil {
		log.Printf("Failed to unescape query: %s\n", unescaped)
		return
	}
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

	// Insert into Redis, "Processing"
	err = cache.Set(id, fmt.Sprintf("Processing url: %s for job: %s", unescaped, id), 0).Err()
	if err != nil {
		fmt.Printf("debug: Failed to set value in cache", err)
		w.WriteHeader(500)
		return
	}

	// Respond with UUID
	w.Write([]byte(id))
}

func fetchData(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// TODO: pass a client around instead of recreating it every time
	cache, err := NewRedisClient()
	if err != nil {
		log.Printf("error: Failed to create a connection to Redis")
		w.WriteHeader(500)
		return
	}

	// Get HTML/Status out of Redis
	id := ps.ByName("id")
	result, err := cache.Get(id).Result()
	if err != nil {
		log.Printf("debug: Failed to get value for id %s: %s", id, err)
		w.WriteHeader(500)
		return
	}

	// Respond with HTML/Status
	w.Write([]byte(result))
}

/// WORKER
type Job struct {
	Id    string `json:"id"`
	Query string `json:"query"`
}

func CreateJob(j Job) error {

	// TODO: Pull rabbit publisher out of CreateJob pass connection around instead
	////////
	conn, err := amqp.Dial(os.Getenv("RABBIT"))
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
	//////

	// Insert message into RabbitMQ
	b, err := json.Marshal(j)
	if err != nil {
		return fmt.Errorf("warn: Failed to marshal job into bytes %s", err)
	}
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

/// REDIS
func NewRedisClient() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS"),
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	_, err := client.Ping().Result()
	if err != nil {
		return nil, fmt.Errorf("error: Error while pinging Redis", err)
	}
	return client, nil
}
