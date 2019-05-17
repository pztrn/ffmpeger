package main

import (
	// stdlib
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	// local
	"github.com/pztrn/ffmpeger/config"
	"github.com/pztrn/ffmpeger/converter"
	"github.com/pztrn/ffmpeger/nats"
)

func main() {
	log.Println("Starting video conversion service")

	config.Initialize()
	nats.Initialize()
	converter.Initialize()

	flag.Parse()

	config.Load()
	nats.StartListening()
	converter.Start()

	// CTRL+C handler.
	signalHandler := make(chan os.Signal, 1)
	shutdownDone := make(chan bool, 1)
	signal.Notify(signalHandler, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalHandler
		nats.Shutdown()
		converter.Shutdown()
		shutdownDone <- true
	}()

	<-shutdownDone
	os.Exit(0)
}
