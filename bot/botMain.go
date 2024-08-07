package bot

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	saver "tgBot/fileSaver/savers"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	env "github.com/joho/godotenv"
)

var numericKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("/WhatToDo"),
		tgbotapi.NewKeyboardButton("/RemoveAll"),
	),
)

type Bot struct {
	botAPI *tgbotapi.BotAPI
	wg     *sync.WaitGroup
}

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
	return &Bot{bot, &wg}, nil
}

func (bot *Bot) StartBot() (*Bot, error) {
	filesBlockTable := make(map[int64]*sync.Mutex, 10)
	limit, err := LookupEnv("GOROUTINE_LIMIT")
	if err != nil {
		return nil, fmt.Errorf("bot start: %w", err)
	}
	goroutineLimit, err := strconv.Atoi(limit)
	if err != nil {
		return nil, fmt.Errorf("bot start: %w", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	bot.wg.Add(goroutineLimit)

	updates := bot.botAPI.GetUpdatesChan(u)
	for i := 0; i < goroutineLimit; i++ {
		go func() {
			defer bot.wg.Done()
			for update := range updates {
				if update.Message == nil {
					continue
				}

				mutex, isExists := filesBlockTable[update.Message.Chat.ID]
				if !isExists {
					filesBlockTable[update.Message.Chat.ID] = &sync.Mutex{}
					mutex = filesBlockTable[update.Message.Chat.ID]
				}

				if update.Message.IsCommand() {
					commandHandler(update.Message, bot.botAPI, mutex)
					continue
				}
				if update.Message.Text != "" {
					textHandler(update.Message, bot.botAPI, mutex)
					continue
				}
			}
		}()
	}
	return bot, nil
}

func (bot *Bot) Stop() {
	bot.botAPI.StopReceivingUpdates()
	bot.wg.Wait()
	fmt.Println("All goroutines have been stopped")
}

func textHandler(message *tgbotapi.Message, bot *tgbotapi.BotAPI, mutex *sync.Mutex) error {
	mutex.Lock()
	defer mutex.Unlock()

	var msg tgbotapi.MessageConfig

	if err := saver.SaveInToToDoListJson(message.Chat.ID, message.From.UserName, message.Text); err != nil {
		return errors.New("textHandler cant write to file")
	} else {
		msg = tgbotapi.NewMessage(message.Chat.ID, "Task has been added")
		_, err := bot.Send(msg)
		if err != nil {
			return errors.New("message wasnt sent")
		}
	}
	return nil
}

func commandHandler(message *tgbotapi.Message, bot *tgbotapi.BotAPI, mutex *sync.Mutex) error {
	mutex.Lock()
	defer mutex.Unlock()

	msg := tgbotapi.NewMessage(message.Chat.ID, "")

	switch message.Command() {
	case "start":
		startHandler(&msg, message.From.FirstName)
	case "help":
		helpHandler(&msg)
	case "GetButton":
		getButtonHandler(&msg)
	case "WhatToDo":
		if err := whatToDoHandler(&msg, message.Chat.ID); err != nil {
			return fmt.Errorf("commandHandler : %w", err)
		}
	case "RemoveAll":
		if err := removeAllHandler(&msg, message.Chat.ID); err != nil {
			return fmt.Errorf("commandHandler : %w", err)
		}
	default:
		msg.Text = "I dont know this command"
	}
	if _, err := bot.Send(msg); err != nil {
		return fmt.Errorf("commandHandler: %w", err)
	}
	return nil
}

func getEnv() error {
	if err := env.Load(); err != nil {
		return errors.New(".env file not found")
	}
	return nil
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

func startHandler(msg *tgbotapi.MessageConfig, name string) {
	buffer := bytes.Buffer{}
	msg.ReplyMarkup = numericKeyboard
	buffer.WriteString("Hi ")
	buffer.WriteString(name)
	buffer.WriteString(", I can save tasks in to ToDoList")
	msg.Text = buffer.String()
}

func helpHandler(msg *tgbotapi.MessageConfig) {
	msg.Text = "I understand: /WhatToDo, /GetButton, /RemoveAll ."
}

func getButtonHandler(msg *tgbotapi.MessageConfig) {
	msg.ReplyMarkup = numericKeyboard
	msg.Text = "Buttons updated"
}

func whatToDoHandler(msg *tgbotapi.MessageConfig, id int64) error {
	bytes, err := saver.GetToDoListJson(id)
	text := "ToDoList:\n" + string(bytes)
	if err != nil {
		log.Print(err)
		return fmt.Errorf("what ToDo handler: %w", err)
	}
	if strings.TrimSpace(text) == "" {
		msg.Text = "ToDo list is empty"
		return nil
	}
	msg.Text = text
	return nil
}

func removeAllHandler(msg *tgbotapi.MessageConfig, id int64) error {
	if err := saver.RemoveToDoListJson(id); err != nil {
		msg.Text = "Failed to clear list"
		return fmt.Errorf("remove all handler: %w", err)
	}
	msg.Text = "List is clear"
	return nil
}
