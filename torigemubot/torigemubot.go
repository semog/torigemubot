package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strings"
)

/*
Following is the command menu for constructing the bot with @BotFather.
Use the /setcommands command and reply with the following list of commands.
---------------------
start - Start a new game.
join - Join the game.
leave - Leave the game.
challenge - Challenge the word that was entered.
help - Display game rules and other instructions.
shutdown - Shutdown the bot (DEBUG ONLY)
*/

var torigemubot = botEventHandlers{
	onMessage:     torigemubotOnMessage,
	onInlineQuery: torigemubotOnInlineQuery}

func torigemubotOnMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) bool {
	msgTextCmd := strings.ToLower(msg.Text)

	log.Printf("[%s] %s", msg.From.UserName, msg.Text)
	if strings.HasPrefix(msgTextCmd, "/shutdown") {
		log.Println("Received shutdown command.")
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Shutting down..."))
		return false
	}

	replyMsg := tgbotapi.NewMessage(msg.Chat.ID, msg.From.FirstName+" said: "+msg.Text)
	//replyMsg.ReplyToMessageID = msg.MessageID
	bot.Send(replyMsg)
	return true
}

func torigemubotOnInlineQuery(bot *tgbotapi.BotAPI, query *tgbotapi.InlineQuery) bool {
	log.Println("Received inline query from " + query.From.FirstName + ": " + query.Query)
	var answer = tgbotapi.InlineConfig{
		InlineQueryID: query.ID,
		IsPersonal:    true}

	if len(query.Query) > 0 {
		answer.Results = append(answer.Results, tgbotapi.NewInlineQueryResultArticle(query.ID, "回答を提出する", query.Query))
	}
	bot.AnswerInlineQuery(answer)
	return true
}
