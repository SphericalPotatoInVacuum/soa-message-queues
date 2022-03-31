package main

import (
	"fmt"

	"github.com/SphericalPotatoInVacuum/soa-message-queues/internal/grabber"
	"github.com/SphericalPotatoInVacuum/soa-message-queues/internal/utils"
	log "github.com/sirupsen/logrus"
)

var getEnv = utils.GetEnv
var failOnError = utils.FailOnError

func getRabbitAddr() string {
	user := getEnv("RABBITMQ_USER", "guest")
	pass := getEnv("RABBITMQ_PASS", "guest")
	host := getEnv("RABBITMQ_HOST", "rabbitmq")
	port := getEnv("RABBITMQ_PORT", "5672")
	return fmt.Sprintf("amqp://%s:%s@%s:%s/", user, pass, host, port)
}

func main() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors: true,
	})

	g := grabber.NewGrabber(getRabbitAddr())
	g.Run()
}
