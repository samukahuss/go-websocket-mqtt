package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	pb "github.com/samukahuss/go-websocket-mqtt/client/internal/proto"
)

func main() {
	// configura um canal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	url := "ws://localhost:8080/ws"
	log.Printf("Conneting to %s", url)

	// websocket dialer
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("Dial failed: %v", err)
	}

	defer c.Close()

	// criando um canal done pra sinalizar quando o loop finalizar
	done := make(chan struct{})

	go func() {
		defer close(done)

		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Printf("Read error: %v", err)
				return
			}

			var receivedMsg pb.WebSocketMessage
			if err := proto.Unmarshal(message, &receivedMsg); err != nil {
				log.Printf("Failed to unmarshal echoed message: %v", err)
				continue
			}

			log.Printf("<- Received echo: Type='%s', Payload='%s'", receivedMsg.Type, string(receivedMsg.Payload))
		}
	}()

	// Logica central do protobuf
	msg := &pb.WebSocketMessage{
		Type:    "greeting",
		Payload: []byte("Hello from a proper Go client!"),
	}

	// Fazendo o Marshall da struct em binario
	marshaledMsg, err := proto.Marshal(msg)
	if err != nil {
		log.Fatal("Marshaling failed: %v", err)
	}

	log.Printf("-> Sending message: Type='%s', Payload='%s'", msg.Type, string(msg.Payload))

	// enviando a msg pelo websocket
	err = c.WriteMessage(websocket.BinaryMessage, marshaledMsg)
	if err != nil {
		log.Println("Write error:", err)
		return
	}

	// fazendo o client esperar ate a interrupcao
	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("Interrupt received, closing connection")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Printf("Write close error: %v", err)
				return
			}

			select {
			case <-done:
			case <-time.After(time.Second):
			}

			return
		}
	}
}
