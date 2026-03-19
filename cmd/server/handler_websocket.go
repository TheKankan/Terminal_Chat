package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/TheKankan/TerminalSecuredChat/internal/auth"
	"github.com/TheKankan/TerminalSecuredChat/internal/database"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	clients   = make(map[*websocket.Conn]string)
	clientsMu sync.Mutex
)

func (cfg *apiConfig) handlerWebSocket(w http.ResponseWriter, r *http.Request) {
	// Extract and validate JWT token
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, "missing auth token", http.StatusUnauthorized)
		return
	}

	userID, err := auth.ValidateJWT(tokenString, cfg.jwtSecret)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	username, err := cfg.db.GetUsernameFromID(r.Context(), userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}

	// Register client
	clientsMu.Lock()
	clients[conn] = username
	clientsMu.Unlock()

	log.Printf("Client %s connected\n", username)

	// Send chat history to the newly connected client
	cfg.sendHistory(conn)

	broadcast(websocket.TextMessage, []byte("User "+username+" connected"))

	defer func() {
		clientsMu.Lock()
		delete(clients, conn)
		clientsMu.Unlock()
		conn.Close()
		log.Printf("Client %s disconnected\n", username)
		broadcast(websocket.TextMessage, []byte("User "+username+" disconnected"))
	}()

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}

		// Save message to DB
		_, err = cfg.db.CreateMessage(context.Background(), database.CreateMessageParams{
			UserID:  userID,
			Content: string(msg),
		})
		if err != nil {
			log.Println("error saving message:", err)
		}

		// Format and broadcast
		now := time.Now().Format("15h04")
		formattedMsg := fmt.Sprintf("[%s] %s: %s", now, username, string(msg))
		log.Printf("Received: %s\n", formattedMsg)
		broadcast(msgType, []byte(formattedMsg))
	}
}

func (cfg *apiConfig) sendHistory(conn *websocket.Conn) {
	messages, err := cfg.db.GetRecentMessages(context.Background(), 20)
	if err != nil {
		log.Println("error fetching history:", err)
		return
	}

	// Messages are ordered DESC, reverse them to show oldest first
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	for _, msg := range messages {
		formatted := fmt.Sprintf("[%s] %s: %s",
			msg.CreatedAt.Format("15h04"),
			msg.Username,
			msg.Content,
		)
		conn.WriteMessage(websocket.TextMessage, []byte(formatted))
	}
}

func broadcast(msgType int, msg []byte) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for client := range clients {
		err := client.WriteMessage(msgType, msg)
		if err != nil {
			log.Println("write error:", err)
			client.Close()
			delete(clients, client)
		}
	}
}
