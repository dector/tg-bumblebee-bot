package main

import (
	"context"
	"net/url"
	"testing"

	"github.com/go-telegram/bot/models"
)

func TestConvertUrl(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantHost     string
		wantPath     string
		wantRawQuery string
		wantOK       bool
	}{
		{
			name:         "instagram host",
			input:        "https://instagram.com/p/abc?utm_source=test",
			wantHost:     "eeinstagram.com",
			wantPath:     "/p/abc",
			wantRawQuery: "",
			wantOK:       true,
		},
		{
			name:         "instagram www host",
			input:        "https://www.instagram.com/p/abc?foo=bar",
			wantHost:     "eeinstagram.com",
			wantPath:     "/p/abc",
			wantRawQuery: "",
			wantOK:       true,
		},
		{
			name:         "x host",
			input:        "https://x.com/someone/status/1?ref=share",
			wantHost:     "fixupx.com",
			wantPath:     "/someone/status/1",
			wantRawQuery: "",
			wantOK:       true,
		},
		{
			name:   "unsupported host",
			input:  "https://example.com/a?b=c",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.input)
			if err != nil {
				t.Fatalf("parse input url: %v", err)
			}

			orig := *u
			got, ok := convertUrl(*u)
			if ok != tt.wantOK {
				t.Fatalf("convertUrl() ok = %v, want %v", ok, tt.wantOK)
			}

			if !tt.wantOK {
				if got != nil {
					t.Fatalf("convertUrl() url = %v, want nil", got)
				}
				if *u != orig {
					t.Fatalf("input url was mutated: got %+v, want %+v", *u, orig)
				}
				return
			}

			if got == nil {
				t.Fatalf("convertUrl() returned nil url with ok=true")
			}

			if got.Host != tt.wantHost {
				t.Fatalf("host = %q, want %q", got.Host, tt.wantHost)
			}
			if got.Path != tt.wantPath {
				t.Fatalf("path = %q, want %q", got.Path, tt.wantPath)
			}
			if got.RawQuery != tt.wantRawQuery {
				t.Fatalf("rawQuery = %q, want %q", got.RawQuery, tt.wantRawQuery)
			}
			if *u != orig {
				t.Fatalf("input url was mutated: got %+v, want %+v", *u, orig)
			}
		})
	}
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
			got := processInline(nil, context.Background(), tt.update)
			if got {
				t.Fatalf("processInline() = true, want false")
			}
		})
	}
}

func TestProcessUrlUnsupportedHostNoop(t *testing.T) {
	u, err := url.Parse("https://example.com/path?x=1")
	if err != nil {
		t.Fatalf("parse input url: %v", err)
	}
	orig := *u

	processUrl(nil, context.Background(), u, nil)

	if *u != orig {
		t.Fatalf("processUrl mutated unsupported url: got %+v, want %+v", *u, orig)
	}
}

func TestHandlerNoInlineAndNoMessageDoesNothing(t *testing.T) {
	update := &models.Update{}

	handler(context.Background(), nil, update)
}
