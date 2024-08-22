package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"tgBot/bot"
	saver "tgBot/fileSaver/savers"
	"tgBot/http"
)

func main() {
	fmt.Println("Start")
	saverImplement := &saver.JsonSaver{}
	server, err := http.ServerStart(saverImplement)
	if err != nil {
		log.Fatal(fmt.Errorf("start http server: %w", err))
	}

	fmt.Println("Server started")

	newBot, err := bot.NewBot()
	if err != nil {
		log.Fatal(fmt.Errorf("new bot: %w", err))
	}

	errChan := make(chan error, 10)

	startedBot, err := newBot.StartBot(saverImplement, errChan)
	if err != nil {
		log.Fatal(fmt.Errorf("start bot: %w", err))
	}

	fmt.Println("Bot is now running.")

	interrupt := make(chan os.Signal, 1)
	var interrupted = false

	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	for !interrupted {
		select {
		case err := <-errChan:
			if err != nil {
				log.Println("goroutine err:", err)
			}

		case <-interrupt:
			interrupted = true
			break
		}
	}

	fmt.Println("Bot and Server are shutting down.")

	startedBot.Stop()
	http.ServerStop(server)
}
