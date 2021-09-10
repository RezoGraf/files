package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func botStart() (tgbotapi.UpdatesChannel, *tgbotapi.BotAPI, error) {
	// используя токен создаем новый инстанс бота
	bot, err := tgbotapi.NewBotAPI(telegramBotToken)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)
	log.Printf("Config file: %s", configFile)
	log.Printf("ChatID: %v", chatID)
	log.Printf("Starting monitoring thread")
	go monitor(bot)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates, err := bot.GetUpdatesChan(updateConfig)
	if err != nil {
		log.Fatal(err)
	}
	return updates, bot, nil
}
