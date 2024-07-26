package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"tgBot/bot"
)

func main() {
	newBot, err := bot.NewBot()
	if err != nil {
		log.Fatal(fmt.Errorf("new bot: %w", err))
	}

	startedBot, err := newBot.StartBot()
	if err != nil {
		log.Fatal(fmt.Errorf("start bot: %w", err))
	}

	fmt.Println("Bot is now running.")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	<-interrupt

	fmt.Println("Bot is shutting down.")

	startedBot.Stop()
}
