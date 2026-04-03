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

var hostMappings = map[string]string{
	"instagram.com": "eeinstagram.com",
	"x.com":         "fixupx.com",
}

var removableSubs = map[string]map[string]struct{}{
	"instagram.com": {
		"www": {},
	},
}

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

func processUrl(b *bot.Bot, ctx context.Context, parsedURL *url.URL, update *models.Update) {
	convertedURL, ok := convertUrl(*parsedURL)
	if !ok {
		return
	}

	sendReply(b, ctx, convertedURL, update)
}

func convertUrl(u url.URL) (*url.URL, bool) {
	normalizedHost := normalizeHost(u.Host)

	mappedHost, ok := hostMappings[normalizedHost]
	if !ok {
		return nil, false
	}

	res := u
	res.Host = mappedHost
	res.RawQuery = ""
	return &res, true
}

func normalizeHost(host string) string {
	parts := strings.Split(host, ".")
	if len(parts) <= 2 {
		return host
	}

	subdomain := parts[0]
	baseHost := strings.Join(parts[1:], ".")

	allowedSubdomains, ok := removableSubs[baseHost]
	if !ok {
		return host
	}

	if _, ok := allowedSubdomains[subdomain]; !ok {
		return host
	}

	return baseHost
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
