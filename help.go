package main

import (
	"time"

	"gopkg.in/telegram-bot-api.v4"
)

func HelpUser(id int64) {
	msg := func(text string) {
		Send(tgbotapi.NewMessage(id, text))
	}
	msg("To use Memorization Bot, you're going to create some flash cards!")
	time.Sleep(3 * time.Second)
	msg("A flash card has a front and a back, where the front is the thing you want to practice and the back is the answer.")
	time.Sleep(3 * time.Second)
	msg("You could have a word in Chinese on the front with the English translation on the back to rehearse your Chinese.")
	time.Sleep(3 * time.Second)
	msg("The front can be anything you can send in Telegram, like a picture of a flag to practice your flag knowledge.")
	time.Sleep(4 * time.Second)
	msg("When you review a card, you first get shown the front of the card, which you should then use to try and remember the back.")
	time.Sleep(4 * time.Second)
	msg("You then reveal the back and indicate how well you remembered it with one of the four given options.")
	time.Sleep(4 * time.Second)
	msg("Depending on how well you did, Memorization Bot will schedule the card to be reviewed again at some later point in the future.")
	time.Sleep(3 * time.Second)
}
