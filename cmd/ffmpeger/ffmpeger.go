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

	err := config.Load()
	if err != nil {
		log.Fatalln("Failed to load configuration file:", err.Error())
	}
	err1 := nats.StartListening()
	if err1 != nil {
		log.Fatalln("Failed to establish connection to NATS:", err1.Error())
	}
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
