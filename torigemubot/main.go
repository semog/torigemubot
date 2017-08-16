package main

import (
	"flag"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
)

var noturns *bool

func main() {
	token := flag.String("token", "Ask @BotFather", "telegram bot token")
	debug := flag.Bool("debug", false, "Show debug information")
	noturns = flag.Bool("noturns", false, "Don't take turns")
	flag.Parse()

	if *token == "Ask @BotFather" {
		log.Fatal("token flag required. Go ask @BotFather.")
	}

	log.Print("Connecting...")
	bot, err := tgbotapi.NewBotAPI(*token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = *debug
	log.Printf("Connected to Bot: %s (%s)", bot.Self.FirstName, bot.Self.UserName)
	runBot(bot, torigemubot)
}
