package main

import (
	// stdlib
	"encoding/json"
	"flag"
	"log"
	"path/filepath"

	// local
	"github.com/pztrn/ffmpeger/config"
	"github.com/pztrn/ffmpeger/converter"
	mynats "github.com/pztrn/ffmpeger/nats"

	// other
	"github.com/nats-io/nats.go"
)

var (
	inputFilename  string
	outputFilename string
)

func main() {
	log.Println("Starting example message sender...")

	flag.StringVar(&inputFilename, "input", "", "Input file name")
	flag.StringVar(&outputFilename, "output", "", "Output file name")

	config.Initialize()

	flag.Parse()

	if inputFilename == "" || outputFilename == "" {
		log.Fatalln("Please specify both input and output file name!")
	}

	var err error
	inputFilename, err = filepath.Abs(inputFilename)
	if err != nil {
		log.Fatalln("Failed to get absolute path for input filename:", err.Error())
	}
	outputFilename, err = filepath.Abs(outputFilename)
	if err != nil {
		log.Fatalln("Failed to get absolute path for output filename:", err.Error())
	}

	err = config.Load()
	if err != nil {
		log.Fatalln("Failed to load configuration file:", err.Error())
	}

	nc, err := nats.Connect(config.Cfg.NATS.ConnectionString)
	if err != nil {
		log.Fatalln("Failed to connect to NATS server:", err.Error())
	}

	t := &converter.Task{
		InputFile:  inputFilename,
		OutputFile: outputFilename,
	}

	data, err1 := json.Marshal(t)
	if err1 != nil {
		log.Fatalln("Failed to encode message:", err1.Error())
	}

	err2 := nc.Publish(mynats.Topic, data)
	if err2 != nil {
		log.Fatalln("Failed to publish message:", err2.Error())
	}

	log.Println("Message published")
	nc.Close()
}
