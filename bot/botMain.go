package bot

import (
	"bytes"
	"log"
	"os"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	env "github.com/joho/godotenv"
)

var numericKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("/WhatToDo"),
		tgbotapi.NewKeyboardButton("/RemoveAll"),
	),
)

func init() {
	if err := env.Load(); err != nil {
		log.Print("Файл .env не найден")
	}
}

func MainBot() {
	token, isHaveValue := os.LookupEnv("TELEGRAM_APITOKEN")
	if !isHaveValue {
		log.Panic(".env ошибка")
		return
	}
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	//bot.Debug = true
	var filesBlockTable map[int64]*sync.Mutex = make(map[int64]*sync.Mutex, 10)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}

		_, isExists := filesBlockTable[update.Message.Chat.ID]
		if !isExists {
			filesBlockTable[update.Message.Chat.ID] = &sync.Mutex{}
		}

		if update.Message.IsCommand() {
			go commandHandler(update, bot, filesBlockTable[update.Message.Chat.ID])
			continue
		}
		if update.Message.Text != "" {
			go textHandler(update, bot, filesBlockTable[update.Message.Chat.ID])
			continue
		}
	}
}

func textHandler(update tgbotapi.Update, bot *tgbotapi.BotAPI, mutex *sync.Mutex) {
	mutex.Lock()
	defer mutex.Unlock()

	var msg tgbotapi.MessageConfig
	file, _ := openFile(update.Message.Chat.ID, os.O_APPEND|os.O_CREATE)

	if file == nil {
		log.Panic("Ошибка при открытии файла")
	}
	defer file.Close()

	if _, err := file.WriteString(update.Message.Text + "\n"); err != nil {
		log.Panic("Невозможно перезаписать файл", err)
	} else {
		msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Задача добавлена")
		bot.Send(msg)
	}
}

func commandHandler(update tgbotapi.Update, bot *tgbotapi.BotAPI, mutex *sync.Mutex) {
	mutex.Lock()
	defer mutex.Unlock()
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

	switch update.Message.Command() {
	case "start":
		buffer := bytes.Buffer{}
		msg.ReplyMarkup = numericKeyboard
		buffer.WriteString("Привет ")
		buffer.WriteString(update.Message.From.FirstName)
		buffer.WriteString(", Я умею сохранять задачи в ToDoList")
		msg.Text = buffer.String()
	case "help":
		msg.Text = "Я понимаю команды: /WhatToDo, /GetButton, /RemoveAll ."
	case "GetButton":
		msg.ReplyMarkup = numericKeyboard
		msg.Text = "Кнопки обновлены"
	case "WhatToDo":
		text, err := getToDoList(update.Message.Chat.ID)
		if err != nil {
			log.Print(err)
			msg.Text = "ToDoList пуст"
			break
		}
		msg.Text = text
	case "RemoveAll":
		if removeToDoList(update.Message.Chat.ID) != nil {
			msg.Text = "Не удалось отчистить список"
			break
		}
		msg.Text = "Список отчищен"
	default:
		msg.Text = "Я не знаю данной команды"
	}

	if strings.TrimSpace(msg.Text) == "" {
		msg.Text = "Список дел пуст!"
	}
	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}
}
