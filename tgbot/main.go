package main

import (
	"context"
	"fmt"
	"log"
	"rabbitTest/internal/model"
	"rabbitTest/internal/rabbit"
	"rabbitTest/internal/repository"
	"rabbitTest/internal/server"

	_ "github.com/go-sql-driver/mysql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	producer, err := rabbit.New("comments")
	if err != nil {
		log.Fatalln(err)
	}

	repo, err := repository.New("root:pass@tcp(localhost:3306)/comments")
	if err != nil {
		log.Fatalln(err)
	}

	bot, err := tgbotapi.NewBotAPI("5236601960:AAFhWXb0zBlckYdC2-pyWIJJDKmFb67IC4U")
	if err != nil {
		log.Panic(err)
	}

	srv := server.New(repo, bot)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message != nil {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			fmt.Println("ИМЯ ПОЛЬЗОВАТЕЛЯ = ", update.Message.From.ID, update.Message.From.FirstName)

			comment := model.Comment{
				ID:     update.Message.MessageID,
				Text:   update.Message.Text,
				UserID: update.Message.From.ID,
				ChatID: update.Message.Chat.ID,
			}

			err = repo.InsertComment(context.Background(), comment)
			if err != nil {
				log.Println(err)
			}

			err = producer.Publish(context.Background(), comment)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
