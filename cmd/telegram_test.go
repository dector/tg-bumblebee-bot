package main

import (
	"context"
	"errors"
	"net/url"
	"testing"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type telegramAPIMock struct {
	sendMessageCalls      int
	answerInlineQueryCall int
	lastSendMessage       *bot.SendMessageParams
	lastInlineAnswer      *bot.AnswerInlineQueryParams
}

func (m *telegramAPIMock) SendMessage(_ context.Context, params *bot.SendMessageParams) (*models.Message, error) {
	m.sendMessageCalls++
	m.lastSendMessage = params
	return nil, nil
}

func (m *telegramAPIMock) AnswerInlineQuery(_ context.Context, params *bot.AnswerInlineQueryParams) (bool, error) {
	m.answerInlineQueryCall++
	m.lastInlineAnswer = params
	return true, nil
}

func TestProcessInlineReturnsFalseWithoutBotCall(t *testing.T) {
	tests := []struct {
		name   string
		update *models.Update
	}{
		{
			name: "no inline query",
			update: &models.Update{
				InlineQuery: nil,
			},
		},
		{
			name: "invalid inline query url",
			update: &models.Update{
				InlineQuery: &models.InlineQuery{Query: "://bad"},
			},
		},
		{
			name: "unsupported inline query host",
			update: &models.Update{
				InlineQuery: &models.InlineQuery{Query: "https://example.com/path"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &telegramAPIMock{}
			got := processInline(mock, context.Background(), tt.update)
			if got {
				t.Fatalf("processInline() = true, want false")
			}
			if mock.answerInlineQueryCall != 0 {
				t.Fatalf("AnswerInlineQuery() calls = %d, want 0", mock.answerInlineQueryCall)
			}
		})
	}
}

func TestProcessInlineCallsAnswerInlineQuery(t *testing.T) {
	mock := &telegramAPIMock{}
	update := &models.Update{
		InlineQuery: &models.InlineQuery{
			ID:    "iq-1",
			Query: "https://x.com/someone/status/1?ref=share",
		},
	}

	ok := processInline(mock, context.Background(), update)
	if !ok {
		t.Fatalf("processInline() = false, want true")
	}
	if mock.answerInlineQueryCall != 1 {
		t.Fatalf("AnswerInlineQuery() calls = %d, want 1", mock.answerInlineQueryCall)
	}
	if mock.lastInlineAnswer == nil {
		t.Fatalf("AnswerInlineQuery() params are nil")
	}
	if mock.lastInlineAnswer.InlineQueryID != "iq-1" {
		t.Fatalf("inline query id = %q, want %q", mock.lastInlineAnswer.InlineQueryID, "iq-1")
	}
	if len(mock.lastInlineAnswer.Results) != 1 {
		t.Fatalf("results len = %d, want 1", len(mock.lastInlineAnswer.Results))
	}

	article, ok := mock.lastInlineAnswer.Results[0].(*models.InlineQueryResultArticle)
	if !ok {
		t.Fatalf("result type = %T, want *models.InlineQueryResultArticle", mock.lastInlineAnswer.Results[0])
	}
	if article.InputMessageContent == nil {
		t.Fatalf("article.InputMessageContent is nil")
	}
	content, ok := article.InputMessageContent.(*models.InputTextMessageContent)
	if !ok {
		t.Fatalf("message content type = %T, want *models.InputTextMessageContent", article.InputMessageContent)
	}
	if content.MessageText != "https://fixupx.com/someone/status/1" {
		t.Fatalf("message text = %q, want %q", content.MessageText, "https://fixupx.com/someone/status/1")
	}
}

func TestProcessURLUnsupportedHostNoop(t *testing.T) {
	u, err := url.Parse("https://example.com/path?x=1")
	if err != nil {
		t.Fatalf("parse input url: %v", err)
	}
	orig := *u
	mock := &telegramAPIMock{}

	processURL(mock, context.Background(), u, nil)

	if *u != orig {
		t.Fatalf("processURL mutated unsupported url: got %+v, want %+v", *u, orig)
	}
	if mock.sendMessageCalls != 0 {
		t.Fatalf("SendMessage() calls = %d, want 0", mock.sendMessageCalls)
	}
}

func TestProcessURLCallsSendMessage(t *testing.T) {
	u, err := url.Parse("https://instagram.com/p/abc?utm_source=test")
	if err != nil {
		t.Fatalf("parse input url: %v", err)
	}
	mock := &telegramAPIMock{}
	update := &models.Update{
		Message: &models.Message{
			ID:   42,
			Chat: models.Chat{ID: 99},
		},
	}

	processURL(mock, context.Background(), u, update)

	if mock.sendMessageCalls != 1 {
		t.Fatalf("SendMessage() calls = %d, want 1", mock.sendMessageCalls)
	}
	if mock.lastSendMessage == nil {
		t.Fatalf("SendMessage() params are nil")
	}
	gotChatID, ok := mock.lastSendMessage.ChatID.(int64)
	if !ok {
		t.Fatalf("chat id type = %T, want int64", mock.lastSendMessage.ChatID)
	}
	if gotChatID != 99 {
		t.Fatalf("chat id = %d, want %d", gotChatID, 99)
	}
	if mock.lastSendMessage.Text != "https://eeinstagram.com/p/abc" {
		t.Fatalf("text = %q, want %q", mock.lastSendMessage.Text, "https://eeinstagram.com/p/abc")
	}
	if mock.lastSendMessage.ReplyParameters == nil || mock.lastSendMessage.ReplyParameters.MessageID != 42 {
		t.Fatalf("reply message id = %v, want %d", mock.lastSendMessage.ReplyParameters, 42)
	}
}

func TestHandlerNoInlineAndNoMessageDoesNothing(t *testing.T) {
	update := &models.Update{}

	handler(context.Background(), nil, update)
}

func TestTelegramBotFactoryUsesConstructor(t *testing.T) {
	called := false
	gotToken := ""
	gotOptionsCount := 0

	factory := newTelegramBotFactoryWithConstructor(func(token string, options ...bot.Option) (*bot.Bot, error) {
		called = true
		gotToken = token
		gotOptionsCount = len(options)
		return &bot.Bot{}, nil
	})

	created, err := factory.New("123:abc", handler)
	if err != nil {
		t.Fatalf("factory.New() error = %v, want nil", err)
	}
	if !called {
		t.Fatalf("constructor was not called")
	}
	if gotToken != "123:abc" {
		t.Fatalf("token = %q, want %q", gotToken, "123:abc")
	}
	if gotOptionsCount == 0 {
		t.Fatalf("options count = %d, want > 0", gotOptionsCount)
	}
	if created == nil {
		t.Fatalf("created bot is nil")
	}
}

func TestTelegramBotFactoryPropagatesConstructorError(t *testing.T) {
	wantErr := errors.New("boom")
	factory := newTelegramBotFactoryWithConstructor(func(token string, options ...bot.Option) (*bot.Bot, error) {
		return nil, wantErr
	})

	created, err := factory.New("123:abc", handler)
	if err == nil {
		t.Fatalf("factory.New() error = nil, want non-nil")
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("factory.New() error = %v, want %v", err, wantErr)
	}
	if created != nil {
		t.Fatalf("created bot = %v, want nil", created)
	}
}

func TestTelegramBotFactoryRejectsNilHandler(t *testing.T) {
	called := false
	factory := newTelegramBotFactoryWithConstructor(func(token string, options ...bot.Option) (*bot.Bot, error) {
		called = true
		return &bot.Bot{}, nil
	})

	created, err := factory.New("123:abc", nil)
	if err == nil {
		t.Fatalf("factory.New() error = nil, want non-nil")
	}
	if called {
		t.Fatalf("constructor should not be called when handler is nil")
	}
	if created != nil {
		t.Fatalf("created bot = %v, want nil", created)
	}
}
