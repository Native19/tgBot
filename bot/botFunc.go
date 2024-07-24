package bot

import (
	"errors"
	"io"
	"os"
	"strconv"
)

func getToDoList(chatID int64) (string, error) {
	baseErr := errors.New("ToDoList пуст")
	file, err := openFile(chatID, os.O_RDONLY)

	if err != nil {
		return "", baseErr
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return "", baseErr
	}
	var outputData []byte = []byte("Список дел:\n")
	outputData = append(outputData, data...)

	return string(outputData), nil
}

func removeToDoList(chatID int64) error {
	baseErr := errors.New("не удалось отчистить файл")
	file, err := openFile(chatID, os.O_WRONLY|os.O_TRUNC)

	if err != nil {
		return baseErr
	}

	defer file.Close()
	return nil
}

func openFile(chatID int64, osOpenFlag int) (*os.File, error) {
	fileName := strconv.FormatInt(chatID, 10) + ".txt"
	file, err := os.OpenFile(fileName, osOpenFlag, 0666)
	if err != nil {
		return nil, errors.New("не удалось открыть файл")
	}
	return file, nil
}
