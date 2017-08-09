package main

import (
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
showhistory - Show the words that have been used in the game.
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

// WordMap will track the current word for each game.
type WordMap map[int64]string

// WordHistoryMap type to track the words used for each game.
type WordHistoryMap map[int64][]string

var currentWord WordMap
var usedWords WordHistoryMap

// PlayerDataMap tracks players in each game
type PlayerDataMap map[int64][]*tg.User

// TODO: Enhance the players struct to track their current score.
var players PlayerDataMap

func torigemubotOnMessage(bot *tg.BotAPI, msg *tg.Message) bool {
	msgTextCmd := strings.ToLower(msg.Text)

	log.Printf("MsgFrom: Chat %s, User %s %s", getGameName(msg.Chat), getUserDisplayName(msg.From), msg.Text)
	switch {
	case strings.HasPrefix(msgTextCmd, "/newgame"):
		doNewGame(bot, msg)
	case strings.HasPrefix(msgTextCmd, "/showword"):
		doShowWord(bot, msg)
	case strings.HasPrefix(msgTextCmd, "/showscores"):
		doShowScores(bot, msg)
	case strings.HasPrefix(msgTextCmd, "/showhistory"):
		doShowHistory(bot, msg)
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
	log.Printf("OnInlineQuery: %s %s", getUserDisplayName(query.From), query.Query)
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
	bot.Send(tg.NewMessage(msg.Chat.ID, fmt.Sprintf("New game started by %s.\nWho wants to go first?", getUserDisplayName(msg.From))))
}

func doShowWord(bot *tg.BotAPI, msg *tg.Message) {
	var reply string
	if currentWord == nil || len(currentWord[msg.Chat.ID]) == 0 {
		reply = "There is no current word."
	} else {
		reply = fmt.Sprintf("Current word: %s", currentWord[msg.Chat.ID])
	}
	bot.Send(tg.NewMessage(msg.Chat.ID, reply))
}

func doShowScores(bot *tg.BotAPI, msg *tg.Message) {
	log.Println("Received showscores command.")
	bot.Send(tg.NewMessage(msg.Chat.ID, "Four Score and Seven Words Ago..."))
}

func doShowHistory(bot *tg.BotAPI, msg *tg.Message) {
	log.Println("Received showhistory command.")
	bot.Send(tg.NewMessage(msg.Chat.ID, "Words used so far:\n"+strings.Join(usedWords[msg.Chat.ID], "\n")))
}

func doChallenge(bot *tg.BotAPI, msg *tg.Message) {
	log.Println("Received challenge command.")
	bot.Send(tg.NewMessage(msg.Chat.ID, "Ready.... FIGHT!!!"))
}

func doWordEntry(bot *tg.BotAPI, msg *tg.Message) {
	log.Println("Received a NEW word!.")
	// Auto-join the game.
	joinGame(bot, msg)
	if currentWord == nil {
		currentWord = make(WordMap)
	}
	currentWord[msg.Chat.ID] = msg.Text

	if usedWords == nil {
		usedWords = make(WordHistoryMap)
	}
	usedWords[msg.Chat.ID] = append(usedWords[msg.Chat.ID], msg.Text)

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
	if currentWord != nil {
		delete(currentWord, chat.ID)
	}
	if usedWords != nil {
		delete(usedWords, chat.ID)
	}
	if players != nil {
		delete(players, chat.ID)
	}
}

func joinGame(bot *tg.BotAPI, msg *tg.Message) {
	if players == nil {
		players = make(PlayerDataMap)
	}

	player := msg.From
	if _, index := findPlayer(msg.Chat.ID, player); index < 0 {
		log.Printf("Adding player %s to game %s.", getUserDisplayName(player), getGameName(msg.Chat))
		players[msg.Chat.ID] = append(players[msg.Chat.ID], player)
		bot.Send(tg.NewMessage(msg.Chat.ID, fmt.Sprintf("%s joined the game.", getUserDisplayName(player))))
	}
}

func getGameName(chat *tg.Chat) string {
	switch chat.Type {
	case "group":
		return fmt.Sprintf("%s [%d]", chat.Title, chat.ID)
	default:
		return fmt.Sprintf("%s %s [%d]", chat.FirstName, chat.LastName, chat.ID)
	}
}

func getUserDisplayName(user *tg.User) string {
	return fmt.Sprintf("%s %s (%s)", user.FirstName, user.LastName, user.UserName)
}

// For when a user leaves the chat.
func leaveGame(bot *tg.BotAPI, msg *tg.Message) {
	player, index := findPlayer(msg.Chat.ID, msg.From)
	if player == nil {
		return
	}

	currentPlayers := players[msg.Chat.ID]
	players[msg.Chat.ID] = append(currentPlayers[:index], currentPlayers[index+1:]...)
	bot.Send(tg.NewMessage(msg.Chat.ID, fmt.Sprintf("%s left the game.", getUserDisplayName(player))))
	if len(players[msg.Chat.ID]) < 2 {
		// Game over.
		newGame(msg.Chat)
	}
}

func findPlayer(chatID int64, user *tg.User) (*tg.User, int) {
	if players == nil {
		return nil, -1
	}

	for i, player := range players[chatID] {
		if player.FirstName == user.FirstName &&
			player.LastName == user.LastName &&
			player.UserName == user.UserName {
			return player, i
		}
	}
	return nil, -1
}
