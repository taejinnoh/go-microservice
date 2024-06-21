package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn      *amqp.Connection
	queueName string
}

func NewConsumer(conn *amqp.Connection) (Consumer, error) {
	consusmer := Consumer{
		conn: conn,
	}

	err := consusmer.setup()
	if err != nil {
		return Consumer{}, err
	}

	return consusmer, nil
}

func (c *Consumer) setup() error {
	channel, err := c.conn.Channel()
	if err != nil {
		return err
	}

	return declareExchange(channel)
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (c *Consumer) Listen(topics []string) error {
	channel, err := c.conn.Channel()
	if err != nil {
		return err
	}

	defer channel.Close()

	q, err := declareRandomQueue(channel)
	if err != nil {
		return err
	}

	for _, s := range topics {
		// QueueBind 함수는 특정 큐를 Exchange에 바인딩
		err = channel.QueueBind(
			q.Name,       // queue name
			s,            // routing key
			"logs_topic", // exchange name
			false,        // no-wait
			nil,          // arguments
		)
		if err != nil {
			return err
		}
	}

	messages, err := channel.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // arguments
	)
	if err != nil {
		return err
	}

	forever := make(chan bool)
	go func() {
		for d := range messages {
			var payload Payload
			_ = json.Unmarshal(d.Body, &payload)

			go handlePayload(payload)
		}
	}()

	fmt.Printf("waiting for message [Exchange, Queue] [logs_topic, %s]\n", q.Name)
	<-forever

	return nil
}

func handlePayload(paylaod Payload) {
	switch paylaod.Name {
	case "log", "event":
		err := logEvent(paylaod)
		if err != nil {
			log.Println(err)
		}
	case "auth":
	default:
		err := logEvent(paylaod)
		if err != nil {
			log.Println(err)
		}
	}
}

func logEvent(payload Payload) error {
	jsonData, _ := json.MarshalIndent(payload, "", "\t")

	logServiceURL := "http://logger-service:8080/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		return err
	}

	return nil
}
