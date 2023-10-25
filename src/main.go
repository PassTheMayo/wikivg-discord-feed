package main

import (
	"log"
	"os"
	"os/signal"
	"time"
)

var (
	lastCheck time.Time = ReadLastFetchTimestamp()
	config    *Config   = &Config{}
)

func init() {
	if err := config.ReadFile("config.yml"); err != nil {
		log.Fatal(err)
	}
}

func main() {
	go ProcessAllFeedGoroutine()

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)
	<-s
}
