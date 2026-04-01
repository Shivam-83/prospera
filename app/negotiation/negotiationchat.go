package negotiation

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/doniacld/prospera/app/gemini"
	"github.com/doniacld/prospera/app/user"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// incomingMessage is the shape the frontend sends over the WebSocket.
type incomingMessage struct {
	Message string `json:"message"`
}

// NegotiationChatWebsocketHandler is the websocket endpoint handler for negotiation chat.
func NegotiationChatWebsocketHandler(c *gin.Context) {
	userID := c.Query("userID")

	// Upgrade FIRST — doing it before the user lookup ensures the browser
	// receives a proper 101 and we can send a clean WS close frame on error,
	// rather than an HTTP 400 that kills the handshake entirely.
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Could not upgrade to WebSocket:", err)
		return
	}
	defer ws.Close()

	log.Println("Negotiation WebSocket connected, userID=", userID)

	// Validate user session.
	userDetails, ok := user.SalaryInfoPerUser[userID]
	if !ok {
		log.Println("Negotiation WS: user not found in memory, userID=", userID)
		_ = ws.WriteMessage(websocket.TextMessage, []byte(
			"⚠️ Your session has expired (the server may have restarted). "+
				"Please go back and fill in the form again to start a new session.",
		))
		ws.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "user not found"))
		return
	}

	log.Println("Negotiation WebSocket ready for user:", userID)

	// Generate AI opening message.
	chatInfo := gemini.NewChatInfo(userID)
	aiResponse, err := gemini.InitiateChat(chatInfo, buildNegotiationCoachPrompt(userDetails))
	if err != nil {
		log.Println("Gemini InitiateChat error:", err)
		_ = ws.WriteMessage(websocket.TextMessage, []byte(
			"⚠️ Sorry, I could not reach the AI service. Please try again in a moment.",
		))
		return
	}

	msg := coachPromptIntro + "\n\n" + aiResponse
	if err := ws.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
		log.Println("WS write error:", err)
		return
	}

	// Message loop.
	for {
		_, rawMsg, err := ws.ReadMessage()
		if err != nil {
			log.Println("Negotiation WS read closed:", err)
			return
		}

		// Decode incoming JSON: { "message": "..." } — fall back to raw text.
		var payload incomingMessage
		if jsonErr := json.Unmarshal(rawMsg, &payload); jsonErr != nil || payload.Message == "" {
			payload.Message = string(rawMsg)
		}

		log.Println("Negotiation WS received message from user:", userID)

		aiResponse, err := gemini.SendMessage(context.Background(), chatInfo, payload.Message)
		if err != nil {
			log.Println("Gemini SendMessage error:", err)
			_ = ws.WriteMessage(websocket.TextMessage, []byte(
				"⚠️ Sorry, I encountered an error. Please try again.",
			))
			continue
		}

		if err := ws.WriteMessage(websocket.TextMessage, []byte(aiResponse)); err != nil {
			log.Println("Negotiation WS write error:", err)
			return
		}
	}
}
