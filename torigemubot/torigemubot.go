package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	tg "github.com/semog/telegram-bot-api"
)

/*
Following is the command menu for constructing the bot with @BotFather.
Use the /setcommands command and reply with the following list of commands.
---------------------
current - Show the current word.
history - Show the words that have been used in the game.
scores - Show the current scores.
nick - Set your nickname.
add - Add a custom word to this group's game.
remove - Remove a custom word from this group's game.
help - Display game rules and other instructions.
*/

var torigemubot = tg.BotEventHandlers{
	OnInitialize: torigemubotOnInitialize,
	OnDispose:    torigemubotOnDispose,
	OnMessage:    torigemubotOnMessage,
}

const lostGamePts = 3
const addWordPts = 1

const newGamePrompt = "始める新しい単語を入力して下さい。"

// TODO: Add cleanup of game data from the database if a chat is destroyed, or the bot is kicked out (same thing).

// Track players in each game
type playerEntry struct {
	chatid    int64
	userid    int
	firstname string
	lastname  string
	username  string
	nickname  string
	score     int
	numWords  int
}
type playerList []*playerEntry

// Track the words used for each game.
type wordEntry struct {
	chatid int64
	userid int
	word   string
	points int
}
type wordList []*wordEntry

var currentCmd,
	historyCmd,
	scoresCmd,
	nicknameCmd,
	helpCmd,
	shutdownCmd,
	addCmd,
	removeCmd *regexp.Regexp
var basicCmd = regexp.MustCompile(`(?i)^/`)
var kanjiExp = regexp.MustCompile(`(\p{Han}|\p{Katakana}|\p{Hiragana}|ー)+`)
var addCustomWordExp = regexp.MustCompile(`(?i)([\p{Han}|\p{Katakana}|\p{Hiragana}|ー]+)[ 　\t]+([\p{Hiragana}|,|、]+)`)
var removeCustomWordExp = regexp.MustCompile(`(?i)([\p{Han}|\p{Katakana}|\p{Hiragana}|ー]+)`)

// Initialize global data
func torigemubotOnInitialize(bot *tg.BotAPI) bool {
	if !initgameDb() {
		log.Println("ERROR: Could not initialize database.")
		return false
	}
	botname := bot.Self.UserName
	currentCmd = regexp.MustCompile(fmt.Sprintf(`(?i)^/?current(@%s)?`, botname))
	historyCmd = regexp.MustCompile(fmt.Sprintf(`(?i)^/?history(@%s)?`, botname))
	scoresCmd = regexp.MustCompile(fmt.Sprintf(`(?i)^/?scores(@%s)?`, botname))
	nicknameCmd = regexp.MustCompile(fmt.Sprintf(`(?i)^/?nick(@%s)?[ 　\t]*`, botname))
	helpCmd = regexp.MustCompile(fmt.Sprintf(`(?i)^/?help(@%s)?`, botname))
	addCmd = regexp.MustCompile(fmt.Sprintf(`(?i)^/?add(@%s)?[ 　\t]*`, botname))
	removeCmd = regexp.MustCompile(fmt.Sprintf(`(?i)^/?remove(@%s)?[ 　\t]*`, botname))
	shutdownCmd = regexp.MustCompile(fmt.Sprintf(`(?i)^/?shutdown(@%s)?[ \t]+now`, botname))
	return true
}

func torigemubotOnDispose(bot *tg.BotAPI) {
	// Do any cleanup of external resources.
}

func torigemubotOnMessage(bot *tg.BotAPI, msg *tg.Message) bool {
	log.Printf("MsgFrom: Chat %s, User %s %s (%s): %s",
		formatChatName(msg.Chat), msg.From.FirstName, msg.From.LastName, msg.From.UserName, msg.Text)
	switch {
	case !basicCmd.MatchString(msg.Text) && len(msg.Text) > 0:
		doWordEntry(bot, msg)
	case currentCmd.MatchString(msg.Text):
		doShowCurrentWord(bot, msg, true)
	case historyCmd.MatchString(msg.Text):
		doShowHistory(bot, msg.Chat.ID)
	case scoresCmd.MatchString(msg.Text):
		doShowScores(bot, msg.Chat.ID)
	case nicknameCmd.MatchString(msg.Text):
		doSetNickname(bot, msg, nicknameCmd.ReplaceAllString(msg.Text, ""))
	case addCmd.MatchString(msg.Text):
		doAddWord(bot, msg, addCmd.ReplaceAllString(msg.Text, ""))
	case removeCmd.MatchString(msg.Text):
		doRemoveWord(bot, msg, removeCmd.ReplaceAllString(msg.Text, ""))
	case helpCmd.MatchString(msg.Text):
		doHelp(bot, msg)
	case shutdownCmd.MatchString(msg.Text):
		doShutdown(bot, msg)
		return false
	}
	return true
}

func doShowCurrentWord(bot *tg.BotAPI, msg *tg.Message, showUserInfo bool) {
	bot.Send(tg.NewMessage(msg.Chat.ID, getCurrentWordEntryDisplay(msg.Chat, showUserInfo)))
}

func doShowScores(bot *tg.BotAPI, chatID int64) {
	log.Println("Received showscores command.")
	scores := "*ゲームの得点は*\n＿＿＿＿＿＿＿＿＿＿＿"
	for _, player := range getPlayers(chatID) {
		scores += fmt.Sprintf("\n%s 【%d得点】「%d言葉」", formatPlayerName(player), player.score, player.numWords)
	}
	msg := tg.NewMessage(chatID, scores)
	msg.ParseMode = tg.ParseModeMarkdown
	bot.Send(msg)
}

func doShowHistory(bot *tg.BotAPI, chatID int64) {
	log.Println("Received showhistory command.")
	wordHistory := "*使用された言葉*\n＿＿＿＿＿＿＿＿＿＿＿"
	for _, entry := range getWordHistory(chatID) {
		wordHistory += "\n" + getWordEntryDisplay(chatID, entry, true)
	}
	msg := tg.NewMessage(chatID, wordHistory)
	msg.ParseMode = tg.ParseModeMarkdown
	bot.Send(msg)
}

func doWordEntry(bot *tg.BotAPI, msg *tg.Message) {
	log.Println("Received a word submission.")
	theWord := msg.Text
	chatID := msg.Chat.ID
	player := getPlayer(chatID, msg.From)
	lastentry := getLastEntry(msg.Chat.ID)
	// Private chats don't have to take turns.
	if lastentry != nil && !msg.Chat.IsPrivate() {
		if userSubmittedLastWord(msg, lastentry) {
			bot.Send(tg.NewMessage(chatID, fmt.Sprintf("%s様お待ち下さい。他の人が最初に行くようにしましょう。\nヽ(^o^)丿", formatPlayerName(player))))
			doShowCurrentWord(bot, msg, false)
			return
		}
		if !respondingToCurrentWord(bot, msg, lastentry) {
			sendReplyMsg(bot, msg, fmt.Sprintf("ヽ(^o^)丿\n%s様は遅いです。\n現在の言葉は：", formatPlayerName(player)))
			doShowCurrentWord(bot, msg, true)
			return
		}
	}

	// Even if the second word is invalid, the first word points need to be applied.
	gamedb.BeginTrans()
	defer gamedb.CommitTrans()
	// If the first word, then no points awarded until at least one other person goes.
	firstword := true
	firstEntry := getFirstEntry(chatID)
	if firstEntry != nil {
		firstword = false
		if firstEntry.points == 0 {
			firstWordPts, _ := getWordPts(chatID, firstEntry.word, nil)
			// Now award the points to the player who went first.
			updateFirstEntryPoints(chatID, firstWordPts)
			updatePlayerScore(chatID, firstEntry.userid, firstWordPts)
		}
	}
	if alreadyUsedWord(chatID, theWord) {
		userLostGame(bot, player, fmt.Sprintf("すでに使用されている言葉: %s", theWord))
		newGame(bot, msg.Chat)
		return
	}
	// Checking word validity is a longer operation, so we do it last.
	entryPts, ptsMsg := getWordPts(chatID, theWord, lastentry)
	if entryPts == 0 {
		userLostGame(bot, player, ptsMsg)
		newGame(bot, msg.Chat)
		return
	}

	if !firstword {
		updatePlayerScore(chatID, player.userid, entryPts)
	} else {
		entryPts = 0
	}
	addEntry(&wordEntry{
		chatid: chatID,
		word:   theWord,
		userid: player.userid,
		points: entryPts})
	doShowCurrentWord(bot, msg, false)
}

func doSetNickname(bot *tg.BotAPI, msg *tg.Message, newNickname string) {
	log.Println("Received set nickname command.")
	player := getPlayer(msg.Chat.ID, msg.From)
	oldName := formatPlayerName(player)
	if oldName == newNickname {
		sendReplyMsg(bot, msg, "新しい名前を入力して下さい。\n(^_^)/")
	} else if nickNameInUse(msg.Chat.ID, newNickname) {
		sendReplyMsg(bot, msg, "その名前は取られます。\nm(_ _)m")
	} else {
		player.nickname = newNickname
		savePlayer(player)
		bot.Send(tg.NewMessage(msg.Chat.ID, fmt.Sprintf("(@^^)/~~~\n%s様は今から%s様とよんでます。", oldName, formatPlayerName(player))))
	}
}

// First parameter is the kanji, second parameter is hiragana pronunciation (can be comma-separated list of multiple pronunciations).
func doAddWord(bot *tg.BotAPI, msg *tg.Message, wordVal string) {
	log.Println("Received add custom word command.")
	// Extract the kanji and kana definitions for this custom word.
	customWord := addCustomWordExp.FindStringSubmatch(wordVal)
	if len(customWord) < 3 {
		sendReplyMsg(bot, msg, "❌誤りです。漢字とひらがながありません。")
		return
	}
	// Replace any hirigana commas with regular commas
	kanji := customWord[1]
	kana := strings.Replace(customWord[2], "、", ",", -1)
	if endsInN(kana) {
		sendReplyMsg(bot, msg, fmt.Sprintf("❌誤りです。無効言葉: %s「%s」。言葉はんを終わることができない。", kanji, kana))
		return
	}
	wordExists, _, _ := lookupStandardKana(msg.Chat.ID, kanji)
	if wordExists {
		sendReplyMsg(bot, msg, fmt.Sprintf("❌言葉は既に存在します：　%s「%s」。", kanji, kana))
		return
	}
	if !addCustomWord(msg.Chat.ID, msg.From.ID, kanji, kana) {
		sendReplyMsg(bot, msg, fmt.Sprintf("❌誤りです。言葉を追加できませんでした：　%s「%s」。", kanji, kana))
		return
	}
	sendReplyMsg(bot, msg, fmt.Sprintf("追加された言葉：　%s「%s」。ありがとうございました！", kanji, kana))
}

func doRemoveWord(bot *tg.BotAPI, msg *tg.Message, wordVal string) {
	log.Println("Received remove custom word command.")
	// Extract the kanji to be removed.
	customWord := removeCustomWordExp.FindStringSubmatch(wordVal)
	if len(customWord) < 2 {
		sendReplyMsg(bot, msg, "❌誤りです。漢字がありません。")
		return
	}
	kanji := customWord[0]
	if !removeCustomWord(msg.Chat.ID, kanji) {
		sendReplyMsg(bot, msg, fmt.Sprintf("言葉を削除できませんでした：　%s.", kanji))
		return
	}
	sendReplyMsg(bot, msg, fmt.Sprintf("削除された：　%s.", kanji))
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
⑦ If the word is katakana and ends in ー, then start with the vowel sound that it is extending. For example:
      マスター, next word should start with あ.
      ルビー, next word should start with い.

Example: sakura 「さくら」 → rajio 「ラジオ」 → onigiri 「おにぎり」 → risu 「りす」 → sumou 「すもう」 → udon 「うどん」

The player who used the word udon lost this game.`))
}

func doShutdown(bot *tg.BotAPI, msg *tg.Message) {
	log.Println("Received shutdown command.")
	bot.Send(tg.NewMessage(msg.Chat.ID, "シャットダウン。。。"))
}

func sendReplyMsg(bot *tg.BotAPI, msg *tg.Message, message string) {
	reply := tg.NewMessage(msg.Chat.ID, message)
	reply.ReplyToMessageID = msg.MessageID
	bot.Send(reply)
}

func newGame(bot *tg.BotAPI, chat *tg.Chat) {
	clearWordHistory(chat.ID)
	bot.Send(tg.NewMessage(chat.ID, fmt.Sprintf("新しいゲームを開始します。\n%s\n(^_^)/", newGamePrompt)))
}

func getCurrentWordEntryDisplay(chat *tg.Chat, showUserInfo bool) string {
	var entryDisplay string
	entry := getLastEntry(chat.ID)
	if entry == nil {
		entryDisplay = newGamePrompt
	} else {
		entryDisplay = fmt.Sprintf("》%s", getWordEntryDisplay(chat.ID, entry, showUserInfo))
	}
	return entryDisplay
}

func getWordEntryDisplay(chatID int64, entry *wordEntry, showUserInfo bool) string {
	playername := ""
	bonus := ""
	pts := entry.points
	if pts == 0 {
		// The points haven't been awarded yet, so we calc them and flag the entry.
		pts = calcWordPoints(entry.word)
		bonus += "★"
	}
	if showUserInfo {
		player, _ := getPlayerByID(chatID, entry.userid)
		playername = fmt.Sprintf("「%s」", formatPlayerName(player))
	}
	return fmt.Sprintf("%s【%d得点】%s%s", entry.word, pts, bonus, playername)
}

func userSubmittedLastWord(msg *tg.Message, lastentry *wordEntry) bool {
	return !*noturns && lastentry.userid == msg.From.ID
}

func userLostGame(bot *tg.BotAPI, player *playerEntry, reason string) {
	updatePlayerScore(player.chatid, player.userid, -lostGamePts)
	bot.Send(tg.NewMessage(player.chatid, fmt.Sprintf("❌%s様はゲームを負けました！\n%s\n＿|￣|○", formatPlayerName(player), reason)))
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
		player.nickname = player.firstname
		if len(player.lastname) != 0 {
			player.nickname += fmt.Sprintf(" %s", player.lastname)
		}
		if len(player.username) != 0 {
			player.nickname += fmt.Sprintf(" (@%s)", player.username)
		}
	}
	return player.nickname
}

func respondingToCurrentWord(bot *tg.BotAPI, msg *tg.Message, lastentry *wordEntry) bool {
	if msg.ReplyToMessage == nil {
		// Must have a reply to message.
		return false
	}
	return lastentry.word == kanjiExp.FindString(msg.ReplyToMessage.Text)
}
