package handlers

import (
	"fmt"
	"github.com/869413421/wechatbot/config"
	"github.com/869413421/wechatbot/dreamstudio"
	"github.com/869413421/wechatbot/gpt"
	"github.com/eatmoreapple/openwechat"
	"github.com/sashabaranov/go-openai"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

var _ MessageHandlerInterface = (*GroupMessageHandler)(nil)

type GroupMessageHandler struct {
	// 获取自己
	self *openwechat.Self
	// 群
	group *openwechat.Group
	// 接收到消息
	msg *openwechat.Message
	// 发送的用户
	sender *openwechat.User
}

// handle 处理消息
func (g *GroupMessageHandler) handle(msg *openwechat.Message) error {
	cfg := config.LoadConfig()
	// 判断文本前缀是PictureToken，例如："生成图片"
	if strings.Contains(msg.Content, cfg.PictureToken) {
		log.Printf("开始生成图片%s", cfg.PictureToken)
		return g.ReplyImage()
	}
	//如果是纯文本，使用ChatGPT进行回复
	if msg.IsText() {
		log.Printf("微信内容:%s", msg.Content)
		return g.ReplyText(msg)
	}
	return nil
}

func NewGroupMessageHandler() MessageHandlerInterface {
	return &GroupMessageHandler{}
}

func (g *GroupMessageHandler) getRequestText(msg *openwechat.Message) []openai.ChatCompletionMessage {
	// 1.去除空格以及换行
	requestText := strings.TrimSpace(msg.Content)
	requestText = strings.Trim(msg.Content, "\n")
	if len(requestText) == 0 {
		log.Println("user message is empty")
		sessionText := make([]openai.ChatCompletionMessage, 0)
		return sessionText
	}

	// 2.替换掉当前用户名称
	replaceText := "@" + g.self.NickName
	requestText = strings.TrimSpace(strings.ReplaceAll(g.msg.Content, replaceText, ""))
	if len(requestText) == 0 {
		log.Println("user message is empty")
		sessionText := make([]openai.ChatCompletionMessage, 0)
		return sessionText
	}

	sessionText := make([]openai.ChatCompletionMessage, 0)
	sessionText = append(sessionText, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: requestText,
	})
	return sessionText

}

func (g *GroupMessageHandler) ReplyText(msg *openwechat.Message) error {
	// 接收群消息
	log.Printf("Received Group %v Text Msg : %v", g.group.NickName, msg.Content)

	// 不是@的不处理
	if !msg.IsAt() {
		return nil
	}

	// 替换掉@文本，然后向GPT发起请求
	requestText := g.getRequestText(msg)
	reply, err := gpt.Completions(requestText)
	if err != nil {
		log.Printf("gtp request error: %v \n", err)
		msg.ReplyText("机器人神了，我一会发现了就去修。")
		return err
	}
	if reply == "" {
		return nil
	}

	// 获取@我的用户
	groupSender, err := msg.SenderInGroup()
	if err != nil {
		log.Printf("get sender in group error :%v \n", err)
		return err
	}

	// 回复@我的用户
	reply = strings.TrimSpace(reply)
	reply = strings.Trim(reply, "\n")
	atText := "@" + groupSender.NickName
	replyText := atText + reply
	_, err = msg.ReplyText(replyText)
	if err != nil {
		log.Printf("response group error: %v \n", err)
	}
	return err
}

func (g *GroupMessageHandler) ReplyImage() error {
	if time.Now().Unix()-g.msg.CreateTime > 60 {
		return nil
	}

	maxInt := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(5)
	time.Sleep(time.Duration(maxInt+1) * time.Second)

	log.Printf("Received Group[%v], Content[%v], CreateTime[%v]", g.group.NickName, g.msg.Content,
		time.Unix(g.msg.CreateTime, 0).Format("2006/01/02 15:04:05"))

	var (
		replyPath string
		err       error
	)
	// 1.不是@的不处理
	if !g.msg.IsAt() {
		return nil
	}
	// 2.整理数据
	cfg := config.LoadConfig()
	text := strings.ReplaceAll(g.msg.Content, cfg.PictureToken, "")
	replaceText := "@" + g.self.NickName
	text = strings.ReplaceAll(text, replaceText, "")
	if text == "" {
		return nil
	}
	// 3.请求图片
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

	// 4.回复图片
	img, _ := os.Open(replyPath)
	defer img.Close()
	_, err = g.msg.ReplyImage(img)
	if err != nil {
		return fmt.Errorf("reply user error: %v ", err)
	}
	return err
}
