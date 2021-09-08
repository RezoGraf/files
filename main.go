package main

import (
	"crypto/tls"
	"encoding/json"
	_ "errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

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

//------------Загрузка файлов------------------
type User struct {
	ID           int    `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`     // optional
	UserName     string `json:"username"`      // optional
	LanguageCode string `json:"language_code"` // optional
	IsBot        bool   `json:"is_bot"`        // optional
}

type BotAPI struct {
	Token  string `json:"token"`
	Debug  bool   `json:"debug"`
	Buffer int    `json:"buffer"`

	Self   User         `json:"-"`
	Client *http.Client `json:"-"`
	// contains filtered or unexported fields
}

type FileConfig struct {
	FileID string
}

type File struct {
	FileID   string `json:"file_id"`
	FileSize int    `json:"file_size"` // optional
	FilePath string `json:"file_path"` // optional
}

type APIResponse struct {
	Ok          bool                `json:"ok"`
	Result      json.RawMessage     `json:"result"`
	ErrorCode   int                 `json:"error_code"`
	Description string              `json:"description"`
	Parameters  *ResponseParameters `json:"parameters"`
}

type ResponseParameters struct {
	MigrateToChatID int64 `json:"migrate_to_chat_id"` // optional
	RetryAfter      int   `json:"retry_after"`        // optional
}

type Error struct {
	Message string
	ResponseParameters
}

func (e Error) Error() string {
	return e.Message
}

const (
	// APIEndpoint is the endpoint for all API methods,
	// with formatting for Sprintf.
	APIEndpoint = "https://api.telegram.org/bot%s/%s"
	// FileEndpoint is the endpoint for downloading a file from Telegram.
	FileEndpoint = "https://api.telegram.org/file/bot%s/%s"
)

func (bot *BotAPI) decodeAPIResponse(responseBody io.Reader, resp *APIResponse) (_ []byte, err error) {
	if !bot.Debug {
		dec := json.NewDecoder(responseBody)
		err = dec.Decode(resp)
		return
	}

	// if debug, read reponse body
	data, err := ioutil.ReadAll(responseBody)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, resp)
	if err != nil {
		return
	}

	return data, nil
}

func (bot *BotAPI) debugLog(context string, v url.Values, message interface{}) {
	if bot.Debug {
		log.Printf("%s req : %+v\n", context, v)
		log.Printf("%s resp: %+v\n", context, message)
	}
}

func (bot *BotAPI) MakeRequest(endpoint string, params url.Values) (APIResponse, error) {
	method := fmt.Sprintf(APIEndpoint, bot.Token, endpoint)

	resp, err := bot.Client.PostForm(method, params)
	if err != nil {
		return APIResponse{}, err
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	bytes, err := bot.decodeAPIResponse(resp.Body, &apiResp)
	if err != nil {
		return apiResp, err
	}

	if bot.Debug {
		log.Printf("%s resp: %s", endpoint, bytes)
	}

	if !apiResp.Ok {
		parameters := ResponseParameters{}
		if apiResp.Parameters != nil {
			parameters = *apiResp.Parameters
		}
		return apiResp, Error{apiResp.Description, parameters}
	}

	return apiResp, nil
}

type FileBytes struct {
	Name  string
	Bytes []byte
}

type BaseChat struct {
	ChatID              int64 // required
	ChannelUsername     string
	ReplyToMessageID    int
	ReplyMarkup         interface{}
	DisableNotification bool
}

type BaseFile struct {
	BaseChat
	File        interface{}
	FileID      string
	UseExisting bool
	MimeType    string
	FileSize    int
}

func (file BaseFile) getFile() interface{} {
	return file.File
}

func (file BaseFile) params() (map[string]string, error) {
	params := make(map[string]string)

	if file.ChannelUsername != "" {
		params["chat_id"] = file.ChannelUsername
	} else {
		params["chat_id"] = strconv.FormatInt(file.ChatID, 10)
	}

	if file.ReplyToMessageID != 0 {
		params["reply_to_message_id"] = strconv.Itoa(file.ReplyToMessageID)
	}

	if file.ReplyMarkup != nil {
		data, err := json.Marshal(file.ReplyMarkup)
		if err != nil {
			return params, err
		}

		params["reply_markup"] = string(data)
	}

	if file.MimeType != "" {
		params["mime_type"] = file.MimeType
	}

	if file.FileSize > 0 {
		params["file_size"] = strconv.Itoa(file.FileSize)
	}

	params["disable_notification"] = strconv.FormatBool(file.DisableNotification)

	return params, nil
}

func (bot *BotAPI) GetFile(config FileConfig) (File, error) {
	v := url.Values{}
	v.Add("file_id", config.FileID)

	resp, err := bot.MakeRequest("getFile", v)
	if err != nil {
		return File{}, err
	}

	var file File
	json.Unmarshal(resp.Result, &file)

	bot.debugLog("GetFile", v, file)

	return file, nil
}

//-------Инициализация бота-------------
func init() {
	SiteList = make(map[string]int)
	flag.StringVar(&configFile, "config", "config.json", "config file")
	flag.StringVar(&telegramBotToken, "telegrambottoken", "", "Telegram Bot Token")
	flag.Int64Var(&chatID, "chatid", 0, "chatId to send messages")

	flag.Parse()

	if telegramBotToken == "" {
		log.Print("-telegrambottoken is required")
		os.Exit(1)
	}

	if chatID == 0 {
		log.Print("-chatid is required")
		os.Exit(1)
	}

	load_list()
}

// ------------Отправка уведомлений бота--------------
func send_notifications(bot *tgbotapi.BotAPI) {
	for site, status := range SiteList {
		if status != 200 {
			alarm := fmt.Sprintf("CRIT - %s ; status: %v", site, status)
			bot.Send(tgbotapi.NewMessage(chatID, alarm))
		}
	}
}

// ----------Сохранение списка сайтов для бота----------
func save_list() {
	data, err := json.Marshal(SiteList)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(configFile, data, 0644)
	if err != nil {
		panic(err)
	}
}

// ---------Загрузка списка сайтов для бота
func load_list() {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Printf("No such file - starting without config")
		return
	}

	if err = json.Unmarshal(data, &SiteList); err != nil {
		log.Printf("Cant read file - starting without config")
		return
	}
	log.Printf(string(data))
}

// -----------Мониторинг сайтов ботом--------
func monitor(bot *tgbotapi.BotAPI) {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	var httpclient = &http.Client{
		Timeout:   time.Second * 10,
		Transport: tr,
	}

	for {
		save_list()
		for site, _ := range SiteList {
			response, err := httpclient.Get(site)
			if err != nil {
				SiteList[site] = 1
				log.Printf("Status of %s: %s", site, "1 - Connection refused")
			} else {
				log.Printf("Status of %s: %s", site, response.Status)
				SiteList[site] = response.StatusCode
			}
		}
		send_notifications(bot)
		time.Sleep(time.Minute * 5)
	}
}

// ----------Закрытие процесса-------
func taskKill(process string) (string, error) {
	cmd := exec.Command("cmd.exe", "/C", ("taskkill /im " + process + " /f"))
	_, err := cmd.CombinedOutput()
	// fmt.Printf("%s\n", stdoutStderr)
	if err != nil {
		// fmt.Println("Где мой ", process, " сучара?")
		return "процесс " + process + " не найден", err
	} else {
		fmt.Println("Заебись, закрыли ", process)
	}
	return "процесс " + process + " успешно убит", nil
}

// ----------Перемещение файла----------
func copyFile(oldpath, newpath string) (string, error) {
	oldpath = oldpath + fileName
	newpath = newpath + fileName
	err := os.Rename(oldpath, newpath)
	// if err != nil {
	// 	log.Fatal(err)
	// }/
	if err != nil {
		// fmt.Println("Ебло, где мой ", fileName, " файл в папке ", oldpath)
		return "неудачно!", err
	} else {
		fmt.Println("Заебись, закопировали файл ", fileName)
	}
	return "успешно!", nil
}

func main() {

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

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	// s := tgbotapi.BaseFile(tgbotapi.NewUpdate)
	// bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprint("Я живой; вот сайты которые буду мониторить: ", SiteList)))

	updates, err := bot.GetUpdatesChan(u)

	bot.GetFile(FileConfig.FileID)
	for update := range updates {
		reply := ""

		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		switch update.Message.Command() {
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

			// case "download":
			// 	s := tgbotapi.GetFile(update.Message.Document)
			// 	reply = string(s)

		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		bot.Send(msg)

		// ----------------работа с процессом и файлами---------------//

		// taskKill(processName)
		// copyFile((oldPath + fileName), (newPath + fileName))

	}

}
