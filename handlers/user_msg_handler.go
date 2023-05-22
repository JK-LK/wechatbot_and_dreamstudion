package handlers

import (
	"fmt"
	"github.com/869413421/wechatbot/config"
	"github.com/869413421/wechatbot/dreamstudio"
	"github.com/869413421/wechatbot/gpt"
	"github.com/eatmoreapple/openwechat"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

var _ MessageHandlerInterface = (*UserMessageHandler)(nil)

// UserMessageHandler 私聊消息处理
type UserMessageHandler struct {
	// 接收到消息
	msg *openwechat.Message
	// 发送的用户
	sender *openwechat.User
}

// handle 处理消息
func (g *UserMessageHandler) handle(msg *openwechat.Message) error {
	cfg := config.LoadConfig()
	//判断文本前缀是PictureToken，例如："生成图片"
	if strings.HasPrefix(g.msg.Content, cfg.PictureToken) {
		return g.ReplyImage()
	}
	//如果是纯文本，使用ChatGPT进行回复
	if g.msg.IsText() {
		return g.ReplyText(msg)
	}
	return nil
}

// NewUserMessageHandler 创建私聊处理器
func NewUserMessageHandler() MessageHandlerInterface {
	return &UserMessageHandler{}
}

// ReplyText 发送文本消息到群
func (g *UserMessageHandler) ReplyText(msg *openwechat.Message) error {
	// 接收私聊消息
	sender, err := msg.Sender()
	log.Printf("Received User %v Text Msg : %v", sender.NickName, msg.Content)

	// 向GPT发起请求
	requestText := strings.TrimSpace(msg.Content)
	requestText = strings.Trim(msg.Content, "\n")
	reply, err := gpt.Completions(requestText)
	if err != nil {
		log.Printf("gtp request error: %v \n", err)
		msg.ReplyText("机器人神了，我一会发现了就去修。")
		return err
	}
	if reply == "" {
		return nil
	}

	// 回复用户
	reply = strings.TrimSpace(reply)
	reply = strings.Trim(reply, "\n")
	_, err = msg.ReplyText(reply)
	if err != nil {
		log.Printf("response user error: %v \n", err)
	}
	return err
}

func (g *UserMessageHandler) ReplyImage() error {
	if time.Now().Unix()-g.msg.CreateTime > 60 {
		return nil
	}
	maxInt := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(5)
	time.Sleep(time.Duration(maxInt+1) * time.Second)

	log.Printf("Received User[%v], Content[%v], CreateTime[%v]", g.sender.NickName, g.msg.Content,
		time.Unix(g.msg.CreateTime, 0).Format("2006/01/02 15:04:05"))

	var (
		replyPath string
		err       error
	)

	cfg := config.LoadConfig()
	// 1.生成图片
	text := strings.Replace(g.msg.Content, cfg.PictureToken, "", -1)
	replyPath, err = dreamstudio.TextToImage(text)

	if err != nil {
		text := err.Error()
		if strings.Contains(err.Error(), "context deadline exceeded") {
			text = deadlineExceededText
		}
		_, err = g.msg.ReplyText(text)
		if err != nil {
			return fmt.Errorf("reply user error: %v ", err)
		}
		return err
	}

	//2.回复图片
	img, _ := os.Open(replyPath)
	defer img.Close()
	_, err = g.msg.ReplyImage(img)
	if err != nil {
		return fmt.Errorf("reply user error: %v ", err)
	}
	return err
}
