package dreamstudio

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/869413421/wechatbot/config"
	"log"
	"net/http"
	"os"
)

type TextToImageImage struct {
	Base64       string `json:"base64"`
	Seed         string `json:"seed"`
	FinishReason string `json:"finishReason"`
}

type TextToImageResponse struct {
	Images []TextToImageImage `json:"artifacts"`
}

type TextPrompt struct {
	Text   string  `json:"text"`
	Weight float64 `json:"weight"`
}

// DreamStdioRequestBody 请求体
type DreamStdioRequestBody struct {
	TextPrompts        []TextPrompt `json:"textPrompts"`
	CfgScale           uint         `json:"cfgScale"`
	ClipGuidancePreset string       `json:"clipGuidancePreset"`
	Height             uint         `json:"height"`
	Width              uint         `json:"width"`
	Samples            uint         `json:"samples"`
	Steps              uint         `json:"steps"`
}

func TextToImage(msg string) (string, error) {
	cfg := config.LoadConfig()
	engineId := cfg.EngineId
	apiHost, hasApiHost := os.LookupEnv("API_HOST")
	if !hasApiHost {
		apiHost = "https://api.stability.ai"
	}
	reqUrl := apiHost + "/v1/generation/" + engineId + "/text-to-image"

	textPrompts := []TextPrompt{
		{
			Text:   msg,
			Weight: 1,
		},
	}

	requestBody := DreamStdioRequestBody{
		TextPrompts:        textPrompts,
		CfgScale:           cfg.CfgScale,
		ClipGuidancePreset: "FAST_BLUE",
		Height:             cfg.PicHeight,
		Width:              cfg.PicWidth,
		Samples:            1,
		Steps:              cfg.Steps,
	}
	requestData, _ := json.Marshal(requestBody)
	log.Printf("dreamstudio request(%d) json:%s\n", 1, string(requestData))

	req, _ := http.NewRequest("POST", reqUrl, bytes.NewBuffer(requestData))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+cfg.DreamStdioApiKey)

	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()

	log.Printf("code:%d", res.StatusCode)

	//当HTTP响应状态码不为200时，打印出响应主体的内容以供调试。
	if res.StatusCode != 200 {
		var body map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
			panic(err)
		}
		log.Printf("Non-200 requese: %s", body)
	}

	var body TextToImageResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		log.Printf("decode json error: %v", err)
	}

	for i, image := range body.Images {
		outFile := fmt.Sprintf("v1_txt2img_%d.png", i)
		file, err := os.Create(outFile)
		if err != nil {
			log.Printf("picture create error: %v", err)
		}

		imageBytes, err := base64.StdEncoding.DecodeString(image.Base64)
		if err != nil {
			log.Printf("picture create error: %v", err)
		}

		if _, err := file.Write(imageBytes); err != nil {
			log.Printf("picture write error: %v", err)
		}

		if err := file.Close(); err != nil {
			log.Printf("picture close error: %v", err)
		}
	}

	return "v1_txt2img_0.png", nil

}
