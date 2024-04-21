package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

const chatMessageKey = "chat_message"

type ChatMessage struct {
	Username string `json:"username"`
	Text     string `json:"text"`
}

var (
	rdb         *redis.Client
	clients     = make(map[*websocket.Conn]bool)
	broadcaster = make(chan ChatMessage)
	upgrader    = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	defer ws.Close()
	clients[ws] = true

	if rdb.Exists(ctx, chatMessageKey).Val() != 0 {
		sendPreviousMessage(ws)
	}

	for {
		var msg ChatMessage
		err := ws.ReadJSON(&msg)
		if err != nil {
			delete(clients, ws)
			break
		}

		broadcaster <- msg
	}
}

func handleMessage() {
	for {
		msg := <-broadcaster
		saveInRedis(msg)
		messageClients(msg)
	}
}

func saveInRedis(msg ChatMessage) {
	json, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	log.Print("Saved message json", json)
	if err := rdb.RPush(ctx, chatMessageKey, json).Err(); err != nil {
		panic(err)
	}
}

func messageClients(msg ChatMessage) {
	for client := range clients {
		sendMessageForClient(client, msg)
	}
}

func sendMessageForClient(client *websocket.Conn, msg ChatMessage) {
	err := client.WriteJSON(msg)
	if err != nil && unsafeError(err) {
		log.Print("error sending message", err)
		client.Close()
		delete(clients, client)
	}
}

func unsafeError(err error) bool {
	return !websocket.IsCloseError(err, websocket.CloseGoingAway) && err != io.EOF
}

func sendPreviousMessage(ws *websocket.Conn) {
	chatMessages, err := rdb.LRange(ctx, chatMessageKey, 0, -1).Result()
	if err != nil {
		panic(err)
	}

	for _, chchatMessage := range chatMessages {
		var msg ChatMessage
		json.Unmarshal([]byte(chchatMessage), &msg)
		sendMessageForClient(ws, msg)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error .env load")
	}

	// Redis
	redisUrl := os.Getenv("REDIS_URL")
	opt, err := redis.ParseURL(redisUrl)
	if err != nil {
		panic(err)
	}
	rdb = redis.NewClient(opt)
	log.Print("Redis connected")

	// Http server
	port := os.Getenv("PORT")
	http.Handle("/", http.FileServer(http.Dir("./static")))

	// Websocket handler
	http.HandleFunc("/ws", handleConnections)
	go handleMessage()

	log.Print("Server starting at localhost:" + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
