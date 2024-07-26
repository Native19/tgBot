package bot

import (
	"errors"
	"io"
	"os"
	"strconv"
)

func GetToDoList(chatID int64) (string, error) {
	file, err := openFile(chatID, os.O_RDONLY)
	if err != nil {
		return "", errors.New("cant open file")
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return "", errors.New("cant read file")
	}
	if len(data) == 0 {
		return "ToDoList is empty", nil
	}

	var outputData []byte = []byte("ToDo list:\n")
	outputData = append(outputData, data...)

	return string(outputData), nil
}

func RemoveToDoList(chatID int64) error {
	file, err := openFile(chatID, os.O_WRONLY|os.O_TRUNC)
	if err != nil {
		errors.New("failed to clean the file")
	}
	defer file.Close()

	return nil
}

func openFile(chatID int64, osOpenFlag int) (*os.File, error) {
	fileName := strconv.FormatInt(chatID, 10) + ".txt"
	file, err := os.OpenFile(fileName, osOpenFlag, 0666)
	if err != nil {
		return nil, errors.New("cant open file")
	}

	return file, nil
}
