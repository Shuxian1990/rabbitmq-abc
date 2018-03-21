package rabbitmq

import (
	"fmt"
	"time"

	"github.com/streadway/amqp"
)

// Header 演示type为Header
func Header(connection *amqp.Connection) {
	// consumer_key1
	// 匹配一个
	go func() {
		channel, _ := connection.Channel()
		defer channel.Close()
		headers := amqp.Table{
			"x-match": "any",
			"key1":    "value1",
			"key3":    "value3", // key3不存在message中，不影响any匹配
		}
		q, _ := channel.QueueDeclare("consumer_key1_queue", false, true, false, true, nil)
		channel.QueueBind(q.Name, "", "amq.headers", false, headers)

		messages, _ := channel.Consume(q.Name, "consumer_key1", false, false, false, false, nil)
		for msg := range messages {
			fmt.Println(`key1 匹配：`, string(msg.Body), "Timestamp:", msg.Timestamp)
		}
	}()

	// consumer_key_all
	// 匹配全部
	go func() {
		channel, _ := connection.Channel()
		defer channel.Close()
		headers := amqp.Table{
			"x-match": "all",
			"key2":    "value2", // key2，key1顺序不影响匹配
			"key1":    "value1",
		}
		q, _ := channel.QueueDeclare("consumer_key_all_queue", false, true, false, true, nil)
		channel.QueueBind(q.Name, "", "amq.headers", false, headers)

		messages, _ := channel.Consume(q.Name, "consumer_key_all", false, false, false, false, nil)
		for msg := range messages {
			fmt.Println(`key all 匹配：`, string(msg.Body), "Timestamp:", msg.Timestamp)
		}
	}()

	// consumer_key_no_in_all
	// 匹配全部，乱序
	go func() {
		channel, _ := connection.Channel()
		defer channel.Close()
		headers := amqp.Table{
			"x-match": "all",
			"key2":    "value2",
			"key1":    "value1",
			"key3":    "value3", // key3 在message中不存在，影响匹配
		}
		q, _ := channel.QueueDeclare("consumer_key_no_in_all_queue", false, true, false, true, nil)
		channel.QueueBind(q.Name, "", "amq.headers", false, headers)

		messages, _ := channel.Consume(q.Name, "consumer_key_no_in_all", false, false, false, false, nil)
		for msg := range messages {
			fmt.Println(`key3 不匹配：`, string(msg.Body), "Timestamp:", msg.Timestamp)
		}
	}()

	// consumer_key_wildcard
	// 尝试使用通配符
	go func() {
		channel, _ := connection.Channel()
		defer channel.Close()
		headers := amqp.Table{
			"x-match": "any",
			"key1":    "*",
		}
		q, _ := channel.QueueDeclare("consumer_key_wildcard_queue", false, true, false, true, nil)
		channel.QueueBind(q.Name, "", "amq.headers", false, headers)

		messages, _ := channel.Consume(q.Name, "consumer_key_wildcard", false, false, false, false, nil)
		for msg := range messages {
			fmt.Println(`key wildcard 匹配：`, string(msg.Body), "Timestamp:", msg.Timestamp)
		}
	}()

	go func() {
		timer := time.NewTicker(1 * time.Second)
		channel, _ := connection.Channel()

		headers := amqp.Table{
			"key1": "value1",
			"key2": "value2",
		}
		for t := range timer.C {
			msg := amqp.Publishing{
				DeliveryMode: 1,
				Timestamp:    t,
				ContentType:  "text/plain",
				Body:         []byte("I'm headers"),
				Headers:      headers,
			}
			channel.Publish("amq.headers", "", false, false, msg)
		}
	}()

	select {}
}
