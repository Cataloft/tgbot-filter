package server

import (
	"net/http"
	"rabbitTest/internal/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/labstack/echo/v4"
)

type Server struct {
	bot *tgbotapi.BotAPI

	repo *repository.Repository
	e    *echo.Echo
}

func New(repo *repository.Repository, bot *tgbotapi.BotAPI) *Server {
	srv := &Server{
		bot:  bot,
		repo: repo,
		e:    echo.New(),
	}

	srv.e.POST("/delete", srv.DeleteMessage)

	return srv
}

func (s *Server) ListenAndServe() error {
	return s.e.Start(":8081")
}

// POST /delete {"id": 123}
func (s *Server) DeleteMessage(c echo.Context) error {

	var body struct {
		ID int `json:"id"`
	}

	if err := c.Bind(&body); err != nil {
		return c.String(http.StatusInternalServerError, "invalid request body")
	}

	comment, err := s.repo.GetComment(c.Request().Context(), body.ID)
	if err != nil {
		return c.String(http.StatusNotFound, "comment not found")
	}

	deleteMessageConfig := tgbotapi.DeleteMessageConfig{
		ChatID:    comment.ChatID,
		MessageID: comment.ID,
	}

	msg := tgbotapi.NewMessage(comment.ChatID, "Ваш комментарий был удален. \nПричина: Комментарий имеет негативную эмоциональную окраску")
	print(comment.ChatID, comment.ID)
	msg.ReplyToMessageID = comment.ID
	s.bot.Send(msg)

	_, err = s.bot.Request(deleteMessageConfig)
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to delete message")
	}

	if err != nil {
		return err
	}
	return c.String(http.StatusOK, "ok")
}
