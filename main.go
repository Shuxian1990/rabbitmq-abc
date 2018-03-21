package main

import (
	"github.com/streadway/amqp"
	"printfcoder.com/rabbitmq-abc/rabbitmq"
)

func main() {

	url := "amqp://it:its123@192.168.10.59:5672"
	connection, _ := amqp.Dial(url)
	defer connection.Close()

	// 测试direct
	rabbitmq.Direct(connection)

}
