package onebot

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ElementType string

const (
	TextType      ElementType = "text"
	FaceType      ElementType = "face"
	ImageType     ElementType = "image"
	RecordType    ElementType = "record"
	VideoType     ElementType = "video"
	AtType        ElementType = "at"
	RpsType       ElementType = "rps"
	DiceType      ElementType = "dice"
	ShakeType     ElementType = "shake"
	PokeType      ElementType = "poke"
	AnonymousType ElementType = "anonymous"
	ShareType     ElementType = "share"
	ContactType   ElementType = "contact"
	LocationType  ElementType = "location"
	MusicType     ElementType = "music"
	ReplyType     ElementType = "reply"
	ForwardType   ElementType = "forward"
	NodeType      ElementType = "node"
	XmlType       ElementType = "xml"
	JsonType      ElementType = "json"
)

type Element struct {
	ElementType ElementType `json:"type"`
	Data        *Message    `json:"data"`
}

type Message struct {
	Text        string      `json:"text,omitempty"`
	Face        Face        `json:"face,omitempty"`
	Image       Image       `json:"image,omitempty"`
	Record      Record      `json:"record,omitempty"`
	Video       Video       `json:"video,omitempty"`
	At          At          `json:"at,omitempty"`
	Rps         Rps         `json:"rps,omitempty"`
	Dice        Dice        `json:"dice,omitempty"`
	Shake       Shake       `json:"shake,omitempty"`
	Poke        Poke        `json:"poke,omitempty"`
	Anonymous   Anonymous   `json:"anonymous,omitempty"`
	Share       Share       `json:"share,omitempty"`
	Contact     Contact     `json:"contact,omitempty"`
	Location    Location    `json:"location,omitempty"`
	Music       Music       `json:"music,omitempty"`
	CustomMusic CustomMusic `json:"customMusic,omitempty"`
	Reply       Reply       `json:"reply,omitempty"`
	Forward     Forward     `json:"forward,omitempty"`
	Node        Node        `json:"node,omitempty"`
	MergeNode   MergeNode   `json:"mergeNode,omitempty"`
	Xml         Xml         `json:"xml,omitempty"`
	Json        Json        `json:"json,omitempty"`
}

// CQEscape 转义CQ码中的特殊字符
func CQEscape(s string) string {
	// 注意转义顺序：先转义&，避免重复转义
	replacements := map[string]string{
		"&": "&amp;",
		"[": "&#91;",
		"]": "&#93;",
		",": "&#44;",
	}

	// 按顺序替换，确保&最先被处理
	result := s
	for oldText, newText := range replacements {
		result = strings.ReplaceAll(result, oldText, newText)
	}

	return result
}

type Face struct {
	Id string `json:"id,omitempty"`
}

func (f Face) String() string {
	return fmt.Sprintf("[CQ:face,id=%s]", f.Id)
}

type Image struct {
	File    string `json:"file,omitempty"`
	Type    string `json:"type,omitempty"`
	Url     string `json:"url,omitempty"`
	Cache   int    `json:"cache,omitempty"`
	Proxy   int    `json:"proxy,omitempty"`
	TimeOut int    `json:"timeout,omitempty"`
}

func (f Image) String() string {
	f.File = CQEscape(f.File)
	f.Type = CQEscape(f.Type)
	f.Url = CQEscape(f.Url)
	return fmt.Sprintf("[CQ:image,file=%s,type=%s,url=%s,cache=%v,proxy=%v,timeout=%v]", f.File, f.Type, f.Url, f.Cache, f.Proxy, f.TimeOut)
}

type Record struct {
	File    string `json:"file,omitempty"`
	Magic   int    `json:"magic,omitempty"`
	Url     string `json:"url,omitempty"`
	Cache   int    `json:"cache,omitempty"`
	Proxy   int    `json:"proxy,omitempty"`
	TimeOut int    `json:"timeout,omitempty"`
}

func (f Record) String() string {
	f.File = CQEscape(f.File)
	f.Url = CQEscape(f.Url)

	return fmt.Sprintf("[CQ:record,file=%s,magic=%s,url=%s,cache=%v,proxy=%v,timeout=%v]", f.File, f.Magic, f.Url, f.Cache, f.Proxy, f.TimeOut)
}

type Video struct {
	File    string `json:"file,omitempty"`
	Url     string `json:"url,omitempty"`
	Cache   int    `json:"cache,omitempty"`
	Proxy   int    `json:"proxy,omitempty"`
	TimeOut int    `json:"timeout,omitempty"`
}

func (f Video) String() string {
	f.File = CQEscape(f.File)
	f.Url = CQEscape(f.Url)

	return fmt.Sprintf("[CQ:video,file=%s,url=%s,cache=%v,proxy=%v,timeout=%v]", f.File, f.Url, f.Cache, f.Proxy, f.TimeOut)
}

type At struct {
	Uid string `json:"qq,omitempty"`
}

func (f At) String() string {
	return fmt.Sprintf("[CQ:at,qq=%s]", f.Uid)
}

type Rps struct {
}

func (f Rps) String() string {
	return "[CQ:rps]"
}

type Dice struct {
}

func (f Dice) String() string {
	return "[CQ:dice]"
}

type Shake struct {
}

func (f Shake) String() string {
	return "[CQ:shake]"
}

type Poke struct {
	Type string `json:"type,omitempty"`
	Id   string `json:"id,omitempty"`
}

func (f Poke) String() string {
	return fmt.Sprintf("[CQ:poke,type=%s,id=%s]", f.Type, f.Id)
}

type Anonymous struct {
	Ignore int `json:"ignore,omitempty"`
}

func (f Anonymous) String() string {
	return fmt.Sprintf("[CQ:anonymous,ignore=%v]", f.Ignore)
}

type Share struct {
	Url     string `json:"url,omitempty"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
	Image   string `json:"image,omitempty"`
}

func (f Share) String() string {
	f.Title = CQEscape(f.Title)
	f.Url = CQEscape(f.Url)
	f.Content = CQEscape(f.Content)
	f.Image = CQEscape(f.Image)
	return fmt.Sprintf("[CQ:share,url=%s,title=%s,content=%s,image=%s]", f.Url, f.Title, f.Content, f.Image)
}

type Contact struct {
	Type string `json:"type,omitempty"`
	Id   string `json:"id,omitempty"`
}

func (f Contact) String() string {
	return fmt.Sprintf("[CQ:contact,id=%s,type=%s]", f.Id, f.Type)
}

type Location struct {
	Lat     string `json:"lat,omitempty"`
	Lon     string `json:"lon,omitempty"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
}

func (f Location) String() string {
	f.Title = CQEscape(f.Title)
	f.Lat = CQEscape(f.Lat)
	f.Content = CQEscape(f.Content)
	f.Lon = CQEscape(f.Lon)
	return fmt.Sprintf("[CQ:location,lat=%s,lon=%s,title=%s,content=%s]", f.Lat, f.Lon, f.Title, f.Content)
}

type Music struct {
	Type string `json:"type,omitempty"`
	Id   string `json:"id,omitempty"`
}

func (f Music) String() string {
	return fmt.Sprintf("[CQ:music,id=%s,type=%s]", f.Id, f.Type)
}

type CustomMusic struct {
	Type    string `json:"type,omitempty"`
	Url     string `json:"url,omitempty"`
	Audio   string `json:"audio,omitempty"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
	Image   string `json:"image,omitempty"`
}

func (f CustomMusic) String() string {
	f.Title = CQEscape(f.Title)
	f.Type = CQEscape(f.Type)
	f.Content = CQEscape(f.Content)
	f.Url = CQEscape(f.Url)
	f.Audio = CQEscape(f.Audio)
	f.Image = CQEscape(f.Image)
	return fmt.Sprintf("[CQ:music,type=%s,url=%s,audio=%s,title=%s,content=%s,image=%s]", f.Type, f.Url, f.Audio, f.Title, f.Content, f.Image)
}

type Reply struct {
	Id string `json:"id,omitempty"`
}

func (f Reply) String() string {
	return fmt.Sprintf("[CQ:reply,id=%s]", f.Id)
}

type Forward struct {
	Id string `json:"id,omitempty"`
}

func (f Forward) String() string {
	return fmt.Sprintf("[CQ:forward,id=%s]", f.Id)
}

type Node struct {
	Id string `json:"id,omitempty"`
}

func (f Node) String() string {
	return fmt.Sprintf("[CQ:node,id=%s]", f.Id)
}

type MergeNode struct {
	UserId   string      `json:"user_id,omitempty"`
	NickName string      `json:"nick_name,omitempty"`
	Content  interface{} `json:"content,omitempty"`
}

func (f MergeNode) String() string {
	f.NickName = CQEscape(f.NickName)
	marshal, err := json.Marshal(f.Content)
	if err == nil {
		f.Content = CQEscape(string(marshal))
	}

	return fmt.Sprintf("[CQ:node,user_id=%s,nick_name=%s,content=%s]", f.UserId, f.NickName, f.Content)
}

type Xml struct {
	Data string `json:"data,omitempty"`
}

func (f Xml) String() string {
	return fmt.Sprintf("[CQ:xml,data=%s]", f.Data)
}

type Json struct {
	Data string `json:"data,omitempty"`
}

func (f Json) String() string {
	return fmt.Sprintf("[CQ:json,data=%s]", f.Data)
}
