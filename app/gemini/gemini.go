package gemini

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid"
	"google.golang.org/api/option"
)

const (
	googleApiKeyEnv = "GOOGLE_API_KEY"
	generativeModel = "gemini-2.0-flash"

	roleModel = "model"
	roleUser  = "user"
)

type ChatInfo struct {
	userID    string
	sessionID string
}

// chatsMu protects ChatsInfoPerUser against concurrent WebSocket goroutines.
var (
	ChatsInfoPerUser = map[ChatInfo]*genai.ChatSession{}
	chatsMu          sync.RWMutex
)

func NewChatInfo(userID string) ChatInfo {
	chatInfo := ChatInfo{userID: userID, sessionID: uuid.NewString()}

	chatsMu.Lock()
	ChatsInfoPerUser[chatInfo] = &genai.ChatSession{}
	chatsMu.Unlock()

	return chatInfo
}

// newClient creates an authenticated Gemini API client.
func newClient(ctx context.Context) (*genai.Client, error) {
	apiKey := os.Getenv(googleApiKeyEnv)
	if apiKey == "" {
		return nil, errors.New("GOOGLE_API_KEY environment variable is not set")
	}
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	return client, nil
}

// sendWithRetry calls cs.SendMessage and retries up to 3 times on 429 rate-limit errors.
func sendWithRetry(ctx context.Context, cs *genai.ChatSession, msg string) (*genai.GenerateContentResponse, error) {
	var res *genai.GenerateContentResponse
	var err error

	for attempt := 0; attempt < 3; attempt++ {
		res, err = cs.SendMessage(ctx, genai.Text(msg))
		if err == nil {
			return res, nil
		}
		if !strings.Contains(err.Error(), "429") {
			// Not a rate-limit error — no point retrying.
			return nil, err
		}
		waitSec := time.Duration(3*(attempt+1)) * time.Second
		log.Printf("Gemini rate limit hit (429), retrying in %v (attempt %d/3)...", waitSec, attempt+1)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(waitSec):
		}
	}
	return nil, fmt.Errorf("gemini rate limit exceeded after retries: %w", err)
}

func InitiateChat(info ChatInfo, msg string) (string, error) {
	ctx := context.Background()

	client, err := newClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	model := client.GenerativeModel(generativeModel)
	cs := model.StartChat()
	// NOTE: Do NOT pre-append the user message to cs.History here.
	// cs.SendMessage() manages the conversation history automatically.
	// Pre-appending before calling SendMessage causes the message to appear
	// twice in the history, corrupting the conversation context.

	res, err := sendWithRetry(ctx, cs, msg)
	if err != nil {
		log.Println("Gemini SendMessage error:", err)
		return "", fmt.Errorf("gemini error: %w", err)
	}

	if len(res.Candidates) == 0 || len(res.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("empty response from Gemini")
	}

	// Save the chat session (history is now managed by the SDK internally).
	chatsMu.Lock()
	ChatsInfoPerUser[info] = cs
	chatsMu.Unlock()

	resp := string(res.Candidates[0].Content.Parts[0].(genai.Text))
	return resp, nil
}

func SendMessage(ctx context.Context, info ChatInfo, msg string) (string, error) {
	chatsMu.RLock()
	chatSession, ok := ChatsInfoPerUser[info]
	chatsMu.RUnlock()

	if !ok {
		return "", errors.New("no chat session found")
	}

	if chatSession == nil {
		log.Println("chatSession is nil for info:", info)
		return "", errors.New("chat session is nil")
	}

	log.Println("Sending follow-up message to Gemini, history length:", len(chatSession.History))

	client, err := newClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	model := client.GenerativeModel(generativeModel)
	// Restore the chat session by reusing the existing history from the SDK session.
	cs := model.StartChat()
	cs.History = chatSession.History

	res, err := sendWithRetry(ctx, cs, msg)
	if err != nil {
		return "", fmt.Errorf("gemini error: %w", err)
	}

	if len(res.Candidates) == 0 || len(res.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("empty response from Gemini")
	}

	// The SDK has already updated cs.History via SendMessage — save it back.
	chatsMu.Lock()
	ChatsInfoPerUser[info] = cs
	chatsMu.Unlock()

	resp := string(res.Candidates[0].Content.Parts[0].(genai.Text))
	return resp, nil
}
