package utils

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/ttacon/chalk"
)

func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", chalk.Red.Color(msg), err)
	}
}

func GetEnv(key string, defaultVal string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return defaultVal
}
