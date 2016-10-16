package main

import (
	"log"
	"time"

	"github.com/getsentry/raven-go"
	"gopkg.in/telegram-bot-api.v4"
)

func poll() (bool, error) {
	tx, err := DB.Beginx()
	if err != nil {
		log.Print(err)
		tx.Rollback()
		return false, err
	}
	rows, err := tx.Queryx("SELECT user_id, card_id FROM scheduled_cards_to_send()")
	if err != nil {
		log.Print(err)
		tx.Rollback()
		return false, err
	}
	users := make([]int, 0, 20)
	cards := make([]int, 0, 20)
	for rows.Next() {
		var userID, cardID int
		err = rows.Scan(&userID, &cardID)
		if err != nil {
			log.Print(err)
			tx.Rollback()
			return false, err
		}
		users = append(users, userID)
		cards = append(cards, cardID)
	}

	for i := 0; i < len(users); i++ {
		userID := users[i]
		cardID := cards[i]
		card, err := GetCard(tx, cardID)
		if err != nil {
			log.Print(err)
			continue
		}

		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(Back),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(EditCard),
				tgbotapi.NewKeyboardButton(ShowReverseOfCard),
			),
		)
		keyboard.OneTimeKeyboard = true

		go func() {
			Send(tgbotapi.NewMessage(int64(userID), "Time for your rehearsal!"))
			card.SendFront(userID, keyboard)
		}()
	}
	tx.Commit()
	return len(users) > 0, nil
}

func Poller() {
	for {
		retry, err := poll()
		if err != nil {
			raven.CaptureError(err, nil)
		}
		if !retry {
			time.Sleep(10 * time.Second)
		}
	}
}
