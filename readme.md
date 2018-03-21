# 1.3.1 节 RabbitMQ中Exchange的name 与 type

本节讲讲**Exchange**的**name**与**type**

其实脱离队列（channel、queue）讲type是没有意义的，但是type比较重要，Exchange推送消息给Queue受type控制，故而在讲Queue之间不得不先讲type。

# 1.3.1 name、type（kind)

## name
name 就是exchange的名称标识。rabbitmq安装好后会有以"amq."开头的几个预定义标准exchnage。client尽量避免和这些初始化的数据撞车，起一些和自己业务见名知义的name。也不要没事找事，重复定义相同的exchange。

## type
有些环境也称为kind，Exchange 控制了messages推送到queues的路由规则，type 就为messages如何被published到侦听了这个规则的queue定义了一些路由算法。主要有**direct**, **fanout**, **topic** 和 **headers**。

### direct
![](https://printfcoder.github.io/myblog/assets/images/rabbitmq/abc/exchange-direct.png)

direct路由规则要求queue绑定的**routingKey**与**producer**发送的message **routingKey完全一致**。这个特性也常用于对producer与consumer关系**强耦合**的环境中，也即peer2peer关系非常直接的环境。

下面我们用代码实现图中的场景，但我们也不需要去实现的队列，我们主要的目的是确认**direct**的特性--与**routingKey**完全一致才能收到消息，然后使用通配符“*”号来尝试是否能匹配任意字符。那么，我们可以定义三个队列：“green”，与green完全不同的“red”，与green半相似的带通配符的“gree*”。

代码：

```golang
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
```
执行，main函数见[最终代码][源码]：

```shell
$ go run main.go -type=direct

# 以下为输出，可见direct为全词匹配，*号不起匹配作用
接收到green: I'm green Timestamp: 2016-03-21 19:37:36 +0800 CST
接收到green: I'm green Timestamp: 2016-03-21 19:37:37 +0800 CST
接收到green: I'm green Timestamp: 2016-03-21 19:37:38 +0800 CST
接收到green: I'm green Timestamp: 2016-03-21 19:37:39 +0800 CST
```
如日志打印一样，direct为全词匹配，*号不起匹配作用。


### fanout

![](https://printfcoder.github.io/myblog/assets/images/rabbitmq/abc/exchange-direct.png)

**fanout**，可以理解为扇形，就是像扇子一样把消息散出去。它的规则比**direct**随意一些。如果exchange定义成**fanout**后，那么bind（绑定）了该exchange的队列都会收到消息。故而这个类型比较适合**生产者-消费者（或者事件驱动）**模型。

假设我们有如下场景：

1. 生产者（员工）发送了一个员工离职的message
2. 财务部（consumer_fi）,人事部（consumer_hr），用人部门（consumer_em）都收到了这个通知。那么:
3. 财务部（consumer_fi）就开始计算员工的工资。
4. 人事部（consumer_hr）将停止员工的门禁卡，并设置该员工的信息为只读，财务部（consumer_fi）也不能再对员工以后的工资条进行写操作。
5. 用人部门（consumer_em）将对员工工作进行转接。

注意：*需要特别指出的是， exchange type为 fanout中的routingKey是无效的*，也即是侦听了这个exchange的queue都会收到消息。如果一定要使用routingKey，可以考虑使用下两节中的*topic*或*headers*匹配。[exchange-fanout官方文档][exchange_fanout]

> A fanout exchange routes messages to all of the queues that are bound to it and the routing key is ignored.
If N queues are bound to a fanout exchange, when a new message is published to that exchange a copy of the 
message is delivered to all N queues. Fanout exchanges are ideal for the broadcast routing of messages.

解释一下
>  绑定到fanout的exchange上的队列都会收到消息，而队列的路由规则会被忽略。如果有N个队列都绑到这个exchange上，当新消息到达时，该消息的副本会被传送到全部N个队列上。Fanout类型适用于广播消息。

好，用代码实现：

```golang

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
```

运行：

```shell
$ go run main.go -type=faount
财务部冻结工资账户: A 员工离职了 Timestamp: 2016-03-21 20:00:01 +0800 CST
非法部门偷侦听: A 员工离职了 Timestamp: 2016-03-21 20:00:01 +0800 CST
人事部冻结OA账户: A 员工离职了 Timestamp: 2016-03-21 20:00:01 +0800 CST
用人部门转接任务: A 员工离职了 Timestamp: 2016-03-21 20:00:01 +0800 CST
```

## topic

![](https://printfcoder.github.io/myblog/assets/images/rabbitmq/abc/exchange-topic.png)

当queue的key全部或局部匹配message的key，则会收到该message。而匹配的词必须以"."号分开，比如，有一条消息以key="first.green.fast"发送给这个type=topic的exchange，那么key为以下的将会收到这个message：

1. *，单词通配符，不能匹配部分，如：*.green.fast，*.*.fast，*.green.*等将会收到，而*t.green.fast，*.red.fast将收不到
2. 全key通配符，如：#，#.fast，first.#等将会收到，而#t.green.fast，#.blue.fast将收不到

我们继续用第一个例子**direct**中的场景，只是修改为模糊匹配。定义三个队列：“*.green.”、“*.red.fast”、“*.*.fast”。发送路由为**first.green.fast**的消息，那么按topic的规则，“*.red.fast”将不会收到消息。

代码：

```golang
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
```

执行：
```shell
$ go run main.go -type=topic
*.green.*: I'm first.green.fast Timestamp: 2016-03-21 20:15:09 +0800 CST
*.*.fast: I'm first.green.fast Timestamp: 2016-03-21 20:15:09 +0800 CST
```

由于“*.green.”、“*.*.fast”都是模糊匹配**first.green.fast**，而“*.red.fast”不行，所以只打印前两者的两条消息。读者可以把“*”改为“#”，试一下效果。


## headers

![](https://printfcoder.github.io/myblog/assets/images/rabbitmq/abc/exchange-header.png)

headers好比http请求中的headers，queues中的headers如果匹配message中的headers，那将收到这个消息。和topic比较相似，不同的是，topic的是将criteria合并在一起的key且顺序是有讲究的，而headers则是将criteria展开为key-value不依赖顺序的方式去匹配。

在headers exchange中，是要全部匹配headers还是匹配一个key-value就行呢？可以通过在queue中设置x-match参数。如果设置为any，则代表匹配一个就可以了，如果是all，则要全部匹配。需要提醒的是，匹配不匹配取决于queue中的key-value是否能在message headers 中找到，message headers中有的，而queue中没有不影响结果。也即：queue去匹配message，而不是message去匹配queue。

而headers中是否可以用#，*等通配符，答案是不可以的，headers的匹配好比加强版的direct，要全词匹配，还是那句话，不要没事找事。

代码：
```golang
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
			"x-match": "all",
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
```

执行：
```shell
$ go run main.go -type=header
key1 匹配： I'm headers Timestamp: 2016-03-21 20:20:43 +0800 CST
key all 匹配： I'm headers Timestamp: 2016-03-21 20:20:43 +0800 CST
```

解读一下代码

```golang
  // consumer_key1
	// 匹配一个
	go func() {
		...
		headers := amqp.Table{
			"x-match": "any",
			"key1":    "value1",
			"key3":    "value3", // key3不存在message中，不影响any匹配
		}
		...
	}()
```
这里我们定义了一个匹配任意header的队列，key1是消息头中有的，而key3没有，但是key3即使不在消息头中，也不影响**any**操作符匹配。也就是说，**any**只要队列头定义中能找到一个在消息头中存在，那么exchange就会把消息传到这个队列当中。

下面这段

```golang
  // consumer_key_all
	// 匹配全部
	go func() {
	  ...
		headers := amqp.Table{
			"x-match": "all",
			"key2":    "value2", // key2，key1顺序不影响匹配
			"key1":    "value1",
		}
	  ...
	}()
```
定义一个队列拥有全部消息头属性时，可以收到，且头顺序不影响。也就是说，**all**要求所有的消息与队列的头部全匹配，exchange才会把消息推送到队列中。

现在我们再尝试，当队列的头比消息的头多时，会是什么情况：

```golang
  // consumer_key_no_in_all
	// 匹配全部，乱序
	go func() {
		...
		headers := amqp.Table{
			"x-match": "all",
			"key2":    "value2",
			"key1":    "value1",
			"key3":    "value3", // key3 在message中不存在，影响匹配
		}
	  ...
	}()
```
**key3**在消息头中是不存在的，从而影响了接收。所以，在**x-match=all**情况下，再次证明消息与队列头必须全部匹配，才能收到消息。

再看这段：

```golang
	// 尝试使用通配符
	go func() {
		...
		headers := amqp.Table{
			"x-match": "any",
			"key1":    "*",
		}
  	...
    }()
```
我们尝试使用通配符，看是否能工作。但是从日志中我们看到，并没有生效。


# 小章总结

本篇我们介绍了exchange的name与type。重点讲了type的作用规则，特别是4个类型direct、fanout、header、topic的用法与场景。希望大家有所收获。下一篇我们讲exchange的两个重要参数[durable, autoDelete][durable_autoDelete]

[源码]: https://github.com/printfcoder/rabbitmq-abc/tree/part_1
[exchange_fanout]: http://www.rabbitmq.com/tutorials/amqp-concepts.html#exchange-fanout
[durable_autoDelete]:https://printfcoder.github.io/myblog/mq/2016/01/21/rabbitmq-abc-part-1-3-2-how-to-use-durable-and-autoDelete/