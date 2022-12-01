package model

type Comment struct {
	ID     int    `json:"id" db:"telegram_id"`
	Text   string `json:"text" db:"text"`
	UserID int64  `db:"user_id" json:"-"`
	ChatID int64  `db:"chat_id" json:"-"`
}
