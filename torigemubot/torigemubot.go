package main

import (
	"container/list"
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strings"
)

/*
Following is the command menu for constructing the bot with @BotFather.
Use the /setcommands command and reply with the following list of commands.
---------------------
newgame - Start a new game with a fresh word list.
showword - Show the current word.
showscores - Show the current scores.
challenge - Challenge the word that was entered.
help - Display game rules and other instructions.
shutdown - Shutdown the bot (DEBUG ONLY)
*/

var torigemubot = botEventHandlers{
	onMessage:     torigemubotOnMessage,
	onInlineQuery: torigemubotOnInlineQuery,
	//	onChosenInlineResult: torigemubotOnChosenInlineResult,
}

// TODO: Keep historical record of words played. Will be used to verify words are not reused, and for reverting after a challenge.
// TODO: Create a map of chatID -> currentWord, chatID -> players.
var currentWord string

// TODO: Enhance the players struct to track their current score.
var players *list.List

func torigemubotOnMessage(bot *tg.BotAPI, msg *tg.Message) bool {
	msgTextCmd := strings.ToLower(msg.Text)

	log.Printf("[%s] %s", msg.From.UserName, msg.Text)
	switch {
	case strings.HasPrefix(msgTextCmd, "/newgame"):
		doNewGame(bot, msg)
	case strings.HasPrefix(msgTextCmd, "/showword"):
		doShowWord(bot, msg)
	case strings.HasPrefix(msgTextCmd, "/showscores"):
		doShowScores(bot, msg)
	case strings.HasPrefix(msgTextCmd, "/challenge"):
		doChallenge(bot, msg)
	case strings.HasPrefix(msgTextCmd, "/help"):
		doHelp(bot, msg)
	case strings.HasPrefix(msgTextCmd, "/shutdown"):
		doShutdown(bot, msg)
		return false
	default:
		doWordEntry(bot, msg)
	}
	return true
}

func torigemubotOnInlineQuery(bot *tg.BotAPI, query *tg.InlineQuery) bool {
	log.Printf("OnInlineQuery: %s %s [%s] %s", query.From.FirstName, query.From.LastName, query.From.UserName, query.Query)
	var answer = tg.InlineConfig{
		InlineQueryID: query.ID,
		IsPersonal:    true}

	if len(query.Query) > 0 {
		answer.Results = append(answer.Results, tg.NewInlineQueryResultArticle(query.ID, "回答を提出する", query.Query))
	}
	bot.AnswerInlineQuery(answer)
	return true
}

func doNewGame(bot *tg.BotAPI, msg *tg.Message) {
	log.Println("Received newgame command.")
	// TODO: Add some safety checks so that one other person must agree. Give a lazy consensus time.
	// If noone objects within 1 minute, then the game starts new. If someone agrees, it starts new right away.
	// If someone objects, then the reset is canceled.
	newGame(msg.Chat)
	joinGame(bot, msg)
	bot.Send(tg.NewMessage(msg.Chat.ID, fmt.Sprintf("New game started by %s %s.\nWho wants to go first?", msg.From.FirstName, msg.From.LastName)))
}

// For when a user leaves the chat.
func doLeave(bot *tg.BotAPI, msg *tg.Message) {
	log.Println("Received leave command.")
	pelem := findPlayer(msg.Chat, msg.From)
	if pelem == nil {
		return
	}

	players.Remove(pelem)
	player := pelem.Value.(*tg.User)
	bot.Send(tg.NewMessage(msg.Chat.ID, fmt.Sprintf("%s %s left the game.", player.FirstName, player.LastName)))
	if players.Len() == 0 {
		// Game over.
		newGame(msg.Chat)
	}
}

func doShowWord(bot *tg.BotAPI, msg *tg.Message) {
	var reply string
	if len(currentWord) == 0 {
		reply = "There is no current word."
	} else {
		reply = fmt.Sprintf("Current word: %s", currentWord)
	}
	bot.Send(tg.NewMessage(msg.Chat.ID, reply))
}

func doShowScores(bot *tg.BotAPI, msg *tg.Message) {
	log.Println("Received showscores command.")
	bot.Send(tg.NewMessage(msg.Chat.ID, "Four Score and Seven Words Ago..."))
}

func doChallenge(bot *tg.BotAPI, msg *tg.Message) {
	log.Println("Received challenge command.")
	bot.Send(tg.NewMessage(msg.Chat.ID, "Ready.... FIGHT!!!"))
}

func doWordEntry(bot *tg.BotAPI, msg *tg.Message) {
	log.Println("Received a NEW word!.")
	// Auto-join the game.
	joinGame(bot, msg)
	currentWord = msg.Text
	// TODO: Calculate scores. If only one person, then no scores awarded.
	// TODO: Display the name of person and amount of points won/lost for this word entry.
	doShowWord(bot, msg)
}

func doHelp(bot *tg.BotAPI, msg *tg.Message) {
	log.Println("Received help command.")
	bot.Send(tg.NewMessage(msg.Chat.ID, "Bots help those who help themselves."))
}

func doShutdown(bot *tg.BotAPI, msg *tg.Message) {
	log.Println("Received shutdown command.")
	bot.Send(tg.NewMessage(msg.Chat.ID, "Shutting down..."))
}

// TODO: Do a new game if the bot is kicked out of chat.
func newGame(chat *tg.Chat) {
	currentWord = ""
	players = nil
}

func joinGame(bot *tg.BotAPI, msg *tg.Message) {
	if players == nil {
		players = list.New()
	}

	player := msg.From
	if findPlayer(msg.Chat, player) == nil {
		var gamename string

		switch msg.Chat.Type {
		case "group":
			gamename = fmt.Sprintf("%s [%d]", msg.Chat.Title, msg.Chat.ID)
		default:
			gamename = fmt.Sprintf("%s %s [%d]", msg.Chat.FirstName, msg.Chat.LastName, msg.Chat.ID)
		}
		log.Printf("Adding player %s %s [%s] to game %s.", player.FirstName, player.LastName, player.UserName, gamename)
		players.PushBack(player)
		bot.Send(tg.NewMessage(msg.Chat.ID, fmt.Sprintf("%s %s joined the game.", player.FirstName, player.LastName)))
	}
}

func findPlayer(chat *tg.Chat, user *tg.User) *list.Element {
	if players == nil {
		return nil
	}

	var player *tg.User
	for e := players.Front(); e != nil; e = e.Next() {
		player = e.Value.(*tg.User)
		if player.FirstName == user.FirstName &&
			player.LastName == user.LastName &&
			player.UserName == user.UserName {
			return e
		}
	}
	return nil
}
