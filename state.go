package main

import (
	"fmt"
	"time"

	"gopkg.in/telegram-bot-api.v4"
)

type Data struct {
	DeckID   int       `json:"d,omitempy"`
	CardID   int       `json:"c,omitempy"`
	Messages []Message `json:"m,omitempy"`
	Front    []Message `json:"f,omitempy"`
	Back     []Message `json:"b,omitempy"`
}

type State uint

const (
	// Main screen that shows a list of decks you can select. You can create decks too.
	// Also has settings and donate buttons or something.
	DeckList = State(iota)

	// Going through all the cards that need to be rehearsed
	Rehearsing

	RehearsingCardReview

	// Create a new deck. Simply takes in a reply for the deck name
	DeckCreate

	// Allows you to type in a query or just list all the cards. Transitions into DeckDetails
	// Transitions back into CardReview or DeckList
	CardSearch

	// Allows searching through the cards. Goes into CardDetails
	// Goes back into CardSearch
	DeckDetails

	// Automatically transitions into another CardCreate
	// A card has two sides, so first take in messages for the front, press 'edit back', send messages for back, and press 'save'
	// You return into CardReview
	CardCreate

	// Same as CardCreate but the backside
	CardCreateBack

	// Confirm the card deletion. After deletion, go to next and review
	CardDelete

	// Shows some stats about a specific card and allows deleting and editing the card.
	// Button for editing front, button for editing back.
	// Goes back into CardReview or DeckDetails
	CardDetails

	// Show the back of the card
	// Buttons:
	// Go back to deck list
	// Add card
	// Edit card
	// Search cards
	CardReview

	// Is editing either the front or the back. Goes back into CardDetails
	CardUpdate

	DeckEdit
	CardEdit
	DeckDelete
	DeckNameEdit
	CardEditFront
	CardEditBack
	SetTimeZone
	Settings

	// Show help and send location
	UserSetup

	SetRehearsalTime

	stateCount
)

func (s State) Show(c *Context) error {
	u := c.u
	tx := c.tx
	reply := c.reply
	createReply := c.createReply
	data := c.data

	switch s {
	case DeckList:
		decks, err := u.GetDecks(tx)
		if err != nil {
			return err
		}
		var replyMessage tgbotapi.MessageConfig

		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(Help),
				tgbotapi.NewKeyboardButton(EditSettings),
				tgbotapi.NewKeyboardButton(AddDeck),
			),
		)

		if len(decks) == 0 {
			replyMessage = createReply("You're now ready to create your first deck, so press '%s' to get started.", AddDeck)
		} else {
			for _, deck := range decks {
				keyboard.Keyboard = append(keyboard.Keyboard, tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(deck.Name),
				))
			}
			replyMessage = createReply("Select the deck you want to work on.")
		}
		keyboard.OneTimeKeyboard = true
		replyMessage.ReplyMarkup = keyboard
		Send(replyMessage)
	case Rehearsing:
		card, err := u.GetScheduledCard(tx)
		if err != nil {
			return err
		}

		if card == nil {
			reply("Done with rehearsal for today!")
			return u.SetAndShowState(c, DeckList, nil)
		} else {
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
			card.SendFront(u.ID, keyboard)
			return nil
		}
	case DeckDetails:
		deck, totalCards, cardsLeft, err := u.GetDeckWithStats(tx, data.DeckID)
		if err != nil {
			return err
		}
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(Back),
				tgbotapi.NewKeyboardButton(EditDeck),
				tgbotapi.NewKeyboardButton(AddCard),
			),
		)
		keyboard.OneTimeKeyboard = true

		if totalCards == 0 {
			msg := createReply("You currently have no cards, so press '%s' to create one.", AddCard)
			msg.ReplyMarkup = keyboard
			Send(msg)
			return nil
		} else if cardsLeft == 0 {
			msg := createReply("No more cards to review today.")
			msg.ReplyMarkup = keyboard
			Send(msg)
			return nil
		} else {
			reply("%d/%d cards left to rehearse in '%s'", cardsLeft, totalCards, deck.Name)

			card, err := deck.GetCardForReview(c)
			if err != nil {
				return err
			}

			keyboard.Keyboard = append(keyboard.Keyboard,
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(EditCard),
					tgbotapi.NewKeyboardButton(ShowReverseOfCard),
				),
			)

			card.SendFront(u.ID, keyboard)
			return nil
		}
	case CardCreate:
		reply("Please send a message to use for the front.")
	case CardCreateBack:
		reply("Please send a message to use for the back.")
	case CardEdit:
		msg := createReply("What would you like to do?")
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(Back),
				tgbotapi.NewKeyboardButton(DeleteCard),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(EditCardFront),
				tgbotapi.NewKeyboardButton(EditCardBack),
			),
		)
		keyboard.OneTimeKeyboard = true
		msg.ReplyMarkup = keyboard
		Send(msg)
	case CardEditFront:
		card, err := GetCard(tx, data.CardID)
		if err != nil {
			return err
		}
		reply("I'm now going to send you the front, please send me back what you want to replace it with.")
		return card.SendFront(u.ID, nil)
	case CardEditBack:
		card, err := GetCard(tx, data.CardID)
		if err != nil {
			return err
		}
		reply("I'm now going to send you the back, please send me back what you want to replace it with.")
		return card.SendBack(u.ID, nil)
	case DeckCreate:
		reply("What's the name of the new deck?")
	case RehearsingCardReview:
		card, err := u.GetScheduledCard(tx)
		if err != nil {
			return err
		}
		card.SendBack(int(c.from), CardReplyKeyboard)
		return nil
	case CardReview:
		deck, err := u.GetDeck(tx, data.DeckID)
		if err != nil {
			return err
		}
		card, err := deck.GetCardForReview(c)
		if err != nil {
			return err
		}
		card.SendBack(int(c.from), CardReplyKeyboard)
		return nil
	case SetTimeZone:
		msg := createReply("Please send me your location, so I can determine your time zone! 🌍")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButtonLocation("Send location"),
			),
		)
		Send(msg)
		return nil
	case DeckDelete:
		_, totalCards, _, err := u.GetDeckWithStats(tx, data.DeckID)
		if err != nil {
			return err
		}
		msg := createReply("Are you sure? You will also delete %d cards.", totalCards)
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(DontDeleteDeck),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(ConfirmDeleteDeck),
			),
		)
		Send(msg)
		return nil
	case DeckEdit:
		deck, err := u.GetDeck(tx, data.DeckID)
		if err != nil {
			return err
		}

		msg := createReply("What do you want to do with '%s'?", deck.Name)
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(Back),
				tgbotapi.NewKeyboardButton(DeleteDeck),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(stringTernary(deck.Scheduled, DisableScheduling, EnableScheduling)),
				tgbotapi.NewKeyboardButton(EditName),
			),
		)
		keyboard.OneTimeKeyboard = true
		msg.ReplyMarkup = keyboard
		Send(msg)
		return nil
	case DeckNameEdit:
		deck, err := u.GetDeck(tx, data.DeckID)
		if err != nil {
			return err
		}
		reply("Please type in the new name for '%s'", deck.Name)
		return nil
	case Settings:
		msg := createReply("What setting do you want to change?")
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(Back),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(fmt.Sprintf(ChangeLocationFormat, u.TimeZone)),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(fmt.Sprintf(ChangeTimeToRehearseFormat, u.RehearsalTime.Format(TimeFormat))),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(stringTernary(u.Scheduled, DisableScheduling, EnableScheduling)),
			),
		)

		keyboard.OneTimeKeyboard = true
		msg.ReplyMarkup = keyboard
		Send(msg)
		return nil
	case UserSetup:
		reply("Hi there!")
		time.Sleep(time.Second)
		HelpUser(int64(u.ID))
		msg := createReply("Now, to get started please send me your location, so I can determine your time zone! 🌍")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButtonLocation("Send location"),
			),
		)
		Send(msg)
	case SetRehearsalTime:
		msg := createReply("Please select your preferred time of day to rehearse. You can also type out the time yourself.")
		keyboard := tgbotapi.NewReplyKeyboard()
		for i := 0; i < 18; i++ {
			keyboard.Keyboard = append(keyboard.Keyboard, tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(time.Date(2000, time.January, 1, 5+i, 0, 0, 0, time.UTC).Format(TimeFormat)),
			))
		}
		keyboard.OneTimeKeyboard = true
		msg.ReplyMarkup = keyboard
		Send(msg)
	}
	return nil
}

func stringTernary(x bool, a string, b string) string {
	if x {
		return a
	} else {
		return b
	}
}
