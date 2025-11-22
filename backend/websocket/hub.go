package websocket

import (
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	pb "github.com/samukahuss/go-websocket-mqtt/proto"
	proto "google.golang.org/protobuf/proto"
)

type TargeredMessage struct {
	TargetID string
	Payload  []byte
}

type Hub struct {
	// map com k=clientId v=clientObj
	clients map[string]*Client
	// pra receber menssagens de um unico client
	Unicast chan *TargeredMessage
	// um channel pra novos clients
	register chan *Client
	// channel pra remover clients desconectados
	unregister chan *Client
	// guardando uma referencia do MQTT client pra publicar menssagens
	mqttClient mqtt.Client
	mqttTopic  string
}

// NewHub eh o consrutor de Hub
func NewHub(mqttClient mqtt.Client, topic string) *Hub {
	return &Hub{
		Unicast:    make(chan *TargeredMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
		mqttClient: mqttClient,
		mqttTopic:  topic,
	}
}

// loop principal
func (h *Hub) Run() {
	for {
		select {
		// Case 1: client connecta
		case client := <-h.register:
			h.clients[client.ID] = client
			log.Printf("Client %s registered", client.ID)

		case client := <-h.unregister:
			// verificando se o client existe
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.send)
				log.Printf("Client %s unregistered")
			}
		case message := <-h.Unicast:

			client, ok := h.clients[message.TargetID]

			if ok {
				// Se existe um client, envia mensagem pro seu canal
				// Estou usando um select exclusivo pra casos onde o canal possa estar cheio
				select {
				case client.send <- message.Payload:
				// se nao consegui enviar estou assumindo que o canal morreu
				default:
					close(client.send)
					delete(h.clients, client.ID)
				}
			} else {
				log.Printf("Client %s not found for unicast message", message.TargetID)
			}

			var msg pb.WebSocketMessage

			if err := proto.Unmarshal(message.Payload, &msg); err == nil {
				log.Printf("Publishing message type '%s' to MQTT topic '%s'", msg.Type, h.mqttTopic)
				h.mqttClient.Publish(h.mqttTopic, 0, false, message.Payload)
			}

		}
	}
}
