package gpt

import (
	"context"
	"errors"
	"github.com/869413421/wechatbot/config"
	"github.com/sashabaranov/go-openai"
	"log"
)

func Completions(msg []openai.ChatCompletionMessage) (string, error) {
	cfg := config.LoadConfig()
	if cfg.ApiKey == "" {
		log.Printf("GPT api key required\n")
		return "", errors.New("GPT api key required")
	}
	var client = openai.NewClient(cfg.ApiKey)
	log.Printf("Request already send")

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:            cfg.Model,
			Messages:         msg,
			MaxTokens:        1024,
			Temperature:      1,
			TopP:             1,
			FrequencyPenalty: 0,
			PresencePenalty:  0,
		},
	)
	if err != nil {
		log.Printf("ChatCompletion error: %v\n", err)
		return "", err
	}
	content := resp.Choices[0].Message.Content
	log.Printf("GPT Response: %s\n", content)
	return content, nil

}
