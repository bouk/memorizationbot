package main

import (
	"gopkg.in/telegram-bot-api.v4"
)

type Message struct {
	Type MessageType `json:"t"`

	Text   string `json:"c,omitempty"`
	FileID string `json:"f,omitempty"`

	Latitude  float64 `json:"la,omitempty"`
	Longitude float64 `json:"lo,omitempty"`
}

func (m *Message) ToMessageConfig(chatID int64, replyMarkup interface{}) tgbotapi.Chattable {
	switch m.Type {
	case TextMessage:
		msg := tgbotapi.NewMessage(chatID, m.Text)
		msg.ReplyMarkup = replyMarkup
		return msg
	case PhotoMessage:
		photo := tgbotapi.NewPhotoShare(chatID, m.FileID)
		photo.Caption = m.Text
		photo.ReplyMarkup = replyMarkup
		return photo
	case AudioMessage:
		audio := tgbotapi.NewAudioShare(chatID, m.FileID)
		audio.ReplyMarkup = replyMarkup
		return audio
	case DocumentMessage:
		document := tgbotapi.NewDocumentShare(chatID, m.FileID)
		document.ReplyMarkup = replyMarkup
		return document
	case StickerMessage:
		sticker := tgbotapi.NewStickerShare(chatID, m.FileID)
		sticker.ReplyMarkup = replyMarkup
		return sticker
	case VideoMessage:
		video := tgbotapi.NewVideoShare(chatID, m.FileID)
		video.Caption = m.Text
		video.ReplyMarkup = replyMarkup
		return video
	case VoiceMessage:
		voice := tgbotapi.NewVoiceShare(chatID, m.FileID)
		voice.ReplyMarkup = replyMarkup
		return voice
	case LocationMessage:
		location := tgbotapi.NewLocation(chatID, m.Latitude, m.Longitude)
		location.ReplyMarkup = replyMarkup
		return location
	default:
		panic("Unknown message type")
	}
}

type MessageType int

const (
	TextMessage = MessageType(iota)
	PhotoMessage
	AudioMessage
	DocumentMessage
	StickerMessage
	VideoMessage
	VoiceMessage
	LocationMessage
)

func processMessage(msg *tgbotapi.Message, messages []Message) []Message {
	if msg.Audio != nil {
		return append(messages, Message{
			Type:   AudioMessage,
			Text:   msg.Caption,
			FileID: msg.Audio.FileID,
		})
	} else if msg.Document != nil {
		return append(messages, Message{
			Type:   DocumentMessage,
			Text:   msg.Caption,
			FileID: msg.Document.FileID,
		})
	} else if msg.Video != nil {
		return append(messages, Message{
			Type:   VideoMessage,
			Text:   msg.Caption,
			FileID: msg.Video.FileID,
		})
	} else if msg.Voice != nil {
		return append(messages, Message{
			Type:   VoiceMessage,
			Text:   msg.Caption,
			FileID: msg.Voice.FileID,
		})
	} else if msg.Sticker != nil {
		return append(messages, Message{
			Type:   StickerMessage,
			FileID: msg.Sticker.FileID,
		})
	} else if msg.Location != nil {
		return append(messages, Message{
			Type:      LocationMessage,
			Latitude:  msg.Location.Latitude,
			Longitude: msg.Location.Longitude,
		})
	} else if msg.Photo != nil {
		return append(messages, Message{
			Type:   PhotoMessage,
			Text:   msg.Caption,
			FileID: (*msg.Photo)[len(*msg.Photo)-1].FileID,
		})
	} else if msg.Text != "" {
		return append(messages, Message{
			Type: TextMessage,
			Text: msg.Text,
		})
	} else {
		return messages
	}
}
