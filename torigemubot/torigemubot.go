package main

import (
	"fmt"
	"log"
	"regexp"

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

const challengePenaltyPts = 5
const lostGamePts = 7
const stdWordPts = 3

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

var newgameCmd,
	currentCmd,
	challengeCmd,
	historyCmd,
	scoresCmd,
	nicknameCmd,
	helpCmd,
	shutdownCmd,
	basicCmd,
	kanjiExp *regexp.Regexp

// Initialize global data
func torigemubotOnInitialize(bot *tg.BotAPI) bool {
	if !initDb() {
		log.Println("ERROR: Could not initialize database.")
		return false
	}
	botname := bot.Self.UserName
	newgameCmd = regexp.MustCompile(fmt.Sprintf(`(?i)^/?newgame(@%s)?`, botname))
	currentCmd = regexp.MustCompile(fmt.Sprintf(`(?i)^/?current(@%s)?`, botname))
	challengeCmd = regexp.MustCompile(fmt.Sprintf(`(?i)^/?challenge(@%s)?`, botname))
	historyCmd = regexp.MustCompile(fmt.Sprintf(`(?i)^/?history(@%s)?`, botname))
	scoresCmd = regexp.MustCompile(fmt.Sprintf(`(?i)^/?scores(@%s)?`, botname))
	nicknameCmd = regexp.MustCompile(fmt.Sprintf(`(?i)^/?nick(@%s)?[ \t]*`, botname))
	helpCmd = regexp.MustCompile(fmt.Sprintf(`(?i)^/?help(@%s)?`, botname))
	shutdownCmd = regexp.MustCompile(fmt.Sprintf(`(?i)^/?shutdown(@%s)?[ \t]+now`, botname))
	basicCmd = regexp.MustCompile(`(?i)^/`)
	kanjiExp = regexp.MustCompile(`(\p{Han}|\p{Katakana}|\p{Hiragana}|ー)+`)
	return true
}

func torigemubotOnDispose(bot *tg.BotAPI) {
	// Do any cleanup of external resources.
}

func torigemubotOnMessage(bot *tg.BotAPI, msg *tg.Message) bool {
	log.Printf("MsgFrom: Chat %s, User %s %s (%s): %s",
		formatChatName(msg.Chat), msg.From.FirstName, msg.From.LastName, msg.From.UserName, msg.Text)
	switch {
	case newgameCmd.MatchString(msg.Text):
		doNewGame(bot, msg)
	case currentCmd.MatchString(msg.Text):
		doShowCurrentWord(bot, msg, true)
	case challengeCmd.MatchString(msg.Text):
		doChallenge(bot, msg)
	case historyCmd.MatchString(msg.Text):
		doShowHistory(bot, msg.Chat.ID)
	case scoresCmd.MatchString(msg.Text):
		doShowScores(bot, msg.Chat.ID)
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
	newGame(bot, msg.Chat)
}

func doShowCurrentWord(bot *tg.BotAPI, msg *tg.Message, showUserInfo bool) {
	bot.Send(tg.NewMessage(msg.Chat.ID, getCurrentWordEntryDisplay(msg.Chat, showUserInfo)))
}

func doShowScores(bot *tg.BotAPI, chatID int64) {
	log.Println("Received showscores command.")
	scores := "ゲームの得点は\n＿＿＿＿＿＿＿＿＿＿＿"
	for _, player := range getPlayers(chatID) {
		scores += fmt.Sprintf("\n%s 【%d得点】「%d言葉」", formatPlayerName(player), player.score, player.numWords)
	}
	bot.Send(tg.NewMessage(chatID, scores))
}

func doShowHistory(bot *tg.BotAPI, chatID int64) {
	log.Println("Received showhistory command.")
	wordHistory := "使用された言葉\n＿＿＿＿＿＿＿＿＿＿＿"
	for _, entry := range getWordHistory(chatID) {
		wordHistory += "\n" + getWordEntryDisplay(chatID, entry)
	}
	bot.Send(tg.NewMessage(chatID, wordHistory))
}

func doChallenge(bot *tg.BotAPI, msg *tg.Message) {
	log.Println("Received challenge command.")
	entry := getLastEntry(msg.Chat.ID)
	if entry == nil {
		bot.Send(tg.NewMessage(msg.Chat.ID, "言葉はありません。"))
		return
	}

	// TODO: If two people challenge the word, then remove it. Don't have to wait for original author.
	if entry.userid == msg.From.ID {
		// Player is challenging their own word, so remove it.
		// Remove points for this word. Also add penalty points.
		updatePlayerScore(msg.Chat.ID, entry.userid, -(entry.points + challengePenaltyPts))
		updatePlayerWords(msg.Chat.ID, entry.userid, -1)
		removeLastEntry(msg.Chat.ID)
		bot.Send(tg.NewMessage(msg.Chat.ID, fmt.Sprintf("%sを消しました", entry.word)))
		doShowCurrentWord(bot, msg, true)
	} else {
		player := getPlayer(msg.Chat.ID, msg.From)
		// TODO: Check translation
		bot.Send(tg.NewMessage(msg.Chat.ID,
			fmt.Sprintf("%sは%sに挑戦します。\n準備をして。。。戦う!!!",
				formatPlayerName(player), getWordEntryDisplay(msg.Chat.ID, entry))))
	}
}

func doWordEntry(bot *tg.BotAPI, msg *tg.Message) {
	log.Println("Received a word submission.")
	theWord := msg.Text
	chatID := msg.Chat.ID
	player := getPlayer(chatID, msg.From)
	lastentry := getLastEntry(msg.Chat.ID)
	if lastentry != nil {
		if userSubmittedLastWord(msg, lastentry) {
			bot.Send(tg.NewMessage(chatID, fmt.Sprintf("%s様お待ちください。他の人が最初に行くようにしましょう。\nヽ(^o^)丿", formatPlayerName(player))))
			doShowCurrentWord(bot, msg, false)
			return
		}
		if !respondingToCurrentWord(bot, msg, lastentry) {
			replyMsg := tg.NewMessage(chatID, fmt.Sprintf("ヽ(^o^)丿\n%s様は遅いです。\n現在の言葉は：", formatPlayerName(player)))
			replyMsg.ReplyToMessageID = msg.MessageID
			bot.Send(replyMsg)
			doShowCurrentWord(bot, msg, true)
			return
		}
	}
	if alreadyUsedWord(chatID, theWord) {
		userLostGame(bot, player, fmt.Sprintf("すでに使用されている言葉: %s", theWord))
		newGame(bot, msg.Chat)
		return
	}
	// Checking word validity is a longer operation, so we do it last.
	entryPts := getWordPts(theWord)
	if entryPts == 0 {
		userLostGame(bot, player, fmt.Sprintf("無効言葉: %s", theWord))
		newGame(bot, msg.Chat)
		return
	}
	if getNumEntries(chatID) > 0 {
		updatePlayerScore(chatID, player.userid, entryPts)
		firstEntry := getFirstEntry(chatID)
		if firstEntry.points == 0 {
			firstWordPts := getWordPts(firstEntry.word)
			// Now award the points to the player who went first.
			updateFirstEntryPoints(chatID, firstWordPts)
			updatePlayerScore(chatID, firstEntry.userid, firstWordPts)
		}
	} else {
		// If the first word, then no points awarded until
		// at least one other person goes.
		entryPts = 0
	}
	entry := &wordEntry{
		chatid: chatID,
		word:   theWord,
		userid: player.userid,
		points: entryPts}
	addEntry(entry)
	// TODO: Display the amount of points won/lost for this word entry.
	doShowCurrentWord(bot, msg, false)
}

func doSetNickname(bot *tg.BotAPI, msg *tg.Message, newNickname string) {
	player := getPlayer(msg.Chat.ID, msg.From)
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
		savePlayer(player)
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

func newGame(bot *tg.BotAPI, chat *tg.Chat) {
	clearWordHistory(chat.ID)
	bot.Send(tg.NewMessage(chat.ID, "新しいゲームを開始します。\n誰が最初に行きたいですか？\n(^_^)/"))
}

func getCurrentWordEntryDisplay(chat *tg.Chat, showUserInfo bool) string {
	var entryDisplay string
	entry := getLastEntry(chat.ID)
	if entry == nil {
		entryDisplay = "現在の言葉はありません。"
	} else {
		var wordDisplay string
		if showUserInfo {
			wordDisplay = getWordEntryDisplay(chat.ID, entry)
		} else {
			wordDisplay = entry.word
		}

		entryDisplay = fmt.Sprintf("》%s", wordDisplay)
	}
	return entryDisplay
}

func getWordEntryDisplay(chatID int64, entry *wordEntry) string {
	bonus := ""
	if entry.points > stdWordPts {
		bonus = "★ "
	}
	player, _ := getPlayerByID(chatID, entry.userid)
	return fmt.Sprintf("%s 【%d得点】%s「%s」", entry.word, entry.points, bonus, formatPlayerName(player))
}

func userSubmittedLastWord(msg *tg.Message, lastentry *wordEntry) bool {
	return !*noturns && lastentry.userid == msg.From.ID
}

func userLostGame(bot *tg.BotAPI, player *playerEntry, reason string) {
	updatePlayerScore(player.chatid, player.userid, -lostGamePts)
	bot.Send(tg.NewMessage(player.chatid, fmt.Sprintf("❌%sはゲームを負けました！\n%s\n＿|￣|○", formatPlayerName(player), reason)))
	doShowScores(bot, player.chatid)
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
		// Must have a reply to message, even in direct chats.
		return false
	}
	return lastentry.word == kanjiExp.FindString(msg.ReplyToMessage.Text)
}

func getWordPts(theWord string) int {
	// TODO
	// Get wordentry of new word.
	// If word score is zero or not found, then return zero. Probably ends in 'no'.
	// Get wordentry of current word.
	// If first kana of new word does not match ending kana of current word, then return zero.
	// Return word points.
	return len(kanjiExp.FindString(theWord))
}
