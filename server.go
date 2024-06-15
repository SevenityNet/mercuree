package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

type mercureeServer struct {
	upgrader *websocket.Upgrader
	clients  *clientList
}

func newMercureeServer() *mercureeServer {
	upgrader := websocket.NewUpgrader()
	upgrader.KeepaliveTime = getKeepaliveTimeout()
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	server := &mercureeServer{
		upgrader: upgrader,
		clients:  newClientList(),
	}

	upgrader.OnMessage(func(c *websocket.Conn, messageType websocket.MessageType, b []byte) {
		data := string(b)
		if data == "PING" {
			c.WriteMessage(websocket.TextMessage, []byte("PONG"))
		} else {
			log.Println("WARNING: received an unsupported message from " + c.RemoteAddr().String() + " -> " + data)
		}
	})

	return server
}

func (s *mercureeServer) subscriber() gin.HandlerFunc {
	return func(c *gin.Context) {
		topics := c.GetStringSlice("topics")

		conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			panic(err)
		}

		client := &mercureeClient{
			conn:   conn,
			topics: make(map[string]bool),
		}

		for _, topic := range topics {
			client.topics[topic] = true
		}

		client.conn.OnClose(func(h *websocket.Conn, err error) {
			if client.conn != h {
				return
			}

			log.Println("INFO: connection closed from " + h.RemoteAddr().String())

			if err != nil && err.Error() != "EOF" {
				log.Println("ERROR: connection closed with error: " + err.Error())
			}

			s.clients.removeClient(client)
		})

		s.clients.addClient(client)

		log.Println("INFO: new connection from " + conn.RemoteAddr().String() + " subscribed to " + strings.Join(topics, ", ") + ".")
	}
}

func (s *mercureeServer) publish(topic, data string) {
	s.clients.broadcast(topic, data)
}

func getKeepaliveTimeout() time.Duration {
	e := os.Getenv("KEEPALIVE_TIMEOUT")
	if e == "" {
		return time.Minute * 10
	}

	d, err := strconv.Atoi(e)
	if err != nil {
		return time.Minute * 10
	}

	return time.Duration(d) * time.Minute
}
