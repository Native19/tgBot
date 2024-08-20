package converters

import (
	"regexp"
	"strings"
	"time"
)

type MessageData struct {
	User         string `json:"username"`
	Task         string `json:"message"`
	Time         string `json:"time,omitempty"`
	IsTimeActive bool   `json:"isTimerActive"`
}

type Message struct {
	ChatID  int64       `json:"chat_id"`
	Message MessageData `json:"message"`
}

func (s *MessageData) SetTime(t time.Time) {
	s.Time = t.Format("15:04")
	t.IsZero()
}

func (s *MessageData) GetTime() (time.Time, error) {
	return time.Parse("15:04", s.Time)
}

func CreateMessageData(userName, message string) MessageData {
	strWithoutSpaces := strings.ReplaceAll(message, " ", "")
	re := regexp.MustCompile(`\d{2}:\d{2}$`)
	isTimer := re.MatchString(strWithoutSpaces)

	messageData := MessageData{
		User:         userName,
		Task:         message,
		IsTimeActive: isTimer,
	}

	if isTimer {
		messageData.Time = strWithoutSpaces[len(strWithoutSpaces)-5:]
		messageData.Task = message[len(strWithoutSpaces)-5:]
	}

	return messageData
}

func (s *MessageData) GetTask() string {
	return s.Task
}
