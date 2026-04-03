package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"

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
	_ = err

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

	factory := NewTelegramBotFactory()
	b, err := factory.New(token, handler)
	if err != nil {
		panic(err)
	}

	fmt.Println("[ Bot started ]")
	b.Start(ctx)
}

func convertURL(u url.URL) (*url.URL, bool) {
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
