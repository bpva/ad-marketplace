package main

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	rawAPIID := strings.TrimSpace(os.Getenv("TG_API_ID"))
	apiHash := strings.TrimSpace(os.Getenv("TG_API_HASH"))
	phone := strings.TrimSpace(os.Getenv("TG_PHONE"))
	password := strings.TrimSpace(os.Getenv("TG_PASSWORD"))
	sessionPath := ".session.json"

	if rawAPIID == "" || apiHash == "" || phone == "" {
		log.Error("config error", "error", "TG_API_ID, TG_API_HASH, TG_PHONE are required")
		os.Exit(1)
	}

	apiID, err := strconv.Atoi(rawAPIID)
	if err != nil {
		log.Error("config error", "error", fmt.Errorf("parse TG_API_ID: %w", err))
		os.Exit(1)
	}

	client := telegram.NewClient(apiID, apiHash, telegram.Options{
		SessionStorage: &session.FileStorage{Path: sessionPath},
	})

	if err := client.Run(context.Background(), func(ctx context.Context) error {
		authClient := client.Auth()

		userAuth := auth.CodeOnly(phone, auth.CodeAuthenticatorFunc(codePrompt))
		if password != "" {
			userAuth = auth.Constant(phone, password, auth.CodeAuthenticatorFunc(codePrompt))
		}

		if err := authClient.IfNecessary(
			ctx,
			auth.NewFlow(userAuth, auth.SendCodeOptions{}),
		); err != nil {
			return fmt.Errorf("login: %w", err)
		}

		me, err := client.Self(ctx)
		if err != nil {
			return fmt.Errorf("self: %w", err)
		}

		log.Info("session ready", "session_path", sessionPath, "user_id", me.ID)
		return nil
	}); err != nil {
		log.Error("failed", "error", err)
		os.Exit(1)
	}
}

func codePrompt(context.Context, *tg.AuthSentCode) (string, error) {
	fmt.Print("Enter Telegram code: ")
	code, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(code), nil
}
