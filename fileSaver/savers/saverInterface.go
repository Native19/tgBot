package savers

import (
	converter "tgBot/fileSaver/converters"
)

type Saver interface {
	GetToDoList(chatID int64) ([]byte, error)
	RemoveToDoList(chatID int64) error
	SaveInToToDoList(chatID int64, data converter.MessageData) error
}
