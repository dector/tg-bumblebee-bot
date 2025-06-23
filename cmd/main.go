package main

import (
	"context"
	"fmt"
	"net/url"
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

				rawUrl := update.Message.Text[e.Offset:(e.Offset + e.Length)]
				aUrl, err := url.Parse(rawUrl)
				if err != nil {
					return
				}

				processUrl(b, ctx, aUrl, update)
			}()
		}
	}()
}

func processUrl(b *bot.Bot, ctx context.Context, url *url.URL, update *models.Update) {
	if url.Host == "www.instagram.com" || url.Host == "instagram.com" {
		url.Host = "kkinstagram.com"
		url.RawQuery = ""

		sendReply(b, ctx, url, update)
	} else if url.Host == "x.com" {
		url.Host = "fixupx.com"
		url.RawQuery = ""

		sendReply(b, ctx, url, update)
	}
}

func sendReply(b *bot.Bot, ctx context.Context, url *url.URL, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   url.String(),
		ReplyParameters: &models.ReplyParameters{
			MessageID: update.Message.ID,
		},
	})
}
