package bot

import (
	"bytes"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

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
