package rabbitmq

import (
	"fmt"

	"github.com/streadway/amqp"

	"time"
)

// Direct 演示type为direct
func Direct(connection *amqp.Connection) {

	// 侦听green
	go func() {
		channel, _ := connection.Channel()
		defer channel.Close()

		q, _ := channel.QueueDeclare("green", false, true, false, true, nil)
		channel.QueueBind(q.Name, "green", "amq.direct", false, nil)
		messages, _ := channel.Consume(q.Name, "consumer0", false, false, false, false, nil)
		for msg := range messages {
			fmt.Println("接收到green:", string(msg.Body), "Timestamp:", msg.Timestamp)
		}

	}()

	// 侦听orange
	go func() {
		channel, _ := connection.Channel()
		defer channel.Close()
		q, _ := channel.QueueDeclare("orange", false, true, false, true, nil)
		channel.QueueBind(q.Name, "orange", "amq.direct", false, nil)
		messages, _ := channel.Consume(q.Name, "consumer1", false, false, false, false, nil)
		for msg := range messages {
			fmt.Println("接收到orange:", string(msg.Body), "Timestamp:", msg.Timestamp)
		}

	}()

	// 侦听gree*
	go func() {
		channel, _ := connection.Channel()
		defer channel.Close()
		q, _ := channel.QueueDeclare("gree*", false, true, false, true, nil)
		channel.QueueBind(q.Name, "gree*", "amq.direct", false, nil)
		messages, _ := channel.Consume(q.Name, "consumer1", false, false, false, false, nil)
		for msg := range messages {
			fmt.Println("接收到gree*:", string(msg.Body), "Timestamp:", msg.Timestamp)
		}

	}()

	go func() {
		timer := time.NewTicker(1 * time.Second)
		channel, _ := connection.Channel()

		for t := range timer.C {
			msg := amqp.Publishing{
				DeliveryMode: 1,
				Timestamp:    t,
				ContentType:  "text/plain",
				Body:         []byte("I'm green"),
			}
			channel.Publish("amq.direct", "green", false, false, msg)
		}
	}()

	select {}
}
