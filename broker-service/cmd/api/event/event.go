package event

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func declareExchange(channel *amqp.Channel) error {
	// exchange: 메시지가 도착했을 때 어떤 큐로 보내질지를 결정하는 라우팅 규칙을 정의
	return channel.ExchangeDeclare(
		"logs_topic", // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
}

func declareRandomQueue(channel *amqp.Channel) (amqp.Queue, error) {
	return channel.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
}
