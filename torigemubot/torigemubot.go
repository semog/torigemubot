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
current - Show the current word.
challenge - Challenge the word that was entered.
history - Show the words that have been used in the game.
scores - Show the current scores.
help - Display game rules and other instructions.
shutdown - Shutdown the bot (DEBUG ONLY)
*/

var torigemubot = botEventHandlers{
	onInitialize: torigemubotOnInitialize,
	onDispose:    torigemubotOnDispose,
	onMessage:    torigemubotOnMessage,
}

type wordEntry struct {
	word   string
	player *tg.User
}

type wordHistoryMap map[int64][]wordEntry

// Track the words used for each game.
var usedWords wordHistoryMap

type playerMap map[int64][]*tg.User

// TODO: Enhance the players struct to track their current score.
// Track players in each game
var players playerMap

// Initialize global data
func torigemubotOnInitialize(bot *tg.BotAPI) bool {
	usedWords = make(wordHistoryMap)
	players = make(playerMap)
	return true
}

func torigemubotOnDispose(bot *tg.BotAPI) {
	// TODO: Any cleanup of external resources.
}

func torigemubotOnMessage(bot *tg.BotAPI, msg *tg.Message) bool {
	msgTextCmd := strings.ToLower(msg.Text)

	log.Printf("MsgFrom: Chat %s, User %s %s", getGameName(msg.Chat), getUserDisplayName(msg.From), msg.Text)
	switch {
	case strings.HasPrefix(msgTextCmd, "/newgame"):
		doNewGame(bot, msg)
	case strings.HasPrefix(msgTextCmd, "/current"):
		doShowCurrentWord(bot, msg, true)
	case strings.HasPrefix(msgTextCmd, "/challenge"):
		doChallenge(bot, msg)
	case strings.HasPrefix(msgTextCmd, "/history"):
		doShowHistory(bot, msg)
	case strings.HasPrefix(msgTextCmd, "/scores"):
		doShowScores(bot, msg)
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

func doNewGame(bot *tg.BotAPI, msg *tg.Message) {
	log.Println("Received newgame command.")
	// TODO: Add some safety checks so that one other person must agree. Give a lazy consensus time.
	// If no one objects within 1 minute, then the game starts new. If someone agrees, it starts new right away.
	// If someone objects, then the reset is canceled.
	newGame(bot, msg.Chat, false)
	joinGame(bot, msg.From, msg.Chat, false)
}

func doShowCurrentWord(bot *tg.BotAPI, msg *tg.Message, showUserInfo bool) {
	var reply string
	entry := getLastEntry(msg.Chat)
	if entry == nil {
		reply = "現在の言葉はありません。"
	} else {
		var wordDisplay string
		if showUserInfo {
			wordDisplay = getWordEntryDisplay(*entry)
		} else {
			wordDisplay = entry.word
		}

		reply = fmt.Sprintf("現在の言葉は: %s", wordDisplay)
	}
	bot.Send(tg.NewMessage(msg.Chat.ID, reply))
}

func doShowScores(bot *tg.BotAPI, msg *tg.Message) {
	log.Println("Received showscores command.")
	// TODO: Tanslate
	bot.Send(tg.NewMessage(msg.Chat.ID, "Four Score and Seven Words Ago..."))
}

func doShowHistory(bot *tg.BotAPI, msg *tg.Message) {
	log.Println("Received showhistory command.")
	wordHistory := "使用された言葉:"
	for _, entry := range usedWords[msg.Chat.ID] {
		wordHistory += "\n" + getWordEntryDisplay(entry)
	}
	bot.Send(tg.NewMessage(msg.Chat.ID, wordHistory))
}

func doChallenge(bot *tg.BotAPI, msg *tg.Message) {
	log.Println("Received challenge command.")
	entry := getLastEntry(msg.Chat)
	if entry == nil {
		// TODO: Tanslate
		bot.Send(tg.NewMessage(msg.Chat.ID, "No words to challenge."))
	} else if entry.player.ID == msg.From.ID {
		// Player is challenging their own word, so remove it.
		currentWords := usedWords[msg.Chat.ID]
		usedWords[msg.Chat.ID] = currentWords[:len(currentWords)-1]
		// TODO: Tanslate
		bot.Send(tg.NewMessage(msg.Chat.ID, "Removed last word."))
		doShowCurrentWord(bot, msg, true)
	} else {
		// TODO: Tanslate
		bot.Send(tg.NewMessage(msg.Chat.ID,
			fmt.Sprintf("%s challenges %s.\nReady.... FIGHT!!!",
				getUserDisplayName(msg.From), getWordEntryDisplay(*entry))))
	}
}

func doWordEntry(bot *tg.BotAPI, msg *tg.Message) {
	log.Println("Received a word submission.")
	theWord := msg.Text
	// Auto-create the game.
	createGame(bot, msg.Chat)
	// Auto-join the game.
	joinGame(bot, msg.From, msg.Chat, true)

	if userSubmittedLastWord(msg) {
		// TODO: Tanslate
		bot.Send(tg.NewMessage(msg.Chat.ID, "Wait your turn."))
		doShowCurrentWord(bot, msg, false)
		return
	}

	if alreadyUsedWord(msg.Chat, theWord) {
		userLostGame(bot, msg, fmt.Sprintf("すでに使用されている言葉: %s", theWord))
		newGame(bot, msg.Chat, false)
		return
	}

	submission := wordEntry{
		word:   theWord,
		player: msg.From}
	usedWords[msg.Chat.ID] = append(usedWords[msg.Chat.ID], submission)

	// TODO: Calculate scores. If only one person, then no scores awarded.
	// TODO: Display the name of person and amount of points won/lost for this word entry.
	doShowCurrentWord(bot, msg, false)
}

func doHelp(bot *tg.BotAPI, msg *tg.Message) {
	log.Println("Received help command.")
	bot.Send(tg.NewMessage(msg.Chat.ID,
		`Game Rules
＿＿＿＿＿＿＿＿＿＿＿
① Two or more people take turns to play.
② Only nouns are permitted.
③ A player who plays a word ending in the mora N 「ん」 loses the game, as no Japanese word begins with that character.
④ Words may not be repeated.
⑤ Phrases connected by no 「の」 are permitted, but only in those cases where the phrase is sufficiently fossilized to be considered a "word".
⑥ When a word ends in a small kana, such as 「じてんしゃ」 (bicycle), continue with the しゃ combination, such as 「しゃこ」 (garage).

Example: sakura 「さくら」 → rajio 「ラジオ」 → onigiri 「おにぎり」 → risu 「りす」 → sumou 「すもう」 → udon 「うどん」

The player who used the word udon lost this game.`))
}

func doShutdown(bot *tg.BotAPI, msg *tg.Message) {
	log.Println("Received shutdown command.")
	bot.Send(tg.NewMessage(msg.Chat.ID, "シャットダウン。。。"))
}

// TODO: Do a new game if the bot is kicked out of chat.
func newGame(bot *tg.BotAPI, chat *tg.Chat, withNewPlayers bool) {
	delete(usedWords, chat.ID)
	if withNewPlayers {
		delete(players, chat.ID)
	}
	bot.Send(tg.NewMessage(chat.ID, "新しいゲームを開始します。\n誰が最初に行きたいですか？"))
}

func createGame(bot *tg.BotAPI, chat *tg.Chat) {
	if len(players[chat.ID]) == 0 {
		newGame(bot, chat, true)
	}
}

func joinGame(bot *tg.BotAPI, player *tg.User, chat *tg.Chat, announce bool) {
	if _, index := findPlayer(chat.ID, player); index < 0 {
		log.Printf("Adding %s to game %s.", getUserDisplayName(player), getGameName(chat))
		players[chat.ID] = append(players[chat.ID], player)
		// TODO: Temporarily commented out to lower the chat volume.
		if false && announce {
			bot.Send(tg.NewMessage(chat.ID, fmt.Sprintf("%sはゲームに参加しました。", getUserDisplayName(player))))
		}
	}
}

// For when a user leaves the chat.
func leaveGame(bot *tg.BotAPI, msg *tg.Message) {
	player, index := findPlayer(msg.Chat.ID, msg.From)
	if player == nil {
		return
	}

	currentPlayers := players[msg.Chat.ID]
	// Remove the player
	players[msg.Chat.ID] = append(currentPlayers[:index], currentPlayers[index+1:]...)
	bot.Send(tg.NewMessage(msg.Chat.ID, fmt.Sprintf("%sはゲームを去った。", getUserDisplayName(player))))
	if len(players[msg.Chat.ID]) < 2 {
		// Game over.
		newGame(bot, msg.Chat, true)
	}
}

func getWordEntryDisplay(entry wordEntry) string {
	//【】『』「」
	return fmt.Sprintf("%s 「%s」", entry.word, getUserDisplayName(entry.player))
}

func getLastEntry(chat *tg.Chat) *wordEntry {
	numWords := len(usedWords[chat.ID])
	if numWords == 0 {
		return nil
	}

	return &usedWords[chat.ID][numWords-1]
}

func userSubmittedLastWord(msg *tg.Message) bool {
	entry := getLastEntry(msg.Chat)
	if entry == nil {
		return false
	}

	return entry.player.ID == msg.From.ID
}

func userLostGame(bot *tg.BotAPI, msg *tg.Message, reason string) {
	// TODO: Deduct score points.
	bot.Send(tg.NewMessage(msg.Chat.ID, fmt.Sprintf("❌%sはゲームを負けました！\n%s", getUserDisplayName(msg.From), reason)))
}

func findPlayer(chatID int64, user *tg.User) (*tg.User, int) {
	for i, player := range players[chatID] {
		if player.FirstName == user.FirstName &&
			player.LastName == user.LastName &&
			player.UserName == user.UserName {
			return player, i
		}
	}
	return nil, -1
}

func alreadyUsedWord(chat *tg.Chat, theWord string) bool {
	wordCheck := strings.ToLower(theWord)
	for _, usedWord := range usedWords[chat.ID] {
		if wordCheck == strings.ToLower(usedWord.word) {
			return true
		}
	}
	return false
}

func getGameName(chat *tg.Chat) string {
	switch chat.Type {
	case "group":
		return fmt.Sprintf("%s [%d]", chat.Title, chat.ID)
	default:
		gameName := chat.FirstName
		if len(chat.LastName) != 0 {
			gameName += fmt.Sprintf(" %s", chat.LastName)
		}
		if len(chat.UserName) != 0 {
			gameName += fmt.Sprintf(" (@%s)", chat.UserName)
		}
		return fmt.Sprintf("%s [%d]", gameName, chat.ID)
	}
}

func getUserDisplayName(user *tg.User) string {
	displayName := user.FirstName
	if len(user.LastName) != 0 {
		displayName += fmt.Sprintf(" %s", user.LastName)
	}
	if len(user.UserName) != 0 {
		displayName += fmt.Sprintf(" (@%s)", user.UserName)
	}
	return displayName
}
