package nats

import (
	// stdlib
	"errors"
	"log"
	"sync"

	// local
	"github.com/pztrn/ffmpeger/config"

	// other
	"github.com/nats-io/nats.go"
)

const (
	Topic = "ffmpeger.v1"
)

var (
	natsConn         *nats.Conn
	natsSubscription *nats.Subscription

	// Handlers.
	handlers      []*Handler
	handlersMutex sync.Mutex
)

// AddHandler adds handler for received NATS messages.
func AddHandler(hndl *Handler) {
	handlersMutex.Lock()
	handlers = append(handlers, hndl)
	handlersMutex.Unlock()
}

// Initialize initializes package.
func Initialize() {
	log.Println("Initializing NATS handler...")

	handlers = make([]*Handler, 0, 8)
}

// Handler for NATS messages.
func messageHandler(msg *nats.Msg) {
	log.Println("Received message:", string(msg.Data))

	handlersMutex.Lock()
	for _, hndl := range handlers {
		hndl.Func(msg.Data)
	}
	handlersMutex.Unlock()
}

// Shutdown unsubscribes from topic and disconnects from NATS.
func Shutdown() error {
	log.Println("Unsuscribing from NATS topic...")
	err := natsSubscription.Unsubscribe()
	if err != nil {
		return errors.New("ERROR unsubscribing " + Topic + " topic: " + err.Error())
	}

	if natsConn != nil {
		log.Println("Closing connection to NATS...")
		natsConn.Close()
	}

	return nil
}

// StartListening connects to NATS and starts to listen for messages.
func StartListening() error {
	nc, err := nats.Connect(config.Cfg.NATS.ConnectionString)
	if err != nil {
		return errors.New("Failed to connect to NATS:" + err.Error())
	}

	natsConn = nc
	log.Println("NATS connection established")

	// Beware - if ffmpeger will be launched more than once and subscribed
	// to same topic (which is hardcoded here) then ALL instances of
	// ffmpeger will receive this message!
	sub, err1 := nc.Subscribe(Topic, messageHandler)
	if err1 != nil {
		return errors.New("Failed to subscribe to " + Topic + " topic: " + err1.Error())
	}
	natsSubscription = sub
	log.Println("Subscribed to topic", Topic)

	return nil
}
