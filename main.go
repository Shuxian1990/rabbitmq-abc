package main

import (
	"flag"

	"github.com/streadway/amqp"
	"printfcoder.com/rabbitmq-abc/rabbitmq"
)

func main() {

	url := "amqp://it:its123@192.168.10.59:5672"
	connection, _ := amqp.Dial(url)
	defer connection.Close()

	tp := flag.String("type", "", "fanout类型")
	flag.Parse()

	// 测试direct
	if *tp == "direct" {
		rabbitmq.Direct(connection)
	}

	// 测试fanout
	if *tp == "fanout" {
		rabbitmq.Fanout(connection)
	}

	// 测试topic
	if *tp == "topic" {
		rabbitmq.Topic(connection)
	}

	// 测试header
	if *tp == "header" {
		rabbitmq.Header(connection)
	}
}
