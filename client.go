package main

import (
	"errors"
	"log"
	"sync"

	"github.com/lesismal/nbio/nbhttp/websocket"
)

type mercureeClient struct {
	topics map[string]bool
	conn   *websocket.Conn
}

func (c *mercureeClient) publish(topic, data string) (bool, error) {
	if c.conn == nil {
		return false, errors.New("connection is not established")
	}

	if _, ok := c.topics[topic]; !ok {
		return false, nil
	}

	if err := c.conn.WriteMessage(websocket.TextMessage, []byte(data)); err != nil {
		return false, err
	}

	return true, nil
}

type clientList struct {
	clients map[*mercureeClient]bool
	lock    *sync.RWMutex
}

func (cl *clientList) addClient(client *mercureeClient) {
	cl.lock.Lock()
	defer cl.lock.Unlock()

	cl.clients[client] = true
}

func (cl *clientList) removeClient(client *mercureeClient) {
	cl.lock.Lock()
	defer cl.lock.Unlock()

	delete(cl.clients, client)
}

func (cl *clientList) broadcast(topic, data string) {
	cl.lock.RLock()
	defer cl.lock.RUnlock()

	for client := range cl.clients {
		success, err := client.publish(topic, data)
		if err != nil {
			log.Println("ERROR: failed to publish message to " + client.conn.RemoteAddr().String() + " -> " + err.Error())
		} else {
			if success {
				log.Println("VERBOSE: message published to " + client.conn.RemoteAddr().String())
			} else {
				log.Println("VERBOSE: message not published to " + client.conn.RemoteAddr().String())
			}
		}
	}
}

func newClientList() *clientList {
	return &clientList{
		clients: make(map[*mercureeClient]bool),
		lock:    &sync.RWMutex{},
	}
}
