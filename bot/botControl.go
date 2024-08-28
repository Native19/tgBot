package bot

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	env "github.com/joho/godotenv"
	"os"
	"strconv"
	"sync"
)

func NewBot() (*Bot, error) {
	err := getEnv()
	if err != nil {
		return nil, fmt.Errorf("bot create: %w", err)
	}

	token, err := LookupEnv("TELEGRAM_APITOKEN")
	if err != nil {
		return nil, fmt.Errorf("bot create: %w", err)
	}

	bot, err := createBot(token)
	if err != nil {
		return nil, fmt.Errorf("bot create: %w", err)
	}

	var wg sync.WaitGroup
	return &Bot{bot, &wg, make(map[int64]*Worker, 10), 0}, nil
}

func (bot *Bot) StartBot(saver Saver, errChan chan<- error) (*Bot, error) {
	saverImplement = saver
	limit, err := LookupEnv("GOROUTINE_LIMIT")
	if err != nil {
		return nil, fmt.Errorf("bot start: %w", err)
	}
	bot.goroutineLimit, err = strconv.Atoi(limit)
	if err != nil {
		return nil, fmt.Errorf("bot start: %w", err)
	}

	if err := StartTimersWhenLaunchingBot(bot, errChan); err != nil {
		return nil, fmt.Errorf("bot start: %w", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	bot.wg.Add(bot.goroutineLimit)

	if err := bot.starHandle(u, errChan); err != nil {
		return nil, fmt.Errorf("bot start: %w", err)
	}

	return bot, nil
}

func (bot *Bot) Stop() {
	bot.botAPI.StopReceivingUpdates()
	for _, worker := range bot.filesBlockTable {
		close(worker.stopChan)
	}
	bot.wg.Wait()
	fmt.Println("All goroutines have been stopped")
}

func LookupEnv(key string) (string, error) {
	token, isHaveValue := os.LookupEnv(key)
	if !isHaveValue {
		return "", errors.New("cant get value by key in .env")
	}
	return token, nil
}

func createBot(token string) (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, errors.New("cant create bot")
	}
	return bot, nil
}

func getEnv() error {
	if err := env.Load(); err != nil {
		return errors.New(".env file not found")
	}
	return nil
}
