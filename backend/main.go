package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	mqttclient "github.com/samukahuss/go-websocket-mqtt/backend/mqtt"
	ws "github.com/samukahuss/go-websocket-mqtt/backend/websocket"
)

const (
	mqttBroker = "localhost:1883"
	mqttTopic  = "go-websocket-mqtt"
	serverAddr = "localhost:8080"
)

// Um upgrader pra todas as connections
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Pra teste toda origem eh liberada
	CheckOrigin: func(r *http.Request) bool { return true },
}

func serveWs(hub *ws.Hub, w http.ResponseWriter, r *http.Request) {
	log.Println("New Websocket connection request received")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection to websocket: %v", err)
	}

	// Cada client precisa de um ID unico
	clientID := uuid.New().String()

	// registra o client no hub
	client := ws.NewClient(hub, conn, clientID)
	hub.Register <- client

	go client.WritePump()
	go client.ReadPump()

	log.Printf("Websocket client %s connected and pumps are started", client.ID)
}

func main() {
	log.Println("--- Go Websocket to MQTT Bridge ---")

	// criando o mqtt client, passando o client id
	mqttClient := mqttclient.NewClient(mqttBroker, "go-websocket-mqtt")
	log.Println("MQTT client created and connecting...")

	// criadno o websocket hub
	hub := ws.NewHub(mqttClient, mqttTopic)

	// rodando o hub em uma go routine separada pra nao bloquear a main trhead
	go hub.Run()
	log.Println("Websocket is running in background")

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) { serveWs(hub, w, r) })

	// iniciando o httpserver em uma go routine
	go func() {
		log.Printf("HTTP server starting, listening on %s", serverAddr)
		if err := http.ListenAndServe(serverAddr, nil); err != nil {
			log.Fatal("ListenAndServe failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Bloqueia ate receber o sinal de shutdown
	<-quit
	log.Println("--- Server is shutting down ---")
}
