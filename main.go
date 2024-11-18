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
			fmt.Println("Upgrade error:", err)
			return
		}
		defer conn.Close()

		id := uuid.New()

		lock.Lock()
		connections[id] = conn
		broadcastTotalConnections()
		lock.Unlock()

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				lock.Lock()
				delete(connections, id)
				broadcastTotalConnections()
				lock.Unlock()

				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					fmt.Println("Normal closure:", err)
				} else if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					fmt.Printf("Unexpected close error: %v\n", err)
				} else {
					fmt.Println("Other error:", err)
				}
				fmt.Printf("total connections: %d\n", len(connections))
				return
			} else {
				fmt.Printf("received from %s: %s\n", conn.RemoteAddr(), string(msg))
				fmt.Printf("total connections: %d\n", len(connections))
			}
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/index.html")
	})

	// Serve static files from the web directory
	fs := http.FileServer(http.Dir("web"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Println("Started server on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
