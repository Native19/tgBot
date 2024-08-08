package savers

import (
	"errors"
	"io"
	"os"
	"strconv"

	converter "tgBot/fileSaver/converters"
)

type TxtSaver struct{}

func (saver *TxtSaver) GetToDoList(chatID int64) ([]byte, error) {
	file, err := openFile(chatID, os.O_RDONLY)

	if err != nil {
		return []byte{}, errors.New("cant open file")
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return []byte{}, errors.New("cant read file")
	}
	if len(data) == 0 {
		// return []byte("ToDoList is empty"), nil
		return []byte{}, nil
	}

	var outputData []byte //= []byte("ToDo list:\n")
	outputData = append(outputData, data...)

	return outputData, nil
}

func (saver *TxtSaver) RemoveToDoList(chatID int64) error {
	file, err := openFile(chatID, os.O_WRONLY|os.O_TRUNC)

	if err != nil {
		errors.New("failed to clean the file")
	}
	defer file.Close()

	return nil
}

func (saver *TxtSaver) SaveInToToDoList(chatID int64, data converter.MessageData) error {
	file, err := openFile(chatID, os.O_APPEND|os.O_CREATE)

	if err != nil {
		return errors.New("textHandler cant open file")
	}
	defer file.Close()

	if _, err := file.WriteString(data.Task + "\n"); err != nil {
		return errors.New("textHandler cant write to file")
	} else {
		return nil
	}
}

func openFile(chatID int64, osOpenFlag int) (*os.File, error) {
	fileName := strconv.FormatInt(chatID, 10) + ".txt"
	file, err := os.OpenFile(fileName, osOpenFlag, 0666)
	if err != nil {
		return nil, errors.New("cant open file")
	}

	return file, nil
}
