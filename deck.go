package main

import (
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
)

type Deck struct {
	ID        int    `db:"id"`
	UserID    int    `db:"user_id"`
	Name      string `db:"name"`
	Scheduled bool   `db:"scheduled"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (d *Deck) Delete(tx *sqlx.Tx) error {
	_, err := tx.Exec("DELETE FROM decks WHERE id=$1", d.ID)
	return err
}

func (d *Deck) SetName(tx *sqlx.Tx, name string) error {
	return tx.Get(d, "UPDATE decks SET name=$1 WHERE id=$2 RETURNING *", name, d.ID)
}

func (d *Deck) SetScheduled(tx *sqlx.Tx, scheduled bool) error {
	return tx.Get(d, "UPDATE decks SET scheduled=$1 WHERE id=$2 RETURNING *", scheduled, d.ID)
}

func (d *Deck) GetCardForReview(c *Context) (*Card, error) {
	var card Card
	err := c.tx.Get(&card, `SELECT *
FROM cards
WHERE
 deck_id=$1 AND
 next_repetition <= date_in_time_zone($2)
ORDER BY
 next_repetition ASC,
 repetition_today ASC,
 random_order ASC
LIMIT 1`, d.ID, c.u.TimeZone)
	return &card, err
}

func (d *Deck) CanSetNameTo(tx *sqlx.Tx, name string) (exists bool, err error) {
	err = tx.Get(&exists, `SELECT NOT EXISTS(SELECT 1 FROM decks WHERE user_id=$1 AND id != $2 AND name=$3)`, d.UserID, d.ID, name)
	return
}

func (d *Deck) CreateCard(tx *sqlx.Tx, front []Message, back []Message) (*Card, error) {
	frontJson, err := json.Marshal(front)
	if err != nil {
		return nil, err
	}
	backJson, err := json.Marshal(back)
	if err != nil {
		return nil, err
	}
	var card Card
	err = tx.Get(&card, "INSERT INTO cards (deck_id, front, back) VALUES ($1, $2, $3) RETURNING *", d.ID, frontJson, backJson)
	return &card, err
}
