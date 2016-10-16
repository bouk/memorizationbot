package main

import (
	"gopkg.in/telegram-bot-api.v4"
)

const (
	Add                        = "➕"
	AddCard                    = "➕ New Card"
	AddDeck                    = "➕ New Deck"
	Back                       = "🔙"
	ChangeLocation             = "🌍 Set location"
	ChangeLocationFormat       = ChangeLocation + " (from %s)"
	ConfirmDeleteDeck          = "🔥 Yes"
	DeleteCard                 = "🗑 Delete"
	DeleteDeck                 = "🗑 Delete"
	Difficulty0                = "😮 No idea"
	Difficulty1                = "😣 Wrong"
	Difficulty2                = "🙂 Recalled"
	Difficulty3                = "☺️ Easy"
	DisableScheduling          = "🙅 Disable rehearsal"
	DontDeleteDeck             = "⛔️ No"
	EditCard                   = "📝 Edit Card"
	EditCardBack               = "✏️ Edit Back"
	EditCardFront              = "✏️ Edit Front"
	EditDeck                   = "✏️ Edit Deck"
	EditName                   = "✏️ Edit Name"
	EditSettings               = "🔧 Settings"
	ChangeTimeToRehearse       = "🕙 Set rehearsal time"
	ChangeTimeToRehearseFormat = ChangeTimeToRehearse + " (from %s)"
	EnableScheduling           = "💁 Enable rehearsal"
	Help                       = "🤔 Help"
	OK                         = "🆗"
	Save                       = "💾"
	ShowReverseOfCard          = "🔄 Show back"
)

var (
	CardReplyKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(Difficulty0),
			tgbotapi.NewKeyboardButton(Difficulty1),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(Difficulty2),
			tgbotapi.NewKeyboardButton(Difficulty3),
		),
	)
)

func init() {
	CardReplyKeyboard.OneTimeKeyboard = true
}
