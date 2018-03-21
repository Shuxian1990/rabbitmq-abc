package rabbitmq

import (
	"fmt"
	"time"

	"github.com/streadway/amqp"
)

// Fanout 演示type为Fanout
func Fanout(connection *amqp.Connection) {
	// 财务部（consumer_fi）
	go func() {
		channel, _ := connection.Channel()
		defer channel.Close()

		q, _ := channel.QueueDeclare("consumer_fi_queue", false, true, false, true, nil)
		channel.QueueBind(q.Name, "employee_leave", "amq.fanout", false, nil)
		messages, _ := channel.Consume(q.Name, "consumer_fi", false, false, false, false, nil)
		for msg := range messages {
			fmt.Println("财务部冻结工资账户:", string(msg.Body), "Timestamp:", msg.Timestamp)
		}

	}()

	// 人事部（consumer_hr）
	go func() {
		channel, _ := connection.Channel()
		defer channel.Close()
		q, _ := channel.QueueDeclare("consumer_hr_queue", false, true, false, true, nil)
		channel.QueueBind(q.Name, "employee_leave", "amq.fanout", false, nil)
		messages, _ := channel.Consume(q.Name, "consumer_hr", false, false, false, false, nil)
		for msg := range messages {
			fmt.Println("人事部冻结OA账户:", string(msg.Body), "Timestamp:", msg.Timestamp)
		}

	}()

	// 用人部门（consumer_em）
	go func() {
		channel, _ := connection.Channel()
		defer channel.Close()
		q, _ := channel.QueueDeclare("consumer_em_queue", false, true, false, true, nil)
		channel.QueueBind(q.Name, "employee_leave", "amq.fanout", false, nil)
		messages, _ := channel.Consume(q.Name, "consumer_em", false, false, false, false, nil)
		for msg := range messages {
			fmt.Println("用人部门转接任务:", string(msg.Body), "Timestamp:", msg.Timestamp)
		}

	}()

	// 非法部门（consumer_illegal）
	go func() {
		channel, _ := connection.Channel()
		defer channel.Close()
		q, _ := channel.QueueDeclare("consumer_illegal_queue", false, true, false, true, nil)
		// 使用通配符
		channel.QueueBind(q.Name, "someKey*", "amq.fanout", false, nil)
		messages, _ := channel.Consume(q.Name, "consumer_illegal", false, false, false, false, nil)
		for msg := range messages {
			fmt.Println("非法部门偷侦听:", string(msg.Body), "Timestamp:", msg.Timestamp)
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
				Body:         []byte("A 员工离职了"),
			}
			channel.Publish("amq.fanout", "employee_leave", false, false, msg)
		}
	}()

	select {}
}
