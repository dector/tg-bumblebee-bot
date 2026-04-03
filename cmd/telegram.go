package main

import (
	"context"
	"fmt"
	"net/url"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type UpdateHandler func(ctx context.Context, api TelegramAPI, update *models.Update)

type TelegramAPI interface {
	SendMessage(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error)
	AnswerInlineQuery(ctx context.Context, params *bot.AnswerInlineQueryParams) (bool, error)
}

type Bot interface {
	TelegramAPI
	Start(ctx context.Context)
}

type BotFactory interface {
	New(token string, updateHandler UpdateHandler) (Bot, error)
}

type botConstructor func(token string, options ...bot.Option) (*bot.Bot, error)

type telegramBotFactory struct {
	newBot botConstructor
}

type telegramBot struct {
	inner *bot.Bot
}

func NewTelegramBotFactory() BotFactory {
	return telegramBotFactory{newBot: bot.New}
}

func newTelegramBotFactoryWithConstructor(newBot botConstructor) BotFactory {
	return telegramBotFactory{newBot: newBot}
}

func (f telegramBotFactory) New(token string, updateHandler UpdateHandler) (Bot, error) {
	if updateHandler == nil {
		return nil, fmt.Errorf("update handler is required")
	}

	constructor := f.newBot
	if constructor == nil {
		constructor = bot.New
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(func(ctx context.Context, tgBot *bot.Bot, update *models.Update) {
			updateHandler(ctx, &telegramBot{inner: tgBot}, update)
		}),
	}

	inner, err := constructor(token, opts...)
	if err != nil {
		return nil, err
	}

	return &telegramBot{inner: inner}, nil
}

func (b *telegramBot) Start(ctx context.Context) {
	b.inner.Start(ctx)
}

func (b *telegramBot) SendMessage(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error) {
	return b.inner.SendMessage(ctx, params)
}

func (b *telegramBot) AnswerInlineQuery(ctx context.Context, params *bot.AnswerInlineQueryParams) (bool, error) {
	return b.inner.AnswerInlineQuery(ctx, params)
}

func handler(ctx context.Context, api TelegramAPI, update *models.Update) {
	ok := processInline(api, ctx, update)
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

				rawURL := update.Message.Text[e.Offset:(e.Offset + e.Length)]
				aURL, err := url.Parse(rawURL)
				if err != nil {
					return
				}

				processURL(api, ctx, aURL, update)
			}()
		}
	}()
}

func processInline(api TelegramAPI, ctx context.Context, update *models.Update) bool {
	query := update.InlineQuery
	if query == nil {
		return false
	}

	fmt.Printf("Inline: %+v\n", update.InlineQuery)

	parsedURL, err := url.Parse(query.Query)
	if err != nil {
		return false
	}
	convertedURL, ok := convertURL(*parsedURL)
	if !ok {
		return false
	}
	if api == nil {
		return false
	}

	_, _ = api.AnswerInlineQuery(ctx, &bot.AnswerInlineQueryParams{
		InlineQueryID: query.ID,
		Results: []models.InlineQueryResult{
			&models.InlineQueryResultArticle{
				ID:    "1",
				Title: "Preview",
				InputMessageContent: &models.InputTextMessageContent{
					MessageText: convertedURL.String(),
				},
			},
		},
	})

	return true
}

func processURL(api TelegramAPI, ctx context.Context, parsedURL *url.URL, update *models.Update) {
	convertedURL, ok := convertURL(*parsedURL)
	if !ok {
		return
	}

	sendReply(api, ctx, convertedURL, update)
}

func sendReply(api TelegramAPI, ctx context.Context, convertedURL *url.URL, update *models.Update) {
	if api == nil || update == nil || update.Message == nil {
		return
	}

	_, _ = api.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   convertedURL.String(),
		ReplyParameters: &models.ReplyParameters{
			MessageID: update.Message.ID,
		},
	})
}
