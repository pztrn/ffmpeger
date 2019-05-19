package nats

import (
	// stdlib
	"flag"
	"testing"

	// local
	"github.com/pztrn/ffmpeger/config"

	// other
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
)

func TestNATSInitialization(t *testing.T) {
	Initialize()
	require.NotNil(t, handlers)
	require.Empty(t, handlers)
}

func TestNATSStartListeningAndShutdown(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet("ffmpeger-test-nats", flag.ExitOnError)

	Initialize()
	require.NotNil(t, handlers)
	require.Empty(t, handlers)

	config.Initialize()
	config.Cfg.NATS.ConnectionString = "nats://127.0.0.1:14222"

	err := StartListening()
	require.Nil(t, err)

	err1 := Shutdown()
	require.Nil(t, err1)
}

func TestNATSShutdownWithoutConnection(t *testing.T) {
	Initialize()
	require.NotNil(t, handlers)
	require.Empty(t, handlers)

	err := Shutdown()
	require.NotNil(t, err)
}

func TestNATSConnectToWrongAddress(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet("ffmpeger-test-nats", flag.ExitOnError)
	Initialize()
	require.NotNil(t, handlers)
	require.Empty(t, handlers)

	config.Initialize()
	config.Cfg.NATS.ConnectionString = "nats://127.0.0.1:14223"

	err := StartListening()
	require.NotNil(t, err)
}

func TestNATSAddHandler(t *testing.T) {
	d := func(data []byte) {}

	Initialize()
	require.NotNil(t, handlers)
	require.Empty(t, handlers)

	hndl := &Handler{
		Name: "testhandler",
		Func: d,
	}
	AddHandler(hndl)
}

func TestNATSReceiveMessage(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet("ffmpeger-test-nats", flag.ExitOnError)

	Initialize()
	require.NotNil(t, handlers)
	require.Empty(t, handlers)

	config.Initialize()
	config.Cfg.NATS.ConnectionString = "nats://127.0.0.1:14222"

	err := StartListening()
	require.Nil(t, err)

	received := make(chan bool, 1)
	d := func(data []byte) {
		t.Log("Received data:", data)
		received <- true
	}

	hndl := &Handler{
		Name: "testhandler",
		Func: d,
	}
	AddHandler(hndl)

	// Send message.
	nc, err1 := nats.Connect(config.Cfg.NATS.ConnectionString)
	require.Nil(t, err1)

	err2 := nc.Publish(Topic, []byte("Hello, world!"))
	require.Nil(t, err2)

	<-received
	nc.Close()

	err3 := Shutdown()
	require.Nil(t, err3)
}
