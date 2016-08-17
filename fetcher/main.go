package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/streadway/amqp"
)

func main() {
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

	go func() {
		for d := range msgs {
			job := Job{}
			err := json.Unmarshal(d.Body, &job)
			//TODO: validate and handle dropped messages
			d.Ack(false) // don't requeue messages
			if err != nil {
				log.Println("debug: Failed to unmarshal message body into job", err)
				continue
			}
			html, err := GetData(&job)
			if err != nil {
				log.Printf("debug: Failed to get data for job: %s query: %s\n%s", job.Id, job.Query, err)
				continue
			}
			fmt.Println(html)
			d.Ack(false)
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

// copying this around cause my environment is borked
type Job struct {
	Id    string `json:"id"`
	Query string `json:"query"`
}
