package main

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"gopkg.in/telegram-bot-api.v4"
)

type Context struct {
	u    *User
	tx   *sqlx.Tx
	data *Data
	from int64
}

func (c *Context) createReply(format string, data ...interface{}) tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(c.from, fmt.Sprintf(format, data...))
}
func (c *Context) reply(format string, data ...interface{}) (tgbotapi.Message, error) {
	return Send(c.createReply(format, data...))
}
