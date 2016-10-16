package main

import (
	"encoding/json"
	"time"

	"github.com/bouk/memorizationbot/sm"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
)

type Card struct {
	ID     int `db:"id"`
	DeckID int `db:"deck_id"`

	// []Message
	Front types.JSONText `db:"front"`
	// []Message
	Back types.JSONText `db:"back"`

	EasinessFactor   int16 `db:"easiness_factor"`
	PreviousInterval int16 `db:"previous_interval"`
	Repetition       int16 `db:"repetition"`
	RepetitionToday  int16 `db:"repetition_today"`
	RandomOrder      int32 `db:"random_order"`

	NextRepetition time.Time `db:"next_repetition"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

func GetCard(tx *sqlx.Tx, id int) (*Card, error) {
	var card Card
	err := tx.Get(&card, "SELECT * FROM cards WHERE id=$1", id)
	return &card, err
}

func (c *Card) SetFront(tx *sqlx.Tx, messages []Message) error {
	var err error
	if err != nil {
		return err
	}
	c.Front, err = json.Marshal(messages)
	return tx.Get(c, "UPDATE cards SET front=$1 WHERE id=$2 RETURNING *", c.Front, c.ID)
}

func (c *Card) SetBack(tx *sqlx.Tx, messages []Message) error {
	var err error
	c.Back, err = json.Marshal(messages)
	if err != nil {
		return err
	}
	return tx.Get(c, "UPDATE cards SET back=$1 WHERE id=$2 RETURNING *", c.Back, c.ID)
}

func (c *Card) Delete(tx *sqlx.Tx) error {
	_, err := tx.Exec("DELETE FROM cards WHERE id=$1", c.ID)
	return err
}

func (c *Card) GetFront() (messages []Message, err error) {
	err = c.Front.Unmarshal(&messages)
	return
}

func (c *Card) GetBack() (messages []Message, err error) {
	err = c.Back.Unmarshal(&messages)
	return
}

func (c *Card) Respond(context *Context, quality int16) error {
	repetition, easinessFactor, interval := sm.SM2Mod.Calc(quality, c.Repetition, c.EasinessFactor, c.PreviousInterval)
	var repetitionToday int16
	if interval == 0 {
		repetitionToday = c.RepetitionToday + 1
	} else {
		repetitionToday = 0
	}

	return context.tx.Get(c, `UPDATE cards
SET
 easiness_factor=$1,
 previous_interval=$2,
 repetition=$3,
 repetition_today=$4,
 random_order=TRUNC(RANDOM() * 2147483647)::INTEGER,
 next_repetition=date_in_time_zone($5) + ($6)::INTEGER
WHERE
 id=$7
RETURNING *`,
		easinessFactor,
		interval,
		repetition,
		repetitionToday,
		context.u.TimeZone,
		interval,
		c.ID,
	)
}

func (c *Card) SendFront(userID int, keyboard interface{}) error {
	messages, err := c.GetFront()
	if err != nil {
		return err
	}

	for i, message := range messages {
		if i == len(messages)-1 {
			Send(message.ToMessageConfig(int64(userID), keyboard))
		} else {
			Send(message.ToMessageConfig(int64(userID), nil))
		}
	}

	return nil
}

func (c *Card) SendBack(userID int, keyboard interface{}) error {
	messages, err := c.GetBack()
	if err != nil {
		return err
	}

	for i, message := range messages {
		if i == len(messages)-1 {
			Send(message.ToMessageConfig(int64(userID), keyboard))
		} else {
			Send(message.ToMessageConfig(int64(userID), nil))
		}
	}

	return nil
}
