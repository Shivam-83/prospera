package tips

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

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

// TipsChatWebsocketHandler is the websocket endpoint handler for confidence tips chat.
func TipsChatWebsocketHandler(c *gin.Context) {
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

	log.Println("Tips WebSocket connected, userID=", userID)

	// Validate user session.
	userDetails, ok := user.SalaryInfoPerUser[userID]
	if !ok {
		log.Println("Tips WS: user not found in memory, userID=", userID)
		_ = ws.WriteMessage(websocket.TextMessage, []byte(
			"⚠️ Your session has expired (the server may have restarted). "+
				"Please go back and fill in the form again to start a new session.",
		))
		ws.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "user not found"))
		return
	}

	log.Println("Tips WebSocket ready for user:", userID)

	intro := `Nice move! Here are some personalized negotiation tips to help you ` +
		`make a confident impression during your salary discussion. ` +
		`Remember, good preparation and the right strategies can make all the difference. ` +
		`Let's get you equipped for success!` + "\n\n"

	// Generate AI opening message.
	chatInfo := gemini.NewChatInfo(userID)
	aiResponse, err := gemini.InitiateChat(chatInfo, buildTipsPrompt(userDetails))
	if err != nil {
		log.Println("Gemini InitiateChat error:", err)
		_ = ws.WriteMessage(websocket.TextMessage, []byte(
			"⚠️ Sorry, I could not reach the AI service. Please try again in a moment.",
		))
		return
	}

	intro += aiResponse

	if err := ws.WriteMessage(websocket.TextMessage, []byte(intro)); err != nil {
		log.Println("WS write error:", err)
		return
	}

	// Message loop.
	for {
		_, rawMsg, err := ws.ReadMessage()
		if err != nil {
			log.Println("Tips WS read closed:", err)
			return
		}

		// Decode incoming JSON: { "message": "..." } — fall back to raw text.
		var payload incomingMessage
		if jsonErr := json.Unmarshal(rawMsg, &payload); jsonErr != nil || payload.Message == "" {
			payload.Message = string(rawMsg)
		}

		log.Println("Tips WS received message from user:", userID)

		aiResponse, err := gemini.SendMessage(context.Background(), chatInfo, payload.Message)
		if err != nil {
			log.Println("Gemini SendMessage error:", err)
			_ = ws.WriteMessage(websocket.TextMessage, []byte(
				"⚠️ Sorry, I encountered an error. Please try again.",
			))
			continue
		}

		if err := ws.WriteMessage(websocket.TextMessage, []byte(aiResponse)); err != nil {
			log.Println("Tips WS write error:", err)
			return
		}
	}
}

func buildTipsPrompt(u user.SalaryInfo) string {
	tips := ""
	tips += fmt.Sprintf("I am a %s with %d years of experience in %s.\n", u.JobTitle, u.YearsExperience, u.Industry)
	tips += fmt.Sprintf("I am currently exploring new opportunities in %s.\n", u.Location)
	tips += fmt.Sprintf("Currently, I earn %d, and my target salary is %d.\n", u.CurrentSalary, u.DesiredSalary)
	tips += fmt.Sprintf("Skills: %s.\n", strings.Join(u.Skills, ", "))
	tips += fmt.Sprintf("Education includes a major in %s, graduated with a %s.\n", u.Major, u.Diploma)

	endingNote := `
Negotiation Tips:
1. **Know Your Worth**: Do thorough research on market salary ranges for roles like %s, especially in the %s industry, and don't hesitate to back up your ask with examples of your achievements and skills.
2. **Negotiate Beyond Salary**: Besides salary, consider asking for benefits like flexible working hours, professional development budgets, or equity shares.
3. **Express Your Value Confidently**: When negotiating, focus on the value you bring to the team and company, showcasing how your skills and experience directly contribute to company goals.

Confidence-Boosting Tips for Women in Tech:
1. **Advocate for Yourself**: Don't wait for recognition—speak up about your achievements and ask for opportunities that align with your career growth.
2. **Leverage a Mentor Network**: Connect with other women in tech for guidance, mentorship, and support, which can help boost both confidence and career development.
3. **Stay Curious and Keep Learning**: Continuously upskill and stay updated with industry trends, reinforcing your confidence in your expertise and readiness for new challenges.

Good luck with your negotiation journey! Believe in your value, and remember that confidence grows with every step you take in advocating for yourself. You've got this!
`

	return tips + fmt.Sprintf(endingNote, u.JobTitle, u.Industry)
}
