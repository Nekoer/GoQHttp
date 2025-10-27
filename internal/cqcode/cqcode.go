package cqcode

import (
	"GoQHttp/internal/onebot"
	"GoQHttp/internal/protocol/tencent/dto"
	"fmt"
	"regexp"
	"strings"
)

// CQCode 表示一个CQ码
type CQCode struct {
	Type   string            // CQ码类型，如"at", "image"
	Params map[string]string // 参数字典
	Raw    string            // 原始字符串
}

// ParseCQCode 解析CQ码字符串，返回CQCode结构体
func (c *CQCode) ParseCQCode(s string) (*CQCode, error) {
	// 匹配CQ码的正则表达式
	pattern := `\[CQ:([^,\]]+)(?:,([^,\]]+=[^,\]]+))*\]`
	re := regexp.MustCompile(pattern)

	match := re.FindStringSubmatch(s)
	if match == nil {
		return nil, fmt.Errorf("无效的CQ码格式: %s", s)
	}

	cq := &CQCode{
		Type:   match[1],
		Params: make(map[string]string),
		Raw:    s,
	}

	// 解析参数
	if len(match) > 2 {
		paramStr := match[2]
		// 处理多个参数的情况
		params := strings.Split(paramStr, ",")
		for _, param := range params {
			parts := strings.SplitN(param, "=", 2)
			if len(parts) == 2 {
				key := parts[0]
				value := parts[1]
				cq.Params[key] = value
			}
		}
	}

	return cq, nil
}

// ParseAllCQCodes 从文本中提取并解析所有CQ码并返回Element元素
func (c *CQCode) ParseAllCQCodes(text string) ([]*onebot.Element, error) {
	pattern := `\[CQ:[^\]]+\]`
	re := regexp.MustCompile(pattern)

	matches := re.FindAllString(text, -1)
	if matches == nil {
		return nil, fmt.Errorf("未找到CQ码")
	}

	var cqCodes []*CQCode
	for _, match := range matches {
		cq, err := c.ParseCQCode(match)
		if err != nil {
			return nil, err
		}
		cqCodes = append(cqCodes, cq)
	}

	var messages []*onebot.Element
	for _, code := range cqCodes {
		//logger.Debugf("%+v,%+v,%+v", code.Raw, code.Type, code.Params)
		switch code.Type {
		case "text":
			{
				text := code.Params["text"]
				messages = append(messages, &onebot.Element{
					ElementType: onebot.TextType,
					Data: onebot.Text{
						Text: text,
					},
				})
			}
		case "image":
			{
				file := code.Params["file"]
				messages = append(messages, &onebot.Element{
					ElementType: onebot.ImageType,
					Data: onebot.Image{
						File: file,
					},
				})
			}
		}
	}
	return messages, nil
}

func (c *CQCode) BuildCQCodeFromAttachments(data []*dto.MessageAttachment) ([]string, error) {
	var attachments []string
	for _, messageAttachment := range data {
		image := onebot.Image{
			File:    messageAttachment.URL,
			Type:    "",
			Url:     messageAttachment.URL,
			Cache:   0,
			Proxy:   0,
			TimeOut: 0,
		}
		attachments = append(attachments, image.String())
	}
	return attachments, nil
}
