package main

import (
	"encoding/json"
	"flag"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
	"html"
	"io"
	"os"
	"os/exec"
)

const (
	ErrBotSend = "bot send"
	BtnPost    = "Yes, Post content 🆗"
	BtnNo      = "No 🚫"
	pathFile   = "data/externalPost"
)

type (
	Config struct {
		BotApiKey string `json:"botApiKey"`
	}

	// Content from bot
	Content struct {
		Title       string
		Description string
		ChatText    string
		ChatId      int64
	}

	// Article for posting to target source
	Article struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	Poster interface {
		Post() error
	}

	Bot struct {
		config *Config
	}
	BotI interface {
		Run()
	}
)

func checkErr(e error, msg string) {
	if e != nil {
		log.Errorln(e, " err "+msg)
	}
}

func main() {
	var (
		pathConfig string
		bot        BotI
	)

	flag.StringVar(&pathConfig, "config-path", "config.json", "path to config file")
	flag.Parse()

	config, err := NewConfig(&pathConfig)
	if err != nil {
		log.Fatal(err)
	}

	bot = &Bot{config: config}
	bot.Run()
}

func NewConfig(path *string) (*Config, error) {
	var conf *Config

	// Open our jsonFile
	jsonFile, err := os.Open(*path)
	if err != nil {
		return nil, err
	}

	defer func(jsonFile *os.File) {
		if err = jsonFile.Close(); err != nil {
			log.Error(err)
		}
	}(jsonFile)

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(byteValue, &conf)

	return conf, err
}

func (a *Article) Post() error {
	// create dir if it doesn't exist
	if _, err := os.Stat(pathFile); os.IsNotExist(err) {
		if err = os.Mkdir(pathFile, os.ModePerm); err != nil {
			log.Errorln(err, "create folder for export data")
		}
	}

	f, err := os.Create(fmt.Sprintf("%s/%s.json", pathFile, a.Title))
	if err != nil {
		return err
	}

	defer func(f *os.File) {
		err = f.Close()
		checkErr(err, "close file")
	}(f)

	p, err := json.Marshal(a)
	if err != nil {
		return err
	}

	if _, err = f.Write(p); err != nil {
		return err
	}

	if err = f.Sync(); err != nil {
		return err
	}

	if err = createWebPage(a); err != nil {
		return err
	}

	return nil
}

// createWebPage create page with HUGO cli
// command create article with hugo cli "hugo new posts/<title-name>.md"
// in folder archetypes/posts.md has snippet
// to detect file in data/externalPost/<title-name>.json and create new article
func createWebPage(article *Article) error {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("hugo new \"posts/%s.md\"", article.Title))

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (b *Bot) Run() {
	var add = &Content{}

	bot, err := tgbotapi.NewBotAPI(b.config.BotApiKey)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message updates
			continue
		}

		// Create a new MessageConfig. We don't have text yet, so we leave it empty.
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		// handle get content after command "/add"
		if add.ChatId == update.Message.Chat.ID {
			handleAddCommand(&update, &msg, add)

			_, err = bot.Send(msg)
			checkErr(err, ErrBotSend)
			continue
		}

		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)

		msg.Text = handleBotCommands(&update, add)

		if _, err = bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}

}

// handleAddCommand get content for article after triggered command '/add'
func handleAddCommand(update *tgbotapi.Update, msg *tgbotapi.MessageConfig, add *Content) {
	switch update.Message.Text {
	case BtnPost:
		article := Poster(&Article{add.Title, add.Description})
		if err := article.Post(); err != nil {
			checkErr(err, "Post content")
			msg.Text = "⚠️ Error posting: " + err.Error()
		} else {
			msg.Text = "Successfully posted"
		}
		add = &Content{}
	case BtnNo:
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		add = &Content{}
	default:
		add.ChatText = update.Message.Text
		var numericKeyboardPost = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(BtnPost),
				tgbotapi.NewKeyboardButton(BtnNo),
			),
		)

		if add.Title == "" {
			add.Title = html.EscapeString(add.ChatText)
			msg.Text = "Please write Post description:"
			return
		}

		msg.ParseMode = "markdown"

		if add.Description == "" {
			add.Description = add.ChatText
			msg.ReplyMarkup = numericKeyboardPost
			msg.Text = fmt.Sprintf("Do you confirm add Post? \n title:*%s* \n description: \n %s", add.Title, add.Description)
			return
		}
	}
}

func handleBotCommands(update *tgbotapi.Update, add *Content) string {
	switch update.Message.Command() {
	case "help":
		return "I understand \n" +
			" /sayhi \n" +
			" /status \n" +
			" /add - Add new article to website (more details: https://github.com/oleksiy-os/bot-content-post ) \n"
	case "sayhi":
		return "Hi :)"
	case "status":
		return "I'm ok."
	case "add":
		add.ChatId = update.Message.Chat.ID
		return "Write Post title:"
	default:
		return "I don't know that command. This bot just for tests. Probably it doesn't work now"
	}
}
