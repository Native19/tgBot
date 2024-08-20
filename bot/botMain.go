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
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	env "github.com/joho/godotenv"
	converter "tgBot/fileSaver/converters"
)

var numericKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("/WhatToDo"),
		tgbotapi.NewKeyboardButton("/RemoveAll"),
	),
)

var saverImplement Saver

type Worker struct {
	mutex    *sync.Mutex
	stopChan chan struct{}
}

type Bot struct {
	botAPI          *tgbotapi.BotAPI
	wg              *sync.WaitGroup
	filesBlockTable map[int64]*Worker
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
	return &Bot{bot, &wg, make(map[int64]*Worker, 10)}, nil
}

func (bot *Bot) StartBot(saver Saver) (*Bot, error) {
	saverImplement = saver
	limit, err := LookupEnv("GOROUTINE_LIMIT")
	if err != nil {
		return nil, fmt.Errorf("bot start: %w", err)
	}
	goroutineLimit, err := strconv.Atoi(limit)
	if err != nil {
		return nil, fmt.Errorf("bot start: %w", err)
	}

	if err := StartTimersWhenLaunchingBot(bot); err != nil {
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

				worker, isExists := bot.filesBlockTable[update.Message.Chat.ID]
				if !isExists {
					worker = &Worker{
						mutex:    &sync.Mutex{},
						stopChan: make(chan struct{}),
					}
					bot.filesBlockTable[update.Message.Chat.ID] = worker
				}

				if update.Message.IsCommand() {
					commandHandler(update.Message, bot.botAPI, worker)
					continue
				}
				if update.Message.Text != "" {
					textHandler(update.Message, bot, worker)
					continue
				}
			}
		}()
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

func textHandler(message *tgbotapi.Message, bot *Bot, worker *Worker) error {
	worker.mutex.Lock()
	defer worker.mutex.Unlock()

	var msg tgbotapi.MessageConfig

	data := converter.CreateMessageData(message.From.UserName, message.Text)
	if err := saverImplement.SaveInToToDoList(message.Chat.ID, data); err != nil {
		errors.New("textHandler cant write to file")
	} else {
		if data.IsTimeActive {
			bot.wg.Add(1)
			go startTimer(data, message.Chat.ID, bot, worker.stopChan)
		}

		msg = tgbotapi.NewMessage(message.Chat.ID, "Task has been added")
		_, err := bot.botAPI.Send(msg)
		if err != nil {
			return errors.New("message wasnt sent")
		}
	}
	return nil
}

func commandHandler(message *tgbotapi.Message, bot *tgbotapi.BotAPI, worker *Worker) error {
	worker.mutex.Lock()
	defer worker.mutex.Unlock()

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
		close(worker.stopChan)
		fmt.Println(worker.stopChan)
		fmt.Println("---------------------------------------------")
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
	bytesOfList, err := saverImplement.GetToDoList(id)
	if err != nil {
		log.Print(err)
		return fmt.Errorf("what ToDo handler: %w", err)
	}

	text := "ToDoList:\n" + string(bytesOfList)
	if strings.TrimSpace(text) == "" {
		msg.Text = "ToDo list is empty"
		return nil
	}
	msg.Text = text
	return nil
}

func removeAllHandler(msg *tgbotapi.MessageConfig, id int64) error {
	if err := saverImplement.RemoveToDoList(id); err != nil {
		msg.Text = "Failed to clear list"
		return fmt.Errorf("remove all handler: %w", err)
	}
	msg.Text = "List is clear"
	return nil
}

func startTimer(data converter.MessageData, chatID int64, bot *Bot, stopChan chan struct{}) error {
	defer bot.wg.Done()

	timeUntilRun, err := getTimeUntilRun(data)
	if err != nil {
		fmt.Println("startTimer: %w", err)
		return fmt.Errorf("startTimer: %w", err)
	}
	timer := time.NewTimer(timeUntilRun)
	defer timer.Stop()

	select {
	case <-timer.C:
		if err := sendMessage(data, chatID, bot.botAPI); err != nil {
			fmt.Println("startTimer: %w", err)
			return fmt.Errorf("startTimer: %w", err)
		}
	case <-stopChan:
		return nil
	}

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := sendMessage(data, chatID, bot.botAPI); err != nil {
				fmt.Println("startTimer: %w", err)
				return fmt.Errorf("startTimer: %w", err)
			}
		case <-stopChan:
			return nil
		}
	}
}

func sendMessage(data converter.MessageData, chatID int64, botAPI *tgbotapi.BotAPI) error {
	msg := tgbotapi.NewMessage(chatID, data.Task)
	_, err := botAPI.Send(msg)
	if err != nil {
		return errors.New("message wasnt sent")
	}
	return nil
}

func getTimeUntilRun(data converter.MessageData) (time.Duration, error) {
	timeRun, err := data.GetTime()
	if err != nil {
		return 0, fmt.Errorf("startTimer: %w", err)
	}

	dateNow := time.Now()
	targetTimeToday := time.Date(dateNow.Year(), dateNow.Month(), dateNow.Day(), timeRun.Hour(), timeRun.Minute(), 0, 0, dateNow.Location())

	if targetTimeToday.Before(dateNow) {
		targetTimeToday = targetTimeToday.Add(24 * time.Hour)
	}

	timeUntilRun := time.Until(targetTimeToday)
	return timeUntilRun, nil
}

func StartTimersWhenLaunchingBot(bot *Bot) error {
	messages, err := saverImplement.GetTasksWithTimer()
	if err != nil {
		return errors.New("cant get messages with timer")
	}

	for _, message := range messages {
		worker, isExists := bot.filesBlockTable[message.ChatID]
		if !isExists {
			worker = &Worker{
				mutex:    &sync.Mutex{},
				stopChan: make(chan struct{}),
			}
			bot.filesBlockTable[message.ChatID] = worker
		}
		bot.wg.Add(1)
		go startTimer(message.Message, message.ChatID, bot, worker.stopChan)
	}
	return nil
}
