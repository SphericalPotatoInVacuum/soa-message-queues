package queue

import (
	"github.com/SphericalPotatoInVacuum/soa-message-queues/internal/utils"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

var failOnError = utils.FailOnError

type Connection struct {
	conn         *amqp.Connection
	ch           *amqp.Channel
	grabberQueue amqp.Queue
	resultQueue  amqp.Queue
}

func NewConnection(addr string) *Connection {
	// establish connection to the RabbitMQ
	conn, err := amqp.Dial(addr)
	failOnError(err, "Failed to connect to RabbitMQ")
	log.Infof("Connected to %s", addr)

	// open channel
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	log.Info("Opened a channel")

	// declare queues
	grabberQueue, err := ch.QueueDeclare(
		"grabber", // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")
	log.Infof("Declared a %s queue", grabberQueue.Name)

	resultQueue, err := ch.QueueDeclare(
		"result", // name
		false,    // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare a queue")
	log.Infof("Declared a %s queue", resultQueue.Name)

	return &Connection{
		conn:         conn,
		ch:           ch,
		grabberQueue: grabberQueue,
		resultQueue:  resultQueue,
	}
}

func (c *Connection) Destroy() {
	c.ch.Close()
	c.conn.Close()
}

func (c *Connection) GrabberProduce(body []byte) error {
	return c.ch.Publish(
		"",                  // exchange
		c.grabberQueue.Name, // routing key
		false,               // mandatory
		false,               // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         body,
		},
	)
}

func (c *Connection) NewGrabberConsumer() <-chan amqp.Delivery {
	msgs, err := c.ch.Consume(
		c.grabberQueue.Name, // queue
		"",                  // consumer
		false,               // auto-ack
		false,               // exclusive
		false,               // no-local
		false,               // no-wait
		nil,                 // args
	)
	failOnError(err, "Failed to register a consumer")
	log.Info("Registered a grabber consumer")
	return msgs
}

func (c *Connection) ResultProduce(body []byte) error {
	return c.ch.Publish(
		"",                 // exchange
		c.resultQueue.Name, // routing key
		false,              // mandatory
		false,              // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         body,
		},
	)
}

func (c *Connection) NewResultConsumer() <-chan amqp.Delivery {
	msgs, err := c.ch.Consume(
		c.resultQueue.Name, // queue
		"",                 // consumer
		true,               // auto-ack
		false,              // exclusive
		false,              // no-local
		false,              // no-wait
		nil,                // args
	)
	failOnError(err, "Failed to register a consumer")
	log.Info("Registered a result consumer")
	return msgs
}
