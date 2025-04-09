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
)

func main() {
	env, err := godotenv.Read(".env")
	if err != nil {
		panic(err)
	}

	token := strings.TrimSpace(env["TG_BOT_TOKEN"])
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
	var entities []models.MessageEntity
	if update.Message != nil {
		entities = update.Message.Entities
	}

	for _, e := range entities {
		// fmt.Printf("%+v\n", e)
		// fmt.Printf("message: %s\n", update.Message.Text)

		url := update.Message.Text[e.Offset:(e.Offset + e.Length)]
		// fmt.Printf("url:%s\n", url)

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
	}
}

const urlPrefixInstagram = "https://www.instagram.com"
const urlPrefixInstagramReplace = "https://www.kkinstagram.com"
