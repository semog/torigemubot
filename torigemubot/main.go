package main

import (
	"flag"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
)

func main() {
	token := flag.String("token", "Ask @BotFather", "telegram bot token")
	debug := flag.Bool("debug", false, "show debug information")
	flag.Parse()

	if *token == "Ask @BotFather" {
		log.Fatal("token flag required. Go ask @BotFather.")
	}

	log.Printf("Connecting to Bot Token: %s", *token)
	bot, err := tgbotapi.NewBotAPI(*token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = *debug
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Auto Reply: "+update.Message.Text)
		//msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)
		if update.Message.Text == "/quit" {
			log.Println("Received quit command.")
			break
		}
	}

	log.Printf("Shutting down %s", bot.Self.UserName)
}
