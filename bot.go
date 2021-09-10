package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func botStart() (tgbotapi.UpdatesChannel, *tgbotapi.BotAPI, error) {
	// используя токен создаем новый инстанс бота
	bot, err := tgbotapi.NewBotAPI(telegramBotToken)

	log.Printf("Authorized on account %s", bot.Self.UserName)
	log.Printf("Config file: %s", configFile)
	log.Printf("ChatID: %v", chatID)
	log.Printf("Starting monitoring thread")
	// go monitor(bot)
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)
	if err != nil {
		log.Fatal(err)
	}
	// updates2, err := bot.GetFile(tgbotapi.FileConfig{})
	return updates, bot, nil
}
