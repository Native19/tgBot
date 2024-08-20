package savers

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	converter "tgBot/fileSaver/converters"
)

type JsonSaver struct{}

var dir = "./data"

func (saver *JsonSaver) GetToDoList(chatID int64) ([]byte, error) {
	file, err := openFileJson(chatID, os.O_RDONLY)
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

	var messages []converter.MessageData
	if err := json.Unmarshal(data, &messages); err != nil {
		return []byte{}, errors.New("failed to unmarshal JSON")
	}

	var outputData []byte //= []byte("ToDo list:\n")
	for _, message := range messages {
		outputData = append(outputData, []byte(message.GetTask()+"\n")...)
	}

	return outputData, nil
}

func (saver *JsonSaver) RemoveToDoList(chatID int64) error {
	file, err := openFileJson(chatID, os.O_WRONLY|os.O_TRUNC)

	if err != nil {
		return errors.New("failed to clean the file")
	}
	defer file.Close()

	return nil
}

func (saver *JsonSaver) SaveInToToDoList(
	chatID int64,
	data converter.MessageData,
) error {
	file, err := openFileJson(chatID, os.O_RDWR|os.O_CREATE)

	if err != nil {
		return errors.New("SaveInToToDoListJson cant open file")
	}
	defer file.Close()

	var messages []converter.MessageData
	byteValue, err := io.ReadAll(file)
	if err != nil {
		return errors.New("SaveInToToDoListJson failed to read file")
	}

	if len(byteValue) > 0 {
		if err := json.Unmarshal(byteValue, &messages); err != nil {
			return errors.New("failed to unmarshal JSON")
		}
	}

	messages = append(messages, data)

	jsonData, err := json.MarshalIndent(messages, "", "    ")
	if err != nil {
		return errors.New("cant converte into json")
	}

	if err := file.Truncate(0); err != nil {
		return errors.New("failed to truncate file")
	}

	// Устанавливаем указатель файла в начало
	if _, err := file.Seek(0, 0); err != nil {
		return errors.New("failed to seek file")
	}

	if _, err := file.Write(jsonData); err != nil {
		return errors.New("cant write in to file")
	}

	return nil
}

func (saver *JsonSaver) GetTasksWithTimer() ([]converter.Message, error) {
	pattern := dir + "/*.json"
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, errors.New("cant get files")
	}

	var messagesWithTimer []converter.Message
	for _, filePath := range files {
		file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
		if err != nil {
			return nil, errors.New("cant open file")
		}

		data, err := io.ReadAll(file)
		if err != nil {
			file.Close()
			continue
		}
		if len(data) == 0 {
			file.Close()
			continue
		}

		var tasks []converter.MessageData

		err = json.Unmarshal(data, &tasks)
		if err != nil {
			file.Close()
			continue
		}

		fileNameWithExt := filepath.Base(filePath)
		fileName := strings.TrimSuffix(fileNameWithExt, ".json")

		chatID, err := strconv.ParseInt(fileName, 10, 64)
		if err != nil {
			file.Close()
			continue
		}

		for _, task := range tasks {
			if task.IsTimeActive {
				messagesWithTimer = append(messagesWithTimer, converter.Message{chatID, task})
			}
		}
		file.Close()
	}
	return messagesWithTimer, nil
}

func openFileJson(chatID int64, osOpenFlag int) (*os.File, error) {
	fileName := strconv.FormatInt(chatID, 10) + ".json"
	filePath := filepath.Join(dir, fileName)

	file, err := os.OpenFile(filePath, osOpenFlag, 0666)
	if err != nil {
		return nil, errors.New("cant open file")
	}

	return file, nil
}
