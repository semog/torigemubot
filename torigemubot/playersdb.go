package main

import (
	"database/sql"
	"fmt"
	"log"

	tg "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/mattn/go-sqlite3"
)

const playersTableName = "players"

func getPlayer(chatID int64, user *tg.User) *playerEntry {
	var update = false
	player, found := getPlayerByID(chatID, user.ID)
	if player.firstname != user.FirstName {
		player.firstname = user.FirstName
		update = true
	}
	if player.lastname != user.LastName {
		player.lastname = user.LastName
		update = true
	}
	if player.username != user.UserName {
		player.username = user.UserName
		update = true
	}

	if !found {
		log.Printf("Adding %s to game [%d].", formatPlayerName(player), chatID)
		createPlayer(player)
	} else if update {
		log.Printf("Updating %s player info.", formatPlayerName(player))
		savePlayer(player)
	}
	return player
}

func getPlayerByID(chatID int64, userid int) (*playerEntry, bool) {
	player := &playerEntry{
		chatid: chatID,
		userid: userid,
	}
	found := queryDb(fmt.Sprintf("SELECT firstname, lastname, username, nickname, score, numwords FROM %s WHERE chatid = %d AND userid = %d",
		playersTableName, player.chatid, player.userid),
		&player.firstname, &player.lastname, &player.username, &player.nickname, &player.score, &player.numWords)
	return player, found
}

func getPlayers(chatID int64) []*playerEntry {
	players := make(playerList, 0)
	// Sort by score ranking.
	multiQueryDb(fmt.Sprintf("SELECT userid, firstname, lastname, username, nickname, score, numwords FROM %s WHERE chatid = %d ORDER BY score DESC", playersTableName, chatID),
		func(rows *sql.Rows) bool {
			player := &playerEntry{
				chatid: chatID,
			}
			rows.Scan(&player.userid, &player.firstname, &player.lastname, &player.username, &player.nickname, &player.score, &player.numWords)
			players = append(players, player)
			return true
		})
	return players
}

func createPlayer(player *playerEntry) bool {
	return execDb(fmt.Sprintf("INSERT INTO %s (chatid, userid, firstname, lastname, username, nickname, score, numwords) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", playersTableName),
		player.chatid, player.userid, player.firstname, player.lastname, player.username, player.nickname, player.score, player.numWords)
}

func savePlayer(player *playerEntry) bool {
	return execDb(fmt.Sprintf("UPDATE %s SET firstname = ?, lastname = ?, username = ?, nickname = ?, score = ?, numwords = ? WHERE chatid = ? AND userid = ?", playersTableName),
		player.firstname, player.lastname, player.username, player.nickname, player.score, player.numWords, player.chatid, player.userid)
}

func updatePlayerScore(chatID int64, userid int, scoreUpdate int) bool {
	return execDb(fmt.Sprintf("UPDATE %s SET score = (SELECT score+%d FROM %s WHERE chatid = %d AND userid = %d) WHERE chatid = %d AND userid = %d",
		playersTableName, scoreUpdate, playersTableName, chatID, userid, chatID, userid))
}

func updatePlayerWords(chatID int64, userid int, wordsUpdate int) bool {
	return execDb(fmt.Sprintf("UPDATE %s SET numwords = (SELECT numwords+%d FROM %s WHERE chatid = %d AND userid = %d) WHERE chatid = %d AND userid = %d",
		playersTableName, wordsUpdate, playersTableName, chatID, userid, chatID, userid))
}

func nickNameInUse(chatID int64, nickName string) bool {
	return queryDb(fmt.Sprintf("SELECT userid FROM %s WHERE chatid = %d and nickname = '%s'", playersTableName, chatID, nickName))
}
