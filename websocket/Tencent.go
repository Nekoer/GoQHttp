package websocket

import (
	"GoQHttp/logger"
	"GoQHttp/onebot"
	"GoQHttp/protocol/tencent/dto"
	"GoQHttp/utils"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Tencent struct {
}

type ATRequest struct {
	AppId        string `json:"appId"`
	ClientSecret string `json:"clientSecret"`
}

type ATResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   string `json:"expires_in"`
}

var ApiUrl = "https://api.sgroup.qq.com"

const SandboxApiUrl = "https://sandbox.api.sgroup.qq.com"

var accessToken string
var timestamp int64
var TencentAppId int
var TencentSecret string

func (t *Tencent) Init(appId int, secret string, sandbox bool) error {
	TencentAppId = appId
	TencentSecret = secret

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
	atRequest := ATRequest{
		AppId:        strconv.Itoa(TencentAppId),
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

	var atResponse ATResponse
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
	logger.Infof("QQ AccessToken: %s", accessToken)
	return nil
}

func (t *Tencent) Upload(file string) (string, error) {

	file = strings.Replace(file, "base64://", "", -1)
	decodeString, err := base64.StdEncoding.DecodeString(file)
	if err != nil {
		return "", err
	}

	// 创建一个缓冲区来存储multipart表单数据
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// 创建表单文件字段
	fileWriter, err := bodyWriter.CreateFormFile("image", "group.jpeg") // "uploadfile"是表单字段名，filename是服务器收到的文件名
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

	if uploadResponse.Success == false {
		return "", fmt.Errorf("图床上传失败")
	}

	return uploadResponse.Url, nil
}

// TODO {"message":"请求参数url无效","code":850028,"err_code":40011021,"trace_id":"47ac2894f7cdd2099fa7f924f31c46f6"}
func (t *Tencent) UploadFile(file string, groupId string) (*dto.RichMediaMsgResp, error) {

	//encoded, _ := base64.StdEncoding.DecodeString(file)
	upload, err := t.Upload(file)
	if err != nil {
		return nil, err
	}

	var groupRichMediaMessageToCreate = dto.GroupRichMediaMessageToCreate{
		FileType:   1,
		Url:        upload,
		SrvSendMsg: false,
		FileData:   nil,
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

	var richMediaMsgResp *dto.RichMediaMsgResp
	err = json.Unmarshal(body, &richMediaMsgResp)
	if err != nil {
		return nil, err
	}

	return richMediaMsgResp, err
}

func (t *Tencent) SendGroupMessage(data MessageRequest) {
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

	for i, element := range data.Message {
		logger.Debugf("%v ElementType: %+v,", i, element.ElementType)
		var groupMessageToCreate *dto.GroupMessageToCreate
		if element.ElementType == onebot.TextType {
			groupMessageToCreate = &dto.GroupMessageToCreate{
				Content:          element.Data.Text,
				MsgType:          dto.C2CMsgTypeText,
				Markdown:         nil,
				Keyboard:         nil,
				Media:            nil,
				Ark:              nil,
				Image:            "",
				MessageReference: nil,
				EventID:          "",
				MsgID:            MessageId,
				MsgReq:           1,
			}
		} else if element.ElementType == onebot.ImageType {
			file, err := t.UploadFile(element.Data.Image.File, GroupId)
			if err != nil {
				logger.Errorf("UploadFile err: %v", err)
				return
			}
			logger.Debugf("UploadFile : %+v", file)

			groupMessageToCreate = &dto.GroupMessageToCreate{
				Content:  " ",
				MsgType:  dto.C2CMsgTypeMedia,
				Markdown: nil,
				Keyboard: nil,
				Media: &dto.FileInfo{
					FileInfo: file.FileInfo,
				},
				Ark:              nil,
				Image:            "",
				MessageReference: nil,
				EventID:          "",
				MsgID:            MessageId,
				MsgReq:           1,
			}
		} else {
			logger.Warnf("暂不支持的消息类型: %s", element.ElementType)
		}
		logger.Warn(groupMessageToCreate)
		if groupMessageToCreate == nil {
			logger.Warnf("群消息构建失败")
			return
		}

		sendData, err := json.Marshal(groupMessageToCreate)
		if err != nil {
			logger.Errorf("json Marshal err: %v", err)
			return
		}
		r, err := http.NewRequest("POST", fmt.Sprintf("%s/v2/groups/%s/messages", ApiUrl, GroupId), bytes.NewBuffer(sendData))
		if err != nil {
			logger.Errorf("Error creating request: %v", err)
			return
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
			return
		}
		defer resp.Body.Close()

		// 读取响应
		body, _ := io.ReadAll(resp.Body)
		logger.Infof("SendGroupMessage %v", string(body))
	}

}
