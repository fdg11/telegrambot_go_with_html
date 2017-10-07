package main

import (
	"flag"
	"gopkg.in/telegram-bot-api.v4"
	"log"
	"os"
	"fmt"
	"net/http"
	"path"
	"html/template"
)

var (
	// глобальная переменная в которой храним токен и chatID
	telegramBotToken string
	chatID int64
	sliceMsg map[string]string = make(map[string]string)
)

func init() {
	// принимаем на входе флаг -telegrambottoken и chatid
	flag.StringVar(&telegramBotToken, "telegrambottoken", "", "Telegram Bot Token")
	flag.Int64Var(&chatID, "chatid", 0, "chatId to send messages")
	flag.Parse()

	// без него не запускаемся
	if telegramBotToken == "" {
		log.Print("-telegrambottoken is required")
		os.Exit(1)
	}

	if chatID == 0 {
		log.Print("-chatid is required")
		os.Exit(1)
	}

}

func client(w http.ResponseWriter, r *http.Request) {
	fp := path.Join("", "client.html")
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Pass template to http.ResponseWriter.
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func process(w http.ResponseWriter, r *http.Request) {
	bot, err := tgbotapi.NewBotAPI(telegramBotToken)
	if err != nil {
		log.Panic(err)
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	r.ParseForm()
	log.Printf("Test name: %s", r.PostForm)

	sliceMsg["name"] = r.PostFormValue("name")
	sliceMsg["phone"] = r.PostFormValue("phone")
	sliceMsg["message"] = r.PostFormValue("messages")

	if sliceMsg["name"] == "" {
		fmt.Fprintln(w, "Не заполнено поле 'Ваше Имя'")
	} else if len(sliceMsg["phone"]) > 12 || sliceMsg["phone"] == ""  {
		fmt.Fprintln(w, "Не заполнено поле 'Телефон', либо не коретно введен номер")
	} else if sliceMsg["message"] == "" {
		fmt.Fprintln(w, "Не заполнено поле 'Сообщение'")
	} else {
		fmt.Fprintln(w, "Сообщение отправлено")
		text := fmt.Sprintf(
			"`%s`\n" +
				"*Имя отправителя:* _%s_\n" +
				"*Телефон:* _%s_\n" +
				"*Сообщение:* %s\n",
				"Сообщение с сайта",
				sliceMsg["name"],
				sliceMsg["phone"],
				sliceMsg["message"])
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "markdown"
		bot.Send(msg)
	}
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
	bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("%s", "Бот запущен, ожидает сообщений с сайт")))

	server := http.Server{
		Addr: ":8080",
	}
	http.HandleFunc("/", client)
	http.HandleFunc("/process", process)
	http.Handle("/modalform/", http.StripPrefix("/modalform/", http.FileServer(http.Dir("modalform"))))
	server.ListenAndServe()

	// используя конфиг u создаем канал в который будут прилетать новые сообщения
	updates, err := bot.GetUpdatesChan(u)

	// в канал updates прилетают структуры типа Update
	// вычитываем их и обрабатываем
	for update := range updates {
		// универсальный ответ на любое сообщение
		_ = "Не знаю что сказать"
		if update.Message == nil {
			continue
		}

		// логируем от кого какое сообщение пришло
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		log.Printf("[%s] %s", update.Message.Chat.ID, update.Message.Text)
		// свитч на обработку комманд
		// комманда - сообщение, начинающееся с "/"
		switch update.Message.Command() {
		case "start":
			_ = "Привет. Я телеграм-бот"
		case "hello":
			_ = "world"
		}

		// создаем ответное сообщение
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		//msg.ReplyToMessageID = update.Message.MessageID
		// отправляем
		bot.Send(msg)
	}

}
