package bot

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"sync"
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
	goroutineLimit  int
}

func (bot *Bot) starHandle(u tgbotapi.UpdateConfig, errChan chan<- error) error {
	updates := bot.botAPI.GetUpdatesChan(u)
	for i := 0; i < bot.goroutineLimit; i++ {
		go func() {
			defer bot.wg.Done()
			for update := range updates {
				if update.Message == nil {
					continue
				}

				worker := getWorker(update.Message.Chat.ID, bot.filesBlockTable)
				if update.Message.IsCommand() {
					errChan <- commandHandler(update.Message, bot.botAPI, worker)
					continue
				}
				if update.Message.Text != "" {
					errChan <- textHandler(update.Message, bot, worker, errChan)
					continue
				}
			}
		}()
	}
	return nil
}

func textHandler(message *tgbotapi.Message, bot *Bot, worker *Worker, errChan chan<- error) error {
	worker.mutex.Lock()
	defer worker.mutex.Unlock()

	var msg tgbotapi.MessageConfig

	data := converter.CreateMessageData(message.From.UserName, message.Text)
	if err := saverImplement.SaveInToToDoList(message.Chat.ID, data); err != nil {
		return errors.New("textHandler cant write to file")
	} else {
		if data.IsTimeActive {
			bot.wg.Add(1)
			go func() {
				errChan <- startTimer(data, message.Chat.ID, bot, worker.stopChan)
			}()
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
	default:
		msg.Text = "I dont know this command"
	}
	if _, err := bot.Send(msg); err != nil {
		return fmt.Errorf("commandHandler: %w", err)
	}
	return nil
}

func sendMessage(data string, chatID int64, botAPI *tgbotapi.BotAPI) error {
	msg := tgbotapi.NewMessage(chatID, data)
	_, err := botAPI.Send(msg)
	if err != nil {
		return errors.New("message wasnt sent")
	}
	return nil
}

func getWorker(ID int64, filesBlockTable map[int64]*Worker) *Worker {
	worker, isExists := filesBlockTable[ID]
	if !isExists {
		worker = &Worker{
			mutex:    &sync.Mutex{},
			stopChan: make(chan struct{}),
		}
		filesBlockTable[ID] = worker
	}
	return worker
}
