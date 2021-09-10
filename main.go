package main

import (
	"encoding/json"
	_ "errors"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	// tgbotapi "github.com/Syfaro/telegram-bot-api"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	fileName string = "123.txt"
	// processName  string = "notepad.exe"
	processName2 string
	// oldPath      string = "./1/"
	// newPath      string = "./2/"
	operator [2]string
)

// -----------Переменные и приветствие бота--------

var (
	SiteList         map[string]int
	chatID           int64
	telegramBotToken string
	configFile       string
	HelpMsg          = "Это простой мониторинг доступности сайтов. Он обходит сайты в списке и ждет что он ответит 200, если возвращается не 200 или ошибки подключения, то бот пришлет уведомления в групповой чат\n" +
		"Список доступных комманд:\n" +
		"/site_list - покажет список сайтов в мониторинге и их статусы (про статусы ниже)\n" +
		"/site_add [url] - добавит url в список мониторинга\n" +
		"/site_del [url] - удалит url из списка мониторинга\n" +
		"/help - отобразить это сообщение\n" +
		"\n" +
		"У сайтов может быть несколько статусов:\n" +
		"0 - никогда не проверялся (ждем проверки)\n" +
		"1 - ошибка подключения \n" +
		"200 - ОК-статус" +
		"все остальные http-коды считаются некорректными\n" +
		"/kill_notepad - Убить процесс notepad если есть и получить результат"
)

// ----------Закрытие процесса-------

// ----------Перемещение файла----------

func main() {
	updates, bot, _ := botStart()
	// Работа с полученными обновлениями в чате
	for update := range updates {
		reply := ""

		if update.Message == nil { // ignore any non-Message updates
			continue
		}

		if !update.Message.IsCommand() { // ignore any non-command Messages
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		switch update.Message.Command() { // Извлекаем команду из Message (может принимать в конце значение default: для прочих случаев)
		case "site_list":
			sl, _ := json.Marshal(SiteList)
			reply = string(sl)

		case "site_add":
			SiteList[update.Message.CommandArguments()] = 0
			reply = "Site added to monitoring list"

		case "site_del":
			delete(SiteList, update.Message.CommandArguments())
			reply = "Site deleted from monitoring list"

		case "help":
			reply = HelpMsg

		case "kill":
			processName2 = update.Message.CommandArguments()
			doneMsg, _ := taskKill(processName2)
			reply = string(doneMsg)

		case "copy":
			oldAndNewPath := update.Message.CommandArguments()
			operators := strings.Fields(oldAndNewPath)
			for idx, operand := range operators {
				operator[idx] = operand
			}
			doneCopyMsg, _ := copyFile(operator[0], operator[1])
			reply = string("Копирование из каталога " + operator[0] + " в каталог " + operator[1] + " закончилось " + doneCopyMsg)

		case "tits":
			fileUplodedName := "133.gif"
			photoBytes, err := ioutil.ReadFile(fileUplodedName)
			if err != nil {
				panic(err)
			}
			photoFileBytes := tgbotapi.FileBytes{
				Name:  "fileUplodedName",
				Bytes: photoBytes,
			}

			message, err := bot.Send(tgbotapi.NewPhotoUpload(int64(chatID), photoFileBytes))
			if err != nil {
				panic(err)

			}
			//------Заглушка на будущее--------
			fmt.Println(message.Text)
			// -----конец заглушки-------------
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "") // На будущее создать новый MessageConfig без данных
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.ReplyToMessageID = update.Message.MessageID
			if _, err := bot.Send(msg); err != nil {
				panic(err)
			}
			reply = string("лови")

		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		bot.Send(msg)

		// -----Отправка сообщения с контролем на ошибку
		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
		// -----Конец отправки сообщения с контролем на ошибку

		// ----------------работа с процессом и файлами---------------//

		// taskKill(processName)
		// copyFile((oldPath + fileName), (newPath + fileName))

	}

}
