package repository

import (
	"context"
	"rabbitTest/internal/model"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	conn *sqlx.DB
}

func New(connURI string) (*Repository, error) {
	conn, err := sqlx.Connect("mysql", connURI)
	if err != nil {
		return nil, err
	}

	return &Repository{conn: conn}, nil
}

func (r *Repository) InsertComment(ctx context.Context, comment model.Comment) error {
	query := `
	insert into comments(telegram_id, text, user_id, chat_id) values (:telegram_id, :text, :user_id, :chat_id)
	on duplicate key update updated_at = NOW()
`

	_, err := r.conn.NamedExecContext(
		ctx,
		query,
		comment,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetComment(ctx context.Context, messageID int) (*model.Comment, error) {
	query := `
		select telegram_id, text, user_id, chat_id from comments 
		where telegram_id = ?
	`

	var comment model.Comment
	err := r.conn.GetContext(ctx, &comment, query, messageID)
	if err != nil {
		return nil, err
	}

	return &comment, nil
}
