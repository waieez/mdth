package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/streadway/amqp"
	redis "gopkg.in/redis.v4"
)

func main() {
	/////// Set up Redis/Rabbit Connections
	cache, err := NewRedisClient()
	if err != nil {
		log.Fatalf("fatal: Failed to create a connection to Redis")
		return
	}
	conn, err := amqp.Dial(os.Getenv("RABBIT"))
	if err != nil {
		log.Fatalf("fatal: Failed to connect to RabbitMQ: %s", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("fatal: Failed to open a channel: %s", err)
		return
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
	if err != nil {
		log.Fatalf("fatal: Failed to declare a queue: %s", err)
	}

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		log.Fatalf("fatal: Failed to open a channel: %s", err)
		return
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("Fatal: Failed to register a consumer")
	}
	forever := make(chan bool)

	//// consume messages off 'job' queue
	go func() {
		for d := range msgs {
			job := Job{}
			err := json.Unmarshal(d.Body, &job)
			//TODO: validate and handle dropped messages
			d.Ack(false) // don't requeue messages
			if err != nil {
				log.Printf("debug: Failed to unmarshal message body into job: %s\n", err)
				continue
			}

			// Fetch data from url
			html, err := GetData(&job)
			if err != nil {
				log.Printf("debug: Failed to get data for job: %s query: %s\n%s", job.Id, job.Query, err)
				continue
			}

			// Insert into Redis
			err = cache.Set(job.Id, html, 0).Err()
			if err != nil {
				fmt.Printf("debug: Failed to set value in cache", err)
				continue
			}
			log.Printf("info: added data to cache for job: %s\n", job.Id)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func GetData(j *Job) (string, error) {
	resp, err := http.Get(j.Query) // hehe
	if err != nil {
		return "", fmt.Errorf("warn: Failed to get data from query: %s", err)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("warn: Failed to read response body: %s", err)
	}
	resp.Body.Close()

	return string(b), nil
}

// WORKER
type Job struct {
	Id    string `json:"id"`
	Query string `json:"query"`
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
