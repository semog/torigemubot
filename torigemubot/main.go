package main

import (
	"flag"
	"log"

	tg "github.com/semog/go-bot-api/v5"
	"k8s.io/klog"
)

var noturns *bool

func main() {
	token := flag.String("token", "Ask @BotFather", "telegram bot token")
	debug := flag.Bool("debug", false, "Show debug information")
	noturns = flag.Bool("noturns", false, "Don't take turns")
	flag.Parse()

	klog.InitFlags(nil)
	if *token == "Ask @BotFather" {
		log.Fatal("token flag required. Go ask @BotFather.")
	}

	log.Print("Connecting...")
	bot, err := tg.NewBotAPI(*token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = *debug
	log.Printf("Connected to Bot: %s (%s)", bot.Self.FirstName, bot.Self.UserName)
	tg.RunBot(bot, torigemubot)
}
