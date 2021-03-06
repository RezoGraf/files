package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	// tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	// tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/technoweenie/multipartstreamer"
)

const (
	// ErrAPIForbidden happens when a token is bad
	ErrAPIForbidden = "forbidden"
)

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

type ChatPhoto struct {
	SmallFileID string `json:"small_file_id"`
	BigFileID   string `json:"big_file_id"`
}

type Chat struct {
	ID                  int64      `json:"id"`
	Type                string     `json:"type"`
	Title               string     `json:"title"`                          // optional
	UserName            string     `json:"username"`                       // optional
	FirstName           string     `json:"first_name"`                     // optional
	LastName            string     `json:"last_name"`                      // optional
	AllMembersAreAdmins bool       `json:"all_members_are_administrators"` // optional
	Photo               *ChatPhoto `json:"photo"`
	Description         string     `json:"description,omitempty"` // optional
	InviteLink          string     `json:"invite_link,omitempty"` // optional
}

type MessageEntity struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
	URL    string `json:"url"`  // optional
	User   *User  `json:"user"` // optional
}

type Audio struct {
	FileID    string `json:"file_id"`
	Duration  int    `json:"duration"`
	Performer string `json:"performer"` // optional
	Title     string `json:"title"`     // optional
	MimeType  string `json:"mime_type"` // optional
	FileSize  int    `json:"file_size"` // optional
}

type PhotoSize struct {
	FileID   string `json:"file_id"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	FileSize int    `json:"file_size"` // optional
}

type Document struct {
	FileID    string     `json:"file_id"`
	Thumbnail *PhotoSize `json:"thumb"`     // optional
	FileName  string     `json:"file_name"` // optional
	MimeType  string     `json:"mime_type"` // optional
	FileSize  int        `json:"file_size"` // optional
}

type ChatAnimation struct {
	FileID    string     `json:"file_id"`
	Width     int        `json:"width"`
	Height    int        `json:"height"`
	Duration  int        `json:"duration"`
	Thumbnail *PhotoSize `json:"thumb"`     // optional
	FileName  string     `json:"file_name"` // optional
	MimeType  string     `json:"mime_type"` // optional
	FileSize  int        `json:"file_size"` // optional
}

type Animation struct {
	FileID   string    `json:"file_id"`
	Thumb    PhotoSize `json:"thumb"`
	FileName string    `json:"file_name"`
	MimeType string    `json:"mime_type"`
	FileSize int       `json:"file_size"`
}

type Game struct {
	Title        string          `json:"title"`
	Description  string          `json:"description"`
	Photo        []PhotoSize     `json:"photo"`
	Text         string          `json:"text"`
	TextEntities []MessageEntity `json:"text_entities"`
	Animation    Animation       `json:"animation"`
}

type Sticker struct {
	FileID    string     `json:"file_id"`
	Width     int        `json:"width"`
	Height    int        `json:"height"`
	Thumbnail *PhotoSize `json:"thumb"`     // optional
	Emoji     string     `json:"emoji"`     // optional
	FileSize  int        `json:"file_size"` // optional
	SetName   string     `json:"set_name"`  // optional
}

type Video struct {
	FileID    string     `json:"file_id"`
	Width     int        `json:"width"`
	Height    int        `json:"height"`
	Duration  int        `json:"duration"`
	Thumbnail *PhotoSize `json:"thumb"`     // optional
	MimeType  string     `json:"mime_type"` // optional
	FileSize  int        `json:"file_size"` // optional
}

type VideoNote struct {
	FileID    string     `json:"file_id"`
	Length    int        `json:"length"`
	Duration  int        `json:"duration"`
	Thumbnail *PhotoSize `json:"thumb"`     // optional
	FileSize  int        `json:"file_size"` // optional
}

type Voice struct {
	FileID   string `json:"file_id"`
	Duration int    `json:"duration"`
	MimeType string `json:"mime_type"` // optional
	FileSize int    `json:"file_size"` // optional
}

type Contact struct {
	PhoneNumber string `json:"phone_number"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"` // optional
	UserID      int    `json:"user_id"`   // optional
}

type Location struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

type Venue struct {
	Location     Location `json:"location"`
	Title        string   `json:"title"`
	Address      string   `json:"address"`
	FoursquareID string   `json:"foursquare_id"` // optional
}

type Invoice struct {
	Title          string `json:"title"`
	Description    string `json:"description"`
	StartParameter string `json:"start_parameter"`
	Currency       string `json:"currency"`
	TotalAmount    int    `json:"total_amount"`
}

const (
	// ErrBadFileType happens when you pass an unknown type
	ErrBadFileType = "bad file type"
	ErrBadURL      = "bad or empty url"
)

type ShippingAddress struct {
	CountryCode string `json:"country_code"`
	State       string `json:"state"`
	City        string `json:"city"`
	StreetLine1 string `json:"street_line1"`
	StreetLine2 string `json:"street_line2"`
	PostCode    string `json:"post_code"`
}

type OrderInfo struct {
	Name            string           `json:"name,omitempty"`
	PhoneNumber     string           `json:"phone_number,omitempty"`
	Email           string           `json:"email,omitempty"`
	ShippingAddress *ShippingAddress `json:"shipping_address,omitempty"`
}

type SuccessfulPayment struct {
	Currency                string     `json:"currency"`
	TotalAmount             int        `json:"total_amount"`
	InvoicePayload          string     `json:"invoice_payload"`
	ShippingOptionID        string     `json:"shipping_option_id,omitempty"`
	OrderInfo               *OrderInfo `json:"order_info,omitempty"`
	TelegramPaymentChargeID string     `json:"telegram_payment_charge_id"`
	ProviderPaymentChargeID string     `json:"provider_payment_charge_id"`
}

type PassportFile struct {
	// Unique identifier for this file
	FileID string `json:"file_id"`

	// File size
	FileSize int `json:"file_size"`

	// Unix time when the file was uploaded
	FileDate int64 `json:"file_date"`
}

type EncryptedPassportElement struct {
	// Element type.
	Type string `json:"type"`

	// Base64-encoded encrypted Telegram Passport element data provided by
	// the user, available for "personal_details", "passport",
	// "driver_license", "identity_card", "identity_passport" and "address"
	// types. Can be decrypted and verified using the accompanying
	// EncryptedCredentials.
	Data string `json:"data,omitempty"`

	// User's verified phone number, available only for "phone_number" type
	PhoneNumber string `json:"phone_number,omitempty"`

	// User's verified email address, available only for "email" type
	Email string `json:"email,omitempty"`

	// Array of encrypted files with documents provided by the user,
	// available for "utility_bill", "bank_statement", "rental_agreement",
	// "passport_registration" and "temporary_registration" types. Files can
	// be decrypted and verified using the accompanying EncryptedCredentials.
	Files []PassportFile `json:"files,omitempty"`

	// Encrypted file with the front side of the document, provided by the
	// user. Available for "passport", "driver_license", "identity_card" and
	// "internal_passport". The file can be decrypted and verified using the
	// accompanying EncryptedCredentials.
	FrontSide *PassportFile `json:"front_side,omitempty"`

	// Encrypted file with the reverse side of the document, provided by the
	// user. Available for "driver_license" and "identity_card". The file can
	// be decrypted and verified using the accompanying EncryptedCredentials.
	ReverseSide *PassportFile `json:"reverse_side,omitempty"`

	// Encrypted file with the selfie of the user holding a document,
	// provided by the user; available for "passport", "driver_license",
	// "identity_card" and "internal_passport". The file can be decrypted
	// and verified using the accompanying EncryptedCredentials.
	Selfie *PassportFile `json:"selfie,omitempty"`
}

type EncryptedCredentials struct {
	// Base64-encoded encrypted JSON-serialized data with unique user's
	// payload, data hashes and secrets required for EncryptedPassportElement
	// decryption and authentication
	Data string `json:"data"`

	// Base64-encoded data hash for data authentication
	Hash string `json:"hash"`

	// Base64-encoded secret, encrypted with the bot's public RSA key,
	// required for data decryption
	Secret string `json:"secret"`
}

type PassportData struct {
	// Array with information about documents and other Telegram Passport
	// elements that was shared with the bot
	Data []EncryptedPassportElement `json:"data"`

	// Encrypted credentials required to decrypt the data
	Credentials *EncryptedCredentials `json:"credentials"`
}

type Message struct {
	MessageID             int                `json:"message_id"`
	From                  *User              `json:"from"` // optional
	Date                  int                `json:"date"`
	Chat                  *Chat              `json:"chat"`
	ForwardFrom           *User              `json:"forward_from"`            // optional
	ForwardFromChat       *Chat              `json:"forward_from_chat"`       // optional
	ForwardFromMessageID  int                `json:"forward_from_message_id"` // optional
	ForwardDate           int                `json:"forward_date"`            // optional
	ReplyToMessage        *Message           `json:"reply_to_message"`        // optional
	EditDate              int                `json:"edit_date"`               // optional
	Text                  string             `json:"text"`                    // optional
	Entities              *[]MessageEntity   `json:"entities"`                // optional
	Audio                 *Audio             `json:"audio"`                   // optional
	Document              *Document          `json:"document"`                // optional
	Animation             *ChatAnimation     `json:"animation"`               // optional
	Game                  *Game              `json:"game"`                    // optional
	Photo                 *[]PhotoSize       `json:"photo"`                   // optional
	Sticker               *Sticker           `json:"sticker"`                 // optional
	Video                 *Video             `json:"video"`                   // optional
	VideoNote             *VideoNote         `json:"video_note"`              // optional
	Voice                 *Voice             `json:"voice"`                   // optional
	Caption               string             `json:"caption"`                 // optional
	Contact               *Contact           `json:"contact"`                 // optional
	Location              *Location          `json:"location"`                // optional
	Venue                 *Venue             `json:"venue"`                   // optional
	NewChatMembers        *[]User            `json:"new_chat_members"`        // optional
	LeftChatMember        *User              `json:"left_chat_member"`        // optional
	NewChatTitle          string             `json:"new_chat_title"`          // optional
	NewChatPhoto          *[]PhotoSize       `json:"new_chat_photo"`          // optional
	DeleteChatPhoto       bool               `json:"delete_chat_photo"`       // optional
	GroupChatCreated      bool               `json:"group_chat_created"`      // optional
	SuperGroupChatCreated bool               `json:"supergroup_chat_created"` // optional
	ChannelChatCreated    bool               `json:"channel_chat_created"`    // optional
	MigrateToChatID       int64              `json:"migrate_to_chat_id"`      // optional
	MigrateFromChatID     int64              `json:"migrate_from_chat_id"`    // optional
	PinnedMessage         *Message           `json:"pinned_message"`          // optional
	Invoice               *Invoice           `json:"invoice"`                 // optional
	SuccessfulPayment     *SuccessfulPayment `json:"successful_payment"`      // optional
	PassportData          *PassportData      `json:"passport_data,omitempty"` // optional
}

type FileBytes struct {
	Name  string
	Bytes []byte
}

type FileReader struct {
	Name   string
	Reader io.Reader
	Size   int64
}

type DocumentConfig struct {
	BaseFile
	Caption   string
	ParseMode string
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

type PhotoConfig struct {
	BaseFile
	Caption   string
	ParseMode string
}

func (e Error) Error() string {
	return e.Message
}

type Chattable interface {
	// contains filtered or unexported methods
}

type Fileable interface {
	Chattable
	params() (map[string]string, error)
	name() string
	getFile() interface{}
	useExistingFile() bool
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

// func (file BaseFile) getFile() interface{} {
// 	return file.File
// }

// func (file BaseFile) params() (map[string]string, error) {
// 	params := make(map[string]string)

// 	if file.ChannelUsername != "" {
// 		params["chat_id"] = file.ChannelUsername
// 	} else {
// 		params["chat_id"] = strconv.FormatInt(file.ChatID, 10)
// 	}

// 	if file.ReplyToMessageID != 0 {
// 		params["reply_to_message_id"] = strconv.Itoa(file.ReplyToMessageID)
// 	}

// 	if file.ReplyMarkup != nil {
// 		data, err := json.Marshal(file.ReplyMarkup)
// 		if err != nil {
// 			return params, err
// 		}

// 		params["reply_markup"] = string(data)
// 	}

// 	if file.MimeType != "" {
// 		params["mime_type"] = file.MimeType
// 	}

// 	if file.FileSize > 0 {
// 		params["file_size"] = strconv.Itoa(file.FileSize)
// 	}

// 	params["disable_notification"] = strconv.FormatBool(file.DisableNotification)

// 	return params, nil
// }

type AnimationConfig struct {
	BaseFile
	Duration  int
	Caption   string
	ParseMode string
}

func NewAnimationUpload(chatID int64, file interface{}) AnimationConfig {
	return AnimationConfig{
		BaseFile: BaseFile{
			BaseChat:    BaseChat{ChatID: chatID},
			File:        file,
			UseExisting: false,
		},
	}
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

//-------?????????????????????????? ????????-------------
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

	// load_list()
}

// ------------???????????????? ?????????????????????? ????????--------------
// func send_notifications(bot *tgbotapi.BotAPI) {
// 	for site, status := range SiteList {
// 		if status != 200 {
// 			alarm := fmt.Sprintf("CRIT - %s ; status: %v", site, status)
// 			bot.Send(tgbotapi.NewMessage(chatID, alarm))
// 		}
// 	}
// }

// ----------???????????????????? ???????????? ???????????? ?????? ????????----------
// func save_list() {
// 	data, err := json.Marshal(SiteList)
// 	if err != nil {
// 		panic(err)
// 	}
// 	err = ioutil.WriteFile(configFile, data, 0644)
// 	if err != nil {
// 		panic(err)
// 	}
// }

// ---------???????????????? ???????????? ???????????? ?????? ????????
// func load_list() {
// 	data, err := ioutil.ReadFile(configFile)
// 	if err != nil {
// 		log.Printf("No such file - starting without config")
// 		return
// 	}

// 	if err = json.Unmarshal(data, &SiteList); err != nil {
// 		log.Printf("Cant read file - starting without config")
// 		return
// 	}
// 	log.Print(data)
// }

// -----------???????????????????? ???????????? ??????????--------
// func monitor(bot *tgbotapi.BotAPI) {

// 	tr := &http.Transport{
// 		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
// 	}

// 	var httpclient = &http.Client{
// 		Timeout:   time.Second * 10,
// 		Transport: tr,
// 	}

// 	for {
// 		save_list()
// 		for site := range SiteList {
// 			response, err := httpclient.Get(site)
// 			if err != nil {
// 				SiteList[site] = 1
// 				log.Printf("Status of %s: %s", site, "1 - Connection refused")
// 			} else {
// 				log.Printf("Status of %s: %s", site, response.Status)
// 				SiteList[site] = response.StatusCode
// 			}
// 		}
// 		send_notifications(bot)
// 		time.Sleep(time.Minute * 5)
// 	}
// }

func (bot *BotAPI) UploadFile(endpoint string, params map[string]string, fieldname string, file interface{}) (APIResponse, error) {
	ms := multipartstreamer.New()

	switch f := file.(type) {
	case string:
		ms.WriteFields(params)

		fileHandle, err := os.Open(f)
		if err != nil {
			return APIResponse{}, err
		}
		defer fileHandle.Close()

		fi, err := os.Stat(f)
		if err != nil {
			return APIResponse{}, err
		}

		ms.WriteReader(fieldname, fileHandle.Name(), fi.Size(), fileHandle)
	case FileBytes:
		ms.WriteFields(params)

		buf := bytes.NewBuffer(f.Bytes)
		ms.WriteReader(fieldname, f.Name, int64(len(f.Bytes)), buf)
	case FileReader:
		ms.WriteFields(params)

		if f.Size != -1 {
			ms.WriteReader(fieldname, f.Name, f.Size, f.Reader)

			break
		}

		data, err := ioutil.ReadAll(f.Reader)
		if err != nil {
			return APIResponse{}, err
		}

		buf := bytes.NewBuffer(data)

		ms.WriteReader(fieldname, f.Name, int64(len(data)), buf)
	case url.URL:
		params[fieldname] = f.String()

		ms.WriteFields(params)
	default:
		return APIResponse{}, errors.New(ErrBadFileType)
	}

	method := fmt.Sprintf(APIEndpoint, bot.Token, endpoint)

	req, err := http.NewRequest("POST", method, nil)
	if err != nil {
		return APIResponse{}, err
	}

	ms.SetupRequest(req)

	res, err := bot.Client.Do(req)
	if err != nil {
		return APIResponse{}, err
	}
	defer res.Body.Close()

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return APIResponse{}, err
	}

	if bot.Debug {
		log.Println(string(bytes))
	}

	var apiResp APIResponse

	err = json.Unmarshal(bytes, &apiResp)
	if err != nil {
		return APIResponse{}, err
	}

	if !apiResp.Ok {
		return APIResponse{}, errors.New(apiResp.Description)
	}

	return apiResp, nil
}

// func (bot *BotAPI) uploadAndSend(method string, config Fileable) (Message, error) {
// 	params, err := config.params()
// 	if err != nil {
// 		return Message{}, err
// 	}

// 	file := config.getFile()

// 	resp, err := bot.UploadFile(method, params, config.name(), file)
// 	if err != nil {
// 		return Message{}, err
// 	}

// 	var message Message
// 	json.Unmarshal(resp.Result, &message)

// 	bot.debugLog(method, nil, message)

// 	return message, nil
// }
