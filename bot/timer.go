package bot

import (
	"errors"
	"fmt"
	"sync"
	converter "tgBot/fileSaver/converters"
	"time"
)

func startTimer(data converter.MessageData, chatID int64, bot *Bot, stopChan chan struct{}) error {
	defer bot.wg.Done()

	if err := firstTick(data, chatID, bot, stopChan); err != nil {
		return fmt.Errorf("startTimer: %w", err)
	}

	if err := startTicker(data, chatID, bot, stopChan); err != nil {
		return fmt.Errorf("startTimer: %w", err)
	}
	return nil
}

func getTimeUntilRun(data converter.MessageData) (time.Duration, error) {
	timeRun, err := data.GetTime()
	if err != nil {
		return 0, fmt.Errorf("startTimer: %w", err)
	}

	dateNow := time.Now()
	targetTimeToday := time.Date(dateNow.Year(),
		dateNow.Month(),
		dateNow.Day(),
		timeRun.Hour(),
		timeRun.Minute(),
		0, 0,
		dateNow.Location())

	if targetTimeToday.Before(dateNow) {
		targetTimeToday = targetTimeToday.Add(24 * time.Hour)
	}

	timeUntilRun := time.Until(targetTimeToday)
	return timeUntilRun, nil
}

func StartTimersWhenLaunchingBot(bot *Bot, errChan chan<- error) error {
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
		go func() {
			errChan <- startTimer(message.Message, message.ChatID, bot, worker.stopChan)
		}()
	}
	return nil
}

func firstTick(data converter.MessageData, chatID int64, bot *Bot, stopChan chan struct{}) error {
	timeUntilRun, err := getTimeUntilRun(data)
	if err != nil {
		return fmt.Errorf("firstTick: %w", err)
	}
	timer := time.NewTimer(timeUntilRun)
	defer timer.Stop()

	select {
	case <-timer.C:
		if err := sendMessage(data.Task, chatID, bot.botAPI); err != nil {
			return fmt.Errorf("firstTick: %w", err)
		}
	case <-stopChan:
		return nil
	}
	return nil
}

func startTicker(data converter.MessageData, chatID int64, bot *Bot, stopChan chan struct{}) error {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := sendMessage(data.Task, chatID, bot.botAPI); err != nil {
				return fmt.Errorf("startTimer: %w", err)
			}
		case <-stopChan:
			return nil
		}
	}
}
