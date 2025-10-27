package openapi

import (
	"GoQHttp/internal/onebot"
	"GoQHttp/internal/protocol"
	dto2 "GoQHttp/internal/protocol/tencent/dto"
	"GoQHttp/logger"
	"GoQHttp/utils"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type OpenApi struct {
}

var ApiUrl = "https://api.sgroup.qq.com"

const SandboxApiUrl = "https://sandbox.api.sgroup.qq.com"

var accessToken string
var timestamp int64
var TencentAppId int
var TencentSecret string
var asyncId int64

func (o *OpenApi) Init(appId int, secret string, sandbox bool) error {
	TencentAppId = appId
	TencentSecret = secret
	asyncId = 0
	if sandbox {
		ApiUrl = SandboxApiUrl
	}

	err := getAppAccessToken()
	if err != nil {
		return err
	}

	return nil
}

func getAppAccessToken() error {
	atRequest := dto2.GetAccessTokenReq{
		AppID:        strconv.Itoa(TencentAppId),
		ClientSecret: TencentSecret,
	}
	data, err := json.Marshal(atRequest)
	if err != nil {
		panic(err)
	}
	r, err := http.Post("https://bots.qq.com/app/getAppAccessToken", "application/json", bytes.NewBuffer(data))
	if err != nil {
		logger.Errorf("获取官Q调用凭证失败 %v", err)
		return nil
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	var atResponse dto2.GetAccessTokenResp
	err = json.Unmarshal(body, &atResponse)
	if err != nil {
		return err
	}

	accessToken = atResponse.AccessToken
	expiresIn, err := strconv.Atoi(atResponse.ExpiresIn)
	if err != nil {
		return err
	}

	timestamp = int64(expiresIn) + time.Now().Unix()
	logger.Infof("QQ凭证获取成功: %s", accessToken)
	return nil
}

func (o *OpenApi) NextAsyncID() int64 {
	asyncId++
	newID := asyncId

	if newID > 1000000 {
		asyncId = rand.Int63n(1060000)
	}
	logger.Debugf("AsyncID: %v", newID)
	return newID
}

func IsImageURL(url string) bool {
	imageExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg"}
	lowerURL := strings.ToLower(url)

	for _, ext := range imageExtensions {
		if strings.HasSuffix(lowerURL, ext) {
			return true
		}
	}
	return false
}

// Upload 第三方上传群图片方式
func (o *OpenApi) Upload(file string) (string, error) {

	file = strings.Replace(file, "base64://", "", -1)
	decodeString, err := base64.StdEncoding.DecodeString(file)
	if err != nil {
		return "", err
	}

	// 创建一个缓冲区来存储multipart表单数据
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// 创建表单文件字段
	fileWriter, err := bodyWriter.CreateFormFile("image", "group.jpeg")
	if err != nil {
		return "", fmt.Errorf("创建表单文件字段失败: %v", err)
	}

	// 将[]byte数据写入表单文件字段
	_, err = io.Copy(fileWriter, bytes.NewReader(decodeString))
	if err != nil {
		return "", fmt.Errorf("写入文件数据失败: %v", err)
	}

	// 关闭multipart写入器以设置尾部边界
	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	// 发送POST请求
	resp, err := http.Post("https://img.scdn.io/api/v1.php", contentType, bodyBuf)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取并打印响应（可选）
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	type UploadResponse struct {
		Success bool   `json:"success"`
		Url     string `json:"url"`
		Message string `json:"message"`
	}
	var uploadResponse UploadResponse
	err = json.Unmarshal(respBody, &uploadResponse)
	if err != nil {
		return "", err
	}
	logger.Errorf("%+v", uploadResponse)

	if uploadResponse.Success == false {
		return "", fmt.Errorf("图床上传失败")
	}

	return uploadResponse.Url, nil
}

// UploadFile 本地上传base64图片
func (o *OpenApi) UploadFile(file string, groupId string) (*dto2.RichMediaMsgResp, error) {

	//encoded, _ := base64.StdEncoding.DecodeString(file)
	//upload, err := o.Upload(file)
	//if err != nil {
	//	return nil, err
	//}
	var groupRichMediaMessageToCreate dto2.GroupRichMediaMessageToCreate
	if IsImageURL(file) {
		groupRichMediaMessageToCreate = dto2.GroupRichMediaMessageToCreate{
			Url:        file,
			FileType:   1,
			SrvSendMsg: false,
		}
	} else {
		tmpFile := strings.Replace(file, "base64://", "", -1)
		groupRichMediaMessageToCreate = dto2.GroupRichMediaMessageToCreate{
			FileType:   1,
			SrvSendMsg: false,
			FileData:   tmpFile,
		}
	}

	sendData, err := json.Marshal(groupRichMediaMessageToCreate)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest("POST", fmt.Sprintf("%s/v2/groups/%s/files", ApiUrl, groupId), bytes.NewBuffer(sendData))
	if err != nil {
		logger.Errorf("Error creating request: %v", err)
		return nil, err
	}

	r.Header = http.Header{
		"Content-Type":  []string{"application/json"},
		"Authorization": []string{"QQBot " + accessToken},
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		logger.Errorf("Error sending request: %v", err)
		return nil, err
	}

	defer resp.Body.Close()

	// 读取响应
	body, _ := io.ReadAll(resp.Body)

	var richMediaMsgResp *dto2.RichMediaMsgResp
	err = json.Unmarshal(body, &richMediaMsgResp)
	if err != nil {
		return nil, err
	}

	return richMediaMsgResp, err
}

func (o *OpenApi) SendPacket() {
	for payload := range protocol.QQOfficialChan {
		switch payload.MessageType {
		case onebot.GroupMessage:
			o.SendGroupMessage(payload)
		default:
			marshal, err := json.Marshal(payload)
			if err != nil {
				logger.Errorf("SendPacket json: %v", err)
				return
			}
			logger.Warnf("暂不支持的数据包: %+v", marshal)
		}

	}
}

func (o *OpenApi) SendGroupMessage(data *onebot.MessageRequest) {
	if time.Now().Unix() <= timestamp {
		err := getAppAccessToken()
		if err != nil {
			logger.Errorf("GetAppAccessToken err: %v", err)
			return
		}
	}
	GroupId, err := utils.DBUtil.GetGroupID(data.GroupId)
	if err != nil {
		logger.Errorf("DBUtil.GetGroupID err: %v", err)
		return
	}

	MessageId, err := utils.DBUtil.GetGroupMessageID(data.GroupId, data.UserId)

	if err != nil {
		logger.Errorf("DBUtil.GetGroupMessageID err: %v", err)
		return
	}
	for _, element := range data.Message {
		var groupMessageToCreate *dto2.GroupMessageToCreate
		seq := o.NextAsyncID()
		if element.ElementType == onebot.TextType {
			marshal, err := json.Marshal(element.Data)
			if err != nil {
				continue
			}
			var text onebot.Text
			err = json.Unmarshal(marshal, &text)
			if err != nil {
				continue
			}
			// 判断是否是单空格内容
			var tempStr = strings.TrimSpace(text.Text)
			if tempStr == "" {
				continue
			}

			groupMessageToCreate = &dto2.GroupMessageToCreate{
				Content:          text.Text,
				MsgType:          dto2.C2CMsgTypeText,
				Markdown:         nil,
				Keyboard:         nil,
				Media:            nil,
				Ark:              nil,
				Image:            "",
				MessageReference: nil,
				EventID:          "",
				MsgID:            MessageId,
				MsgReq:           uint(seq),
			}
		} else if element.ElementType == onebot.ImageType {
			marshal, err := json.Marshal(element.Data)
			if err != nil {
				continue
			}
			var image onebot.Image
			err = json.Unmarshal(marshal, &image)
			if err != nil {
				continue
			}
			file, err := o.UploadFile(image.File, GroupId)
			if err != nil {
				logger.Errorf("UploadFile err: %v", err)
				continue
			}
			logger.Debugf("UploadFile : %+v", file)

			groupMessageToCreate = &dto2.GroupMessageToCreate{
				Content:  " ",
				MsgType:  dto2.C2CMsgTypeMedia,
				Markdown: nil,
				Keyboard: nil,
				Media: &dto2.FileInfo{
					FileInfo: file.FileInfo,
				},
				Ark:              nil,
				Image:            "",
				MessageReference: nil,
				EventID:          "",
				MsgID:            MessageId,
				MsgReq:           uint(seq),
			}
		} else {
			logger.Warnf("暂不支持的消息类型: %s", element.ElementType)
			continue
		}

		if groupMessageToCreate == nil {
			logger.Warnf("群消息构建失败")
			continue
		}

		sendData, err := json.Marshal(groupMessageToCreate)
		if err != nil {
			logger.Errorf("json Marshal err: %v", err)
			continue
		}
		r, err := http.NewRequest("POST", fmt.Sprintf("%s/v2/groups/%s/messages", ApiUrl, GroupId), bytes.NewBuffer(sendData))
		if err != nil {
			logger.Errorf("Error creating request: %v", err)
			continue
		}

		r.Header = http.Header{
			"Content-Type":  []string{"application/json"},
			"Authorization": []string{"QQBot " + accessToken},
		}

		// 发送请求
		client := &http.Client{}
		resp, err := client.Do(r)
		if err != nil {
			logger.Errorf("Error sending request: %v", err)
			continue
		}
		defer resp.Body.Close()

		// 读取响应
		//_, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			logger.Warnf("发送群消息失败: %v", string(body))
		} else {
			logger.Info("发送群消息成功")
		}
	}

}

func (o *OpenApi) GetGuild(guildId string) (*dto2.Guild, error) {
	r, err := http.NewRequest("GET", fmt.Sprintf("%s/guilds/%s", ApiUrl, guildId), nil)
	if err != nil {
		logger.Errorf("Error creating request: %v", err)
		return nil, err
	}

	r.Header = http.Header{
		"Content-Type":  []string{"application/json"},
		"Authorization": []string{"QQBot " + accessToken},
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		logger.Errorf("Error sending request: %v", err)
		return nil, err
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	logger.Debugf("%v", string(body))
	var guild *dto2.Guild
	err = json.Unmarshal(body, &guild)

	if err != nil {
		logger.Errorf("json Unmarshal err: %v", err)
		return nil, err
	}

	return guild, nil
}

func (o *OpenApi) GetChannel(channelId string) (*dto2.Channel, error) {
	r, err := http.NewRequest("GET", fmt.Sprintf("%s/channels/%s", ApiUrl, channelId), nil)
	if err != nil {
		logger.Errorf("Error creating request: %v", err)
		return nil, err
	}

	r.Header = http.Header{
		"Content-Type":  []string{"application/json"},
		"Authorization": []string{"QQBot " + accessToken},
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		logger.Errorf("Error sending request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	logger.Debugf("%v", string(body))
	var channel *dto2.Channel
	err = json.Unmarshal(body, &channel)

	if err != nil {
		logger.Errorf("json Unmarshal err: %v", err)
		return nil, err
	}

	return channel, nil
}
