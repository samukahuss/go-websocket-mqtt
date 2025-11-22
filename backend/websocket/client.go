package websocket

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
	pb "github.com/samukahuss/go-websocket-mqtt/proto"
	proto "google.golang.org/protobuf/proto"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

type Client struct {
	ID   string
	Hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

// Client constructor
func NewClient(hub *Hub, conn *websocket.Conn, id string) *Client {
	return &Client{
		ID:   id,
		Hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
	}
}

// da um pump na mensagem do websocket pro hub
func (c *Client) ReadPump() {
	// defer roda antes da func ReadPump() e faz uma limpa
	defer func() {
		c.Hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	// pytonico ou nao?
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	// loop infinito
	for {
		// pra cada mensagem no websocket
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("Read error for client %s: %v", c.ID, err)
			break
		}

		// Como exemplo vou apenas mandar um echo devolta pra quem mandou
		// e publicar a mensagem no MQTT topic

		var msg pb.WebSocketMessage
		if err := proto.Unmarshal(messageBytes, &msg); err != nil {
			log.Printf("protobuf unmarshal error: %v", err)
			continue
		}

		log.Printf("Received message type %s from client %s", msg.Type, c.ID)

		// criando uma targeredMessage pra mandar devolta
		targeredMessage := &TargeredMessage{
			TargetID: c.ID,
			Payload:  messageBytes,
		}

		// manda a msg pro unicast
		c.Hub.Unicast <- targeredMessage
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	// garantindo que a conexao esteja fechada e o ticker parado qndo sair da funcao
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		// tem uma msg pra ser enviada no send channel
		case message, ok := <-c.send:

			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				// O hub fechou o canal
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// escreve a mensagem
			c.conn.WriteMessage(websocket.BinaryMessage, message)

		// ticker foi disparado. Enviando um ping
		case <-ticker.C:

			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
