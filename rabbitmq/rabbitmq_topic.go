package rabbitmq

import (
	"fmt"
	"time"

	"github.com/streadway/amqp"
)

// Topic 演示type为Topic
func Topic(connection *amqp.Connection) {
	// consumer_green
	go func() {
		channel, _ := connection.Channel()
		defer channel.Close()

		q, _ := channel.QueueDeclare("consumer_green_queue", false, true, false, true, nil)
		channel.QueueBind(q.Name, "*.green.*", "amq.topic", false, nil)
		messages, _ := channel.Consume(q.Name, "consumer_green", false, false, false, false, nil)
		for msg := range messages {
			fmt.Println("*.green.*:", string(msg.Body), "Timestamp:", msg.Timestamp)
		}

	}()

	// consumer_red_fast
	go func() {
		channel, _ := connection.Channel()
		defer channel.Close()
		q, _ := channel.QueueDeclare("consumer_red_fast_queue", false, true, false, true, nil)
		channel.QueueBind(q.Name, "*.red.fast", "amq.topic", false, nil)
		messages, _ := channel.Consume(q.Name, "consumer_red_fast", false, false, false, false, nil)
		for msg := range messages {
			fmt.Println("*.red.fast:", string(msg.Body), "Timestamp:", msg.Timestamp)
		}

	}()

	// consumer_all_fast
	go func() {
		channel, _ := connection.Channel()
		defer channel.Close()
		q, _ := channel.QueueDeclare("consumer_all_fast_queue", false, true, false, true, nil)
		channel.QueueBind(q.Name, "*.*.fast", "amq.topic", false, nil)
		messages, _ := channel.Consume(q.Name, "consumer_all_fast", false, false, false, false, nil)
		for msg := range messages {
			fmt.Println("*.*.fast:", string(msg.Body), "Timestamp:", msg.Timestamp)
		}

	}()

	// consumer_some_green_queue
	// 尝试*是否能匹配半个单词（答案是不能）
	go func() {
		channel, _ := connection.Channel()
		defer channel.Close()
		q, _ := channel.QueueDeclare("consumer_some_green_queue", false, true, false, true, nil)
		// 使用通配符
		channel.QueueBind(q.Name, "*t.green.fast", "amq.topic", false, nil)
		messages, _ := channel.Consume(q.Name, "consumer_some_green", false, false, false, false, nil)
		for msg := range messages {
			fmt.Println("*t.green.fast:", string(msg.Body), "Timestamp:", msg.Timestamp)
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
				Body:         []byte("I'm first.green.fast"),
			}
			channel.Publish("amq.topic", "first.green.fast", false, false, msg)
		}
	}()

	select {}
}
