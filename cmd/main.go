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
	ok := processInline(b, ctx, update)
	if ok {
		return
	}

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

func processInline(b *bot.Bot, ctx context.Context, update *models.Update) bool {
	query := update.InlineQuery
	if query == nil {
		return false
	}

	fmt.Printf("Inline: %+v\n", update.InlineQuery)

	url, err := url.Parse(query.Query)
	if err != nil {
		return false
	}
	url, ok := convertUrl(*url)
	if !ok {
		return false
	}

	b.AnswerInlineQuery(ctx, &bot.AnswerInlineQueryParams{
		InlineQueryID: query.ID,
		Results: []models.InlineQueryResult{
			&models.InlineQueryResultArticle{
				ID:    "1",
				Title: "Preview",
				InputMessageContent: &models.InputTextMessageContent{
					MessageText: url.String(),
				},
			},
		},
	})

	return true
}

// TODO reuse existing url method
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

func convertUrl(url url.URL) (*url.URL, bool) {
	if url.Host == "www.instagram.com" || url.Host == "instagram.com" {
		res := *(&url)
		res.Host = "kkinstagram.com"
		res.RawQuery = ""
		return &res, true
	} else if url.Host == "x.com" {
		res := *(&url)
		res.Host = "fixupx.com"
		res.RawQuery = ""
		return &res, true
	}

	return nil, false
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
