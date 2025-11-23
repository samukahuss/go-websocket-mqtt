package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"google.golang.org/protobuf/proto"

	// Import our shared protobuf package
	pb "github.com/samukahuss/go-websocket-mqtt/proto"
)

// These must match the settings in your backend server
const (
	mqttBroker = "localhost:1883"
	mqttTopic  = "go-websocket-messages"
)

// messagePubHandler is the callback that gets executed every time a message
// is received on a topic that we are subscribed to.
var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received message on topic: %s", msg.Topic())

	// We receive the raw binary data, so we need to unmarshal it
	// back into our WebSocketMessage struct.
	var receivedMsg pb.WebSocketMessage
	if err := proto.Unmarshal(msg.Payload(), &receivedMsg); err != nil {
		log.Printf("Failed to unmarshal protobuf message: %v", err)
		return
	}

	log.Printf("--- MQTT Message Decoded ---")
	log.Printf("Type:    %s", receivedMsg.Type)
	log.Printf("Payload: %s", string(receivedMsg.Payload))
	log.Printf("--------------------------")
}

// connectHandler is executed upon a successful connection.
// We use it to subscribe to our topic.
var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("Connected to MQTT broker")
	// Subscribe to the topic
	token := client.Subscribe(mqttTopic, 1, nil) // QoS 1
	token.Wait()
	log.Printf("Subscribed to topic: %s", mqttTopic)
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("Connection to MQTT broker lost: %v", err)
}

func main() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", mqttBroker))
	opts.SetClientID("go_mqtt_subscriber")

	// Set the default message handler to our function
	opts.SetDefaultPublishHandler(messagePubHandler)
	// Set the OnConnect handler to subscribe upon connection
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	log.Println("MQTT Subscriber started. Waiting for messages. Press Ctrl+C to exit.")

	// Wait for an interrupt signal to gracefully shut down.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down subscriber.")
	client.Disconnect(250)
}
