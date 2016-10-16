package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
)

type User struct {
	ID            int            `db:"id"`
	CreatedAt     time.Time      `db:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at"`
	NextRehearsal time.Time      `db:"rehearsal"`
	RehearsalTime time.Time      `db:"rehearsal_time"`
	State         State          `db:"state"`
	TimeZone      string         `db:"time_zone"`
	Data          types.JSONText `db:"data"`
	Scheduled     bool           `db:"scheduled"`
}

func (u *User) GetDecks(tx *sqlx.Tx) ([]Deck, error) {
	decks := []Deck{}
	err := tx.Select(&decks, "SELECT * FROM decks WHERE user_id=$1 ORDER BY name ASC", u.ID)
	return decks, err
}

func (u *User) HasDeckWithName(tx *sqlx.Tx, name string) (exists bool, err error) {
	err = tx.Get(&exists, `SELECT EXISTS(SELECT 1 FROM decks WHERE user_id=$1 AND name=$2)`, u.ID, name)
	return
}

// return deck, total_cards, cards_left
func (u *User) GetDeck(tx *sqlx.Tx, id int) (*Deck, error) {
	var deck Deck
	err := tx.Get(&deck, `SELECT * FROM decks WHERE user_id=$1 AND id=$2 LIMIT 1`, u.ID, id)
	return &deck, err
}

// return deck, total_cards, cards_left
func (u *User) GetDeckWithStats(tx *sqlx.Tx, id int) (*Deck, int, int, error) {
	var result struct {
		Deck
		TotalCards int `db:"total_cards"`
		CardsLeft  int `db:"cards_left"`
	}
	err := tx.Get(&result, `WITH deck AS (SELECT * FROM cards WHERE deck_id=$3)
 SELECT
 (SELECT COUNT(*) FROM deck) AS total_cards,
 (SELECT COUNT(CASE WHEN next_repetition <= date_in_time_zone($1) THEN TRUE END) FROM deck) AS cards_left,
 *
 FROM decks
 WHERE user_id=$2 AND id=$3
 LIMIT 1`, u.TimeZone, u.ID, id)
	return &result.Deck, result.TotalCards, result.CardsLeft, err
}

func (u *User) GetDeckByOffset(tx *sqlx.Tx, offset int) (*Deck, error) {
	var deck Deck
	err := tx.Get(&deck, "SELECT * FROM decks WHERE user_id=$1 ORDER BY name ASC LIMIT 1 OFFSET $2", u.ID, offset)
	if err == sql.ErrNoRows {
		return nil, nil
	} else {
		return &deck, err
	}
}

func (u *User) GetDeckByName(tx *sqlx.Tx, name string) (*Deck, error) {
	var deck Deck
	err := tx.Get(&deck, "SELECT * FROM decks WHERE user_id=$1 AND name=$2 LIMIT 1", u.ID, name)
	if err == sql.ErrNoRows {
		return nil, nil
	} else {
		return &deck, err
	}
}

func (u *User) SetState(tx *sqlx.Tx, state State, data *Data) error {
	var err error
	if data == nil {
		u.Data = []byte{'{', '}'}
	} else {
		u.Data, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}
	u.State = state
	_, err = tx.Exec("UPDATE users SET state=$1, data=$2 WHERE id=$3", u.State, u.Data, u.ID)
	return err
}

func (u *User) SetAndShowState(c *Context, state State, data *Data) error {
	c.data = data
	if err := u.SetState(c.tx, state, data); err != nil {
		return err
	}
	return state.Show(c)
}

func (u *User) SetTimeZone(tx *sqlx.Tx, timeZoneID string) (err error) {
	err = tx.Get(u, "UPDATE users SET time_zone=$1 WHERE id=$2 RETURNING *", timeZoneID, u.ID)
	return
}

func (u *User) SetScheduled(tx *sqlx.Tx, scheduled bool) (err error) {
	err = tx.Get(u, "UPDATE users SET scheduled=$1 WHERE id=$2 RETURNING *", scheduled, u.ID)
	return
}

func (u *User) SetRehearsalTime(tx *sqlx.Tx, t time.Time) (err error) {
	err = tx.Get(u, "UPDATE users SET rehearsal_time=$1 WHERE id=$2 RETURNING *", t.Format(TimeFormat), u.ID)
	return
}

func (u *User) CreateDeck(tx *sqlx.Tx, name string) (*Deck, error) {
	var deck Deck
	err := tx.Get(&deck, "INSERT INTO decks (user_id, name) VALUES ($1, $2) RETURNING *", u.ID, name)
	return &deck, err
}

func (u *User) GetScheduledCard(tx *sqlx.Tx) (*Card, error) {
	var card Card
	err := tx.Get(&card, "SELECT * FROM scheduled_card_for_user($1)", u.ID)

	if err == sql.ErrNoRows {
		return nil, nil
	} else {
		return &card, err
	}
}

func WithUser(ID int, f func(*User, *sqlx.Tx) error) error {
	tx, err := DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	log.Printf("getting state %d", ID)

	var user User
	err = tx.Get(&user, "SELECT * FROM users WHERE id=$1 FOR UPDATE", ID)
	if err == sql.ErrNoRows {
		_, err = tx.Exec("INSERT INTO users (id, state) VALUES ($1, $2) ON CONFLICT DO NOTHING", ID, DeckList)
		if err != nil {
			return err
		}

		err = tx.Get(&user, "SELECT * FROM users WHERE id=$1 FOR UPDATE", ID)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	log.Printf("state=%+v", user)

	err = f(&user, tx)
	if err != nil {
		return err
	}

	return tx.Commit()
}
