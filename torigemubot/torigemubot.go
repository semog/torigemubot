package main

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"

	tg "github.com/go-telegram-bot-api/telegram-bot-api"
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
nick - Set your nickname.
help - Display game rules and other instructions.
*/

var torigemubot = botEventHandlers{
	onInitialize: torigemubotOnInitialize,
	onDispose:    torigemubotOnDispose,
	onMessage:    torigemubotOnMessage,
}

const wordPts = 3
const firstWordPts = 2
const challengePenaltyPts = 5
const lostGamePts = 7

// Track players in each game
type playerEntry struct {
	id        int
	user      *tg.User
	firstname string
	lastname  string
	username  string
	nickname  string
	score     int
	numWords  int
}
type playerList []*playerEntry
type playerMap map[int64]playerList

var players playerMap

// Track the words used for each game.
type wordEntry struct {
	word   string
	player *playerEntry
	points int
}
type wordHistoryMap map[int64][]*wordEntry

var usedWords wordHistoryMap

// Initialize global data
func torigemubotOnInitialize(bot *tg.BotAPI) bool {
	if !initDb() {
		log.Println("ERROR: Could not initialize database.")
		return false
	}
	usedWords = make(wordHistoryMap)
	players = make(playerMap)
	return true
}

func torigemubotOnDispose(bot *tg.BotAPI) {
	// TODO: Any cleanup of external resources.
	// TODO: saving is temporary. Refactor to operate straight out of database.
	//savePlayers()
}

var newgameCmd = regexp.MustCompile(`(?i)^/?newgame(@torigemubot)?`)
var currentCmd = regexp.MustCompile(`(?i)^/?current(@torigemubot)?`)
var challengeCmd = regexp.MustCompile(`(?i)^/?challenge(@torigemubot)?`)
var historyCmd = regexp.MustCompile(`(?i)^/?history(@torigemubot)?`)
var scoresCmd = regexp.MustCompile(`(?i)^/?scores(@torigemubot)?`)
var nicknameCmd = regexp.MustCompile(`(?i)^/?nick(@torigemubot)?[ \t]*`)
var helpCmd = regexp.MustCompile(`(?i)^/?help(@torigemubot)?`)
var shutdownCmd = regexp.MustCompile(`(?i)^/?shutdown(@torigemubot)?[ \t]+now`)
var basicCmd = regexp.MustCompile(`(?i)^/`)
var kanjiExp = regexp.MustCompile(`(\p{Han}|\p{Katakana}|\p{Hiragana}|ー)+`)

func torigemubotOnMessage(bot *tg.BotAPI, msg *tg.Message) bool {
	log.Printf("MsgFrom: Chat %s, User %s %s", formatChatName(msg.Chat), formatUserName(msg.From), msg.Text)
	switch {
	case newgameCmd.MatchString(msg.Text):
		doNewGame(bot, msg)
	case currentCmd.MatchString(msg.Text):
		doShowCurrentWord(bot, msg, true)
	case challengeCmd.MatchString(msg.Text):
		doChallenge(bot, msg)
	case historyCmd.MatchString(msg.Text):
		doShowHistory(bot, msg)
	case scoresCmd.MatchString(msg.Text):
		doShowScores(bot, msg.Chat)
	case nicknameCmd.MatchString(msg.Text):
		doSetNickname(bot, msg, nicknameCmd.ReplaceAllString(msg.Text, ""))
	case helpCmd.MatchString(msg.Text):
		doHelp(bot, msg)
	case shutdownCmd.MatchString(msg.Text):
		doShutdown(bot, msg)
		return false
	// Don't add a word that was a command attempt.
	case !basicCmd.MatchString(msg.Text) && len(msg.Text) > 0:
		doWordEntry(bot, msg)
	}
	return true
}

func doNewGame(bot *tg.BotAPI, msg *tg.Message) {
	log.Println("Received newgame command.")
	// TODO: Add some safety checks so that one other person must agree. Give a lazy consensus time.
	// If no one objects within 1 minute, then the game starts new. If someone agrees, it starts new right away.
	// If someone objects, then the reset is canceled.
	newGame(bot, msg.Chat, false, false)
	joinGame(bot, msg.From, msg.Chat, false)
}

func doShowCurrentWord(bot *tg.BotAPI, msg *tg.Message, showUserInfo bool) {
	bot.Send(tg.NewMessage(msg.Chat.ID, getCurrentWordEntryDisplay(msg.Chat, showUserInfo)))
}

func doShowScores(bot *tg.BotAPI, chat *tg.Chat) {
	log.Println("Received showscores command.")
	scores := "ゲームの得点は\n＿＿＿＿＿＿＿＿＿＿＿"
	// Sort by score ranking.
	sort.Sort(players[chat.ID])
	for _, player := range players[chat.ID] {
		scores += fmt.Sprintf("\n%s 【%d得点】「%d言葉」", formatPlayerName(player), player.score, player.numWords)
	}
	bot.Send(tg.NewMessage(chat.ID, scores))
}

func doShowHistory(bot *tg.BotAPI, msg *tg.Message) {
	log.Println("Received showhistory command.")
	wordHistory := "使用された言葉\n＿＿＿＿＿＿＿＿＿＿＿"
	for _, entry := range usedWords[msg.Chat.ID] {
		wordHistory += "\n" + getWordEntryDisplay(entry)
	}
	bot.Send(tg.NewMessage(msg.Chat.ID, wordHistory))
}

func doChallenge(bot *tg.BotAPI, msg *tg.Message) {
	log.Println("Received challenge command.")
	entry := getLastEntry(msg.Chat)
	if entry == nil {
		bot.Send(tg.NewMessage(msg.Chat.ID, "言葉はありません。"))
		return
	}

	if entry.player.user.ID == msg.From.ID {
		// Player is challenging their own word, so remove it.
		currentWords := usedWords[msg.Chat.ID]
		// Remove points for this word. Also add penalty points.
		entry.player.score -= entry.points + challengePenaltyPts
		usedWords[msg.Chat.ID] = currentWords[:len(currentWords)-1]
		bot.Send(tg.NewMessage(msg.Chat.ID, fmt.Sprintf("%sを消しました", entry.word)))
		doShowCurrentWord(bot, msg, true)
	} else {
		// Auto-join the game.
		player := joinGame(bot, msg.From, msg.Chat, true)
		// TODO: Check tanslation
		bot.Send(tg.NewMessage(msg.Chat.ID,
			fmt.Sprintf("%sは%sに挑戦します。\n準備をして。。。戦う!!!",
				formatPlayerName(player), getWordEntryDisplay(entry))))
	}
}

func doWordEntry(bot *tg.BotAPI, msg *tg.Message) {
	log.Println("Received a word submission.")
	theWord := msg.Text
	chatID := msg.Chat.ID
	// Auto-create the game.
	createGame(bot, msg.Chat)
	// Auto-join the game.
	player := joinGame(bot, msg.From, msg.Chat, true)
	if userSubmittedLastWord(msg) {
		bot.Send(tg.NewMessage(chatID, fmt.Sprintf("%s様お待ちください。他の人が最初に行くようにしましょう。\nヽ(^o^)丿", formatPlayerName(player))))
		doShowCurrentWord(bot, msg, false)
		return
	}
	if !respondingToCurrentWord(bot, msg) {
		replyMsg := tg.NewMessage(chatID, fmt.Sprintf("ヽ(^o^)丿\n%s様は遅いです。\n現在の言葉は：", formatPlayerName(player)))
		replyMsg.ReplyToMessageID = msg.MessageID
		bot.Send(replyMsg)
		doShowCurrentWord(bot, msg, true)
		return
	}
	if alreadyUsedWord(chatID, theWord) {
		userLostGame(bot, player, msg, fmt.Sprintf("すでに使用されている言葉: %s", theWord))
		newGame(bot, msg.Chat, false, false)
		return
	}
	if !isValidWord(theWord) {
		userLostGame(bot, player, msg, fmt.Sprintf("無効言葉: %s", theWord))
		newGame(bot, msg.Chat, false, false)
		return
	}

	// Calculate points. If the first word, then no points awarded.
	entryPts := 0
	if getNumWords(chatID) > 0 {
		entryPts = wordPts
		firstWord := usedWords[chatID][0]
		if firstWord.points == 0 {
			// Now award the points to the player who went first.
			firstWord.points = firstWordPts
			firstWord.player.score += firstWordPts
		}
	}
	player.score += entryPts
	player.numWords++
	submission := &wordEntry{
		word:   theWord,
		player: player,
		points: entryPts}
	usedWords[chatID] = append(usedWords[chatID], submission)
	// TODO: Display the amount of points won/lost for this word entry.
	doShowCurrentWord(bot, msg, false)
}

func doSetNickname(bot *tg.BotAPI, msg *tg.Message, newNickname string) {
	// Auto-join the game.
	player := joinGame(bot, msg.From, msg.Chat, false)
	oldName := formatPlayerName(player)
	if oldName == newNickname {
		reply := tg.NewMessage(msg.Chat.ID, "新しい名前を入力して下さい。\n(^_^)/")
		reply.ReplyToMessageID = msg.MessageID
		bot.Send(reply)
	} else if nickNameInUse(msg.Chat.ID, newNickname) {
		reply := tg.NewMessage(msg.Chat.ID, "その名前は取られます。\nm(_ _)m")
		reply.ReplyToMessageID = msg.MessageID
		bot.Send(reply)
	} else {
		player.nickname = newNickname
		bot.Send(tg.NewMessage(msg.Chat.ID, fmt.Sprintf("(@^^)/~~~\n%s様は今から%s様とよんでます。", oldName, formatPlayerName(player))))
	}
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
func newGame(bot *tg.BotAPI, chat *tg.Chat, autostarted bool, withNewPlayers bool) {
	delete(usedWords, chat.ID)
	if withNewPlayers {
		delete(players, chat.ID)
	}
	promptText := ""
	if !autostarted {
		promptText = "\n誰が最初に行きたいですか？"
	}
	bot.Send(tg.NewMessage(chat.ID, fmt.Sprintf("新しいゲームを開始します。%s\n(^_^)/", promptText)))
}

func createGame(bot *tg.BotAPI, chat *tg.Chat) {
	if len(players[chat.ID]) == 0 {
		newGame(bot, chat, true, true)
	}
}

func joinGame(bot *tg.BotAPI, user *tg.User, chat *tg.Chat, announce bool) *playerEntry {
	player, _ := findPlayer(chat.ID, user)
	if player == nil {
		player = &playerEntry{user: user}
		playerName := formatPlayerName(player)
		log.Printf("Adding %s to game %s.", playerName, formatChatName(chat))
		players[chat.ID] = append(players[chat.ID], player)
		// TODO: Temporarily commented out to lower the chat volume.
		if false && announce {
			bot.Send(tg.NewMessage(chat.ID, fmt.Sprintf("%sはゲームに参加しました。", playerName)))
		}
	}
	return player
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
	bot.Send(tg.NewMessage(msg.Chat.ID, fmt.Sprintf("%sはゲームを去った。", formatPlayerName(player))))
	if len(players[msg.Chat.ID]) < 2 {
		// Game over. No one to play with.
		newGame(bot, msg.Chat, false, true)
	}
}

func findPlayer(chatID int64, user *tg.User) (*playerEntry, int) {
	for i, player := range players[chatID] {
		if player.user.FirstName == user.FirstName &&
			player.user.LastName == user.LastName &&
			player.user.UserName == user.UserName {
			return player, i
		}
	}
	return nil, -1
}

func getCurrentWordEntryDisplay(chat *tg.Chat, showUserInfo bool) string {
	var entryDisplay string
	entry := getLastEntry(chat)
	if entry == nil {
		entryDisplay = "現在の言葉はありません。"
	} else {
		var wordDisplay string
		if showUserInfo {
			wordDisplay = getWordEntryDisplay(entry)
		} else {
			wordDisplay = entry.word
		}

		entryDisplay = fmt.Sprintf("》%s", wordDisplay)
	}
	return entryDisplay
}

func getWordEntryDisplay(entry *wordEntry) string {
	var bonus = ""
	if entry.points > wordPts {
		bonus = "★"
	}
	return fmt.Sprintf("%s 【%d%s得点】「%s」", entry.word, entry.points, bonus, formatPlayerName(entry.player))
}

func getLastEntry(chat *tg.Chat) *wordEntry {
	numWords := len(usedWords[chat.ID])
	if numWords == 0 {
		return nil
	}
	return usedWords[chat.ID][numWords-1]
}

func userSubmittedLastWord(msg *tg.Message) bool {
	entry := getLastEntry(msg.Chat)
	if entry == nil {
		return false
	}
	return !*noturns && entry.player.user.ID == msg.From.ID
}

func userLostGame(bot *tg.BotAPI, player *playerEntry, msg *tg.Message, reason string) {
	player.score -= lostGamePts
	bot.Send(tg.NewMessage(msg.Chat.ID, fmt.Sprintf("❌%sはゲームを負けました！\n%s\n＿|￣|○", formatPlayerName(player), reason)))
	doShowScores(bot, msg.Chat)
}

func nickNameInUse(chatID int64, nickName string) bool {
	for _, player := range players[chatID] {
		if player.nickname == nickName {
			return true
		}
	}
	return false
}

func alreadyUsedWord(chatID int64, theWord string) bool {
	wordCheck := strings.ToLower(theWord)
	for _, usedWord := range usedWords[chatID] {
		if wordCheck == strings.ToLower(usedWord.word) {
			return true
		}
	}
	return false
}

func formatChatName(chat *tg.Chat) string {
	switch chat.Type {
	case "group":
		return fmt.Sprintf("%s [%d]", chat.Title, chat.ID)
	default:
		chatName := chat.FirstName
		if len(chat.LastName) != 0 {
			chatName += fmt.Sprintf(" %s", chat.LastName)
		}
		if len(chat.UserName) != 0 {
			chatName += fmt.Sprintf(" (@%s)", chat.UserName)
		}
		return fmt.Sprintf("%s [%d]", chatName, chat.ID)
	}
}

func formatPlayerName(player *playerEntry) string {
	if len(player.nickname) == 0 {
		player.nickname = formatUserName(player.user)
	}
	return player.nickname
}

func formatUserName(user *tg.User) string {
	displayname := user.FirstName
	if len(user.LastName) != 0 {
		displayname += fmt.Sprintf(" %s", user.LastName)
	}
	if len(user.UserName) != 0 {
		displayname += fmt.Sprintf(" (@%s)", user.UserName)
	}
	return displayname
}

func isValidWord(theWord string) bool {
	// TODO: Do database lookup of the noun word.
	return len(kanjiExp.FindString(theWord)) > 0
}

func respondingToCurrentWord(bot *tg.BotAPI, msg *tg.Message) bool {
	entry := getLastEntry(msg.Chat)
	if entry == nil {
		// No words yet, so any response is valid.
		return true
	}
	if msg.ReplyToMessage == nil {
		// Must have a reply to message, even in direct chats.
		return false
	}
	log.Printf("Matching %s to %s", entry.word, kanjiExp.FindString(msg.ReplyToMessage.Text))
	return entry.word == kanjiExp.FindString(msg.ReplyToMessage.Text)
}

func getNumWords(chatID int64) int {
	return len(usedWords[chatID])
}

// Implement Len(), Less() and Swap() for sorting.
func (entries playerList) Len() int {
	return len(entries)
}

func (entries playerList) Less(i, j int) bool {
	return entries[i].score > entries[j].score
}

func (entries playerList) Swap(i, j int) {
	entries[i], entries[j] = entries[j], entries[i]
}
