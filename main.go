package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	_ "github.com/Syfaro/telegram-bot-api"
)

var fileName string = "123.txt"

var (
	// глобальная переменная в которой храним токен
	telegramBotToken string = "1983127564:AAHrGY2gFPYj5749aoeTxx5Fp6mCTdHgQ9g"
)

func init() {
	// принимаем на входе флаг -telegrambottoken
	flag.StringVar(&telegramBotToken, "telegrambottoken", "", "Telegram Bot Token")
	flag.Parse()

	// без него не запускаемся
	if telegramBotToken == "" {
		log.Print("-telegrambottoken is required")
		os.Exit(1)
	}
}

func taskKill(processName string) error {
	cmd := exec.Command("cmd.exe", "/C", ("taskkill /im " + processName + " /f"))
	_, err := cmd.CombinedOutput()
	// fmt.Printf("%s\n", stdoutStderr)
	if err != nil {
		fmt.Println("Где мой ", processName, " сучара?")
		return err
	} else {
		fmt.Println("Заебись, закрыли ", processName)
	}
	return nil
}

func copyFile(oldpath, newpath string) error {
	err := os.Rename(oldpath, newpath)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	if err != nil {
		fmt.Println("Ебло, где мой ", fileName, " файл в папке ", oldpath)
		return err
	} else {
		fmt.Println("Заебись, закопировали файл ", fileName)
	}
	return nil
}

func main() {

	// используя токен создаем новый инстанс бота
	bot, err := tgbotapi.NewBotAPI(telegramBotToken)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// u - структура с конфигом для получения апдейтов
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// используя конфиг u создаем канал в который будут прилетать новые сообщения
	updates, err := bot.GetUpdatesChan(u)

	// в канал updates прилетают структуры типа Update
	// вычитываем их и обрабатываем
	for update := range updates {
		// универсальный ответ на любое сообщение
		reply := "Не знаю что сказать"
		if update.Message == nil {
			continue
		}

		// логируем от кого какое сообщение пришло
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		// свитч на обработку комманд
		// комманда - сообщение, начинающееся с "/"
		switch update.Message.Command() {
		case "start":
			reply = "Привет. Я телеграм-бот"
		case "hello":
			reply = "world"
		}

		// создаем ответное сообщение
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		// отправляем
		bot.Send(msg)
	}

	// ----------------работа с процессом и файлами---------------//

	processName := "notepad.exe"
	oldPath := "./1/"
	newPath := "./2/"

	taskKill(processName)
	copyFile((oldPath + fileName), (newPath + fileName))

}
