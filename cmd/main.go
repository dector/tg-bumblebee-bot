package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/joho/godotenv"

	_ "golang.org/x/crypto/x509roots/fallback"
)

func main() {
	err := godotenv.Load()
	// if err != nil {
	// 	panic(err)
	// }

	var token string
	token, ok := os.LookupEnv("TG_BOT_TOKEN")
	if !ok {
		panic("TG_BOT_TOKEN is not provided")
	}
	token = strings.TrimSpace(token)
	if token == "" {
		panic("TG_BOT_TOKEN is not provided")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}
	b, err := bot.New(token, opts...)
	if err != nil {
		panic(err)
	}

	fmt.Println("[ Bot started ]")
	b.Start(ctx)
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	entities := update.Message.Entities

	go func() {
		defer func() {
			recover()
		}()

		for _, e := range entities {
			defer func() {
				recover()
			}()

			go func() {
				if e.Type != models.MessageEntityTypeURL {
					return
				}

				url := update.Message.Text[e.Offset:(e.Offset + e.Length)]

				if strings.HasPrefix(url, urlPrefixInstagram) {
					newUrl := strings.Replace(url, urlPrefixInstagram, urlPrefixInstagramReplace, 1)

					b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID: update.Message.Chat.ID,
						Text:   newUrl,
						ReplyParameters: &models.ReplyParameters{
							MessageID: update.Message.ID,
						},
					})
				}
			}()
		}
	}()
}

const urlPrefixInstagram = "https://www.instagram.com"
const urlPrefixInstagramReplace = "https://www.kkinstagram.com"
