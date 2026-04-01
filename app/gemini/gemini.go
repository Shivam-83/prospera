package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	// Keep the same environment variable name so you don't have to change it in Render,
	// but paste your "sk-or-v1-..." OpenRouter key into it.
	googleApiKeyEnv = "GOOGLE_API_KEY"
	openRouterURL   = "https://openrouter.ai/api/v1/chat/completions"

	// Using the highly reliable, fully free Llama 3.3 70B model from Meta.
	primaryModel = "meta-llama/llama-3.3-70b-instruct:free"
)

type ChatInfo struct {
	userID    string
	sessionID string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatSession struct {
	History []Message
}

type OpenRouterRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type OpenRouterResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

var (
	// ChatsInfoPerUser holds OpenRouter manual memory instances instead of genai.ChatSession
	ChatsInfoPerUser = map[ChatInfo]*ChatSession{}
	chatsMu          sync.RWMutex
	httpClient       = &http.Client{Timeout: 30 * time.Second}
)

func NewChatInfo(userID string) ChatInfo {
	chatInfo := ChatInfo{userID: userID, sessionID: uuid.NewString()}

	chatsMu.Lock()
	ChatsInfoPerUser[chatInfo] = &ChatSession{
		History: []Message{},
	}
	chatsMu.Unlock()

	return chatInfo
}

func sendToOpenRouter(ctx context.Context, history []Message) (Message, error) {
	apiKey := os.Getenv(googleApiKeyEnv)
	if apiKey == "" {
		return Message{}, errors.New("GOOGLE_API_KEY environment variable is not set with OpenRouter key")
	}

	reqBody := OpenRouterRequest{
		Model:    primaryModel,
		Messages: history,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return Message{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", openRouterURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return Message{}, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://prospera-bnny.onrender.com")
	req.Header.Set("X-Title", "Prospera AI Coach")

	resp, err := httpClient.Do(req)
	if err != nil {
		return Message{}, fmt.Errorf("openrouter request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return Message{}, fmt.Errorf("openrouter returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var orResp OpenRouterResponse
	if err := json.Unmarshal(bodyBytes, &orResp); err != nil {
		return Message{}, fmt.Errorf("failed to parse openrouter response: %w (%s)", err, string(bodyBytes))
	}

	if len(orResp.Choices) == 0 {
		return Message{}, errors.New("empty choices from openrouter")
	}

	return orResp.Choices[0].Message, nil
}

func InitiateChat(info ChatInfo, msg string) (string, error) {
	ctx := context.Background()

	// 1. Initial history with the user's message
	history := []Message{
		{Role: "user", Content: msg},
	}

	// 2. Send to OpenRouter
	aiMsg, err := sendToOpenRouter(ctx, history)
	if err != nil {
		log.Println("OpenRouter InitiateChat error:", err)
		return "", err
	}

	// 3. Append the AI's response to the history
	history = append(history, aiMsg)

	// 4. Save session
	chatsMu.Lock()
	ChatsInfoPerUser[info] = &ChatSession{History: history}
	chatsMu.Unlock()

	return aiMsg.Content, nil
}

func SendMessage(ctx context.Context, info ChatInfo, msg string) (string, error) {
	chatsMu.RLock()
	chatSession, ok := ChatsInfoPerUser[info]
	chatsMu.RUnlock()

	if !ok || chatSession == nil {
		return "", errors.New("no chat session found")
	}

	// 1. Append new user message to existing history
	chatSession.History = append(chatSession.History, Message{Role: "user", Content: msg})

	// 2. Send full history to OpenRouter
	aiMsg, err := sendToOpenRouter(ctx, chatSession.History)
	if err != nil {
		log.Println("OpenRouter SendMessage error:", err)
		// Revert the user message addition if the API call failed so we don't skew the history
		chatSession.History = chatSession.History[:len(chatSession.History)-1]
		return "", err
	}

	// 3. Append AI response to history and save
	chatSession.History = append(chatSession.History, aiMsg)

	chatsMu.Lock()
	ChatsInfoPerUser[info] = chatSession
	chatsMu.Unlock()

	return aiMsg.Content, nil
}
