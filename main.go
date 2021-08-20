package main

import (
	"flag"
	"fmt"
	"github.com/SINTEF-Infosec/demokit/core"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func usage() {
	fmt.Println("./demokit-utils [COMMAND] ARGUMENTS")
	fmt.Println("")
	fmt.Println("Available commands:")
	fmt.Println("\t- monitor: shows events in the network")
	fmt.Println("\t- send: send an event in the network")
	fmt.Println("")
	fmt.Println("Example: ./demokit-utils monitor")
}

func main() {
	_ = flag.NewFlagSet("monitor", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	eventName := sendCmd.String("event", "", "Event name")
	payload := sendCmd.String("payload", "", "Event payload")
	emitter := sendCmd.String("emitter", "cli", "Emitter of the event, default \"cli\"")
	receiver := sendCmd.String("receiver", "*", "Receiver of the event, default \"*\"")

	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	// Connecting to the network
	network := core.NewRabbitMQEventNetwork(core.ConnexionDetails{
		Username: "guest",
		Password: "guest",
		Host:     "localhost",
		Port:     "5672",
	}, logrus.WithField("name", "cli"))

	switch os.Args[1] {
	case "monitor":
		network.SetReceivedEventCallback(handleEvent)
		network.StartListeningForEvents()

		sigs := make(chan os.Signal, 1)
		done := make(chan bool, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigs
			logrus.WithField("name", "cli").Info("Stopping monitor...")
			done <- true
		}()
		<-done
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			usage()
			os.Exit(1)
		}
		event := &core.Event{
			Name:     *eventName,
			Emitter:  *emitter,
			Payload:  *payload,
			Receiver: *receiver,
		}
		logrus.Debugf("%v", event)
		network.BroadcastEvent(event)
	default:
		usage()
		os.Exit(1)
	}
}

func handleEvent(event *core.Event) {
	currentTime := time.Now()
	fmt.Printf("%s - EVENT NAME: %-15s\tEmitter: %-20s\tReceiver: %-20s\tPayload: %-40s\n", currentTime.Format("2006.01.02 15:04:05"), event.Name, event.Emitter, event.Receiver, event.Payload)
}
