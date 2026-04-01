package gemini

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid"
	"google.golang.org/api/option"
)

const (
	googleApiKeyEnv = "GOOGLE_API_KEY"
	generativeModel = "gemini-1.5-flash"

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

func InitiateChat(info ChatInfo, msg string) (string, error) {
	ctx := context.Background()

	apiKey := os.Getenv(googleApiKeyEnv)
	if apiKey == "" {
		return "", errors.New("GOOGLE_API_KEY environment variable is not set")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		// Return error instead of log.Fatal so we don't crash the server.
		return "", fmt.Errorf("failed to create Gemini client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel(generativeModel)
	cs := model.StartChat()

	cs.History = append(
		cs.History,
		&genai.Content{Parts: []genai.Part{genai.Text(msg)}, Role: roleUser},
	)

	res, err := cs.SendMessage(ctx, genai.Text(msg))
	if err != nil {
		log.Println("Gemini SendMessage error:", err)
		return "", fmt.Errorf("gemini error: %w", err)
	}

	if len(res.Candidates) == 0 || len(res.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("empty response from Gemini")
	}

	cs.History = append(
		cs.History,
		&genai.Content{Parts: res.Candidates[0].Content.Parts, Role: roleModel},
	)

	chatsMu.Lock()
	ChatsInfoPerUser[info] = cs
	chatsMu.Unlock()

	// Use type assertion to get plain string — NOT %#v which produces Go syntax.
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

	// Correct nil check — comparing the pointer value, not address-of-pointer.
	if chatSession == nil {
		log.Println("chatSession is nil for info:", info)
		return "", errors.New("chat session is nil")
	}

	log.Println("Sending follow-up message to Gemini, history length:", len(chatSession.History))

	apiKey := os.Getenv(googleApiKeyEnv)
	if apiKey == "" {
		return "", errors.New("GOOGLE_API_KEY environment variable is not set")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		// Return error instead of log.Fatal so we don't crash the server.
		return "", fmt.Errorf("failed to create Gemini client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel(generativeModel)
	cs := model.StartChat()
	cs.History = chatSession.History

	res, err := cs.SendMessage(ctx, genai.Text(msg))
	if err != nil {
		return "", fmt.Errorf("gemini error: %w", err)
	}

	if len(res.Candidates) == 0 || len(res.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("empty response from Gemini")
	}

	chatSession.History = append(
		chatSession.History,
		&genai.Content{Parts: []genai.Part{genai.Text(msg)}, Role: roleUser},
	)
	chatSession.History = append(
		chatSession.History,
		&genai.Content{Parts: res.Candidates[0].Content.Parts, Role: roleModel},
	)

	chatsMu.Lock()
	ChatsInfoPerUser[info] = chatSession
	chatsMu.Unlock()

	// Use type assertion to get plain string — NOT %#v which produces Go syntax.
	resp := string(res.Candidates[0].Content.Parts[0].(genai.Text))

	return resp, nil
}
