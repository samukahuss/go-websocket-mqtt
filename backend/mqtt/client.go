package mqtt

import (
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("Connected to MQTT broker")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("Connection to MQTT broker lost: %v", err)
}

func NewClient(broker, clientID string) mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", broker))
	opts.SetClientID(clientID)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	token := client.Connect()

	if token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	return client
}

// Packages: The package mqtt declaration at the top must match the name of the directory the file is in. This is how Go organizes code.
// Callbacks: The OnConnect and OnConnectionLost handlers are a great example of a common programming pattern. You give another piece of code (the MQTT library) a function to execute when a specific event occurs.
// Asynchronous Operations: client.Connect() doesn't wait. It starts the work in the background. The token it returns is an object that lets you check the status of that work later, for example by calling token.Wait().
// Error Handling: The if token.Wait() && token.Error() != nil block is a standard way to handle errors from asynchronous operations in this library.
