package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var connections = make(map[uuid.UUID]*websocket.Conn)
var lock = sync.Mutex{}

func broadcastTotalConnections() {
	message := []byte(fmt.Sprintf("UserCount:%d", len(connections)))
	for _, conn := range connections {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			fmt.Println("Error broadcasting total connections:", err)
		}
	}
}

func main() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer conn.Close()

		id := uuid.New()

		lock.Lock()
		connections[id] = conn
		broadcastTotalConnections()
		lock.Unlock()

		for {
			// Read message from browser
			_, msg, err := conn.ReadMessage()
			if err != nil {
				lock.Lock()
				delete(connections, id) // Remove connection 
				broadcastTotalConnections()
				lock.Unlock()

				fmt.Println("Error reading message:", err)
				return
			}
			if string(msg) == "Websocket closed" {
				// If disconenct message is sent without error (Will this actually happen?)
				lock.Lock()
				connections[id] = conn
				broadcastTotalConnections()
				lock.Unlock()
			} else {
				fmt.Printf("received from %s: %s\n", conn.RemoteAddr(), string(msg))
			}
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/index.html")
	})

	// Serve static files from the web directory
	fs := http.FileServer(http.Dir("web"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))


	log.Fatal(http.ListenAndServe(":8080", nil))
}
