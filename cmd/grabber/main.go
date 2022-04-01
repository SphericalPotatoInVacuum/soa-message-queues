package main

import (
	"fmt"
	"os"

	"github.com/SphericalPotatoInVacuum/soa-message-queues/internal/grabber"
	"github.com/SphericalPotatoInVacuum/soa-message-queues/internal/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	g := grabber.NewGrabber(getRabbitAddr())
	g.Run()
}
