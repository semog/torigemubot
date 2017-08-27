package main

import (
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
		log.Printf("Adding %s to game %d.", formatPlayerName(player), chatID)
		createPlayer(player)
	} else if update {
		log.Printf("Updating %s player info.", formatPlayerName(player))
		savePlayer(player)
	}
	return player
}

func getPlayerByID(chatID int64, playerid int) (*playerEntry, bool) {
	player := &playerEntry{
		chatid: chatID,
		userid: playerid,
	}
	found := queryDb(fmt.Sprintf("SELECT firstname, lastname, username, nickname, score, numwords FROM %s WHERE chatid = %d AND userid = %d",
		playersTableName, player.chatid, player.userid),
		&player.firstname, &player.lastname, &player.username, &player.nickname, &player.score, &player.numWords)
	return player, found
}

func getPlayers(chatID int64) []*playerEntry {
	players := make(playerList, 0)
	// Sort by score ranking.
	rows, err := db.Query(fmt.Sprintf("SELECT userid, firstname, lastname, username, nickname, score, numwords FROM %s WHERE chatid = %d ORDER BY score DESC", playersTableName, chatID))
	defer closeRows(rows)
	if err == nil {
		for rows.Next() {
			player := &playerEntry{
				chatid: chatID,
			}
			rows.Scan(&player.userid, &player.firstname, &player.lastname, &player.username, &player.nickname, &player.score, &player.numWords)
			players = append(players, player)
		}
	}
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

func updatePlayerScore(chatID int64, playerid int, scoreUpdate int) bool {
	var score int
	return queryDb(fmt.Sprintf("SELECT score FROM %s WHERE chatid = %d AND userid = %d", playersTableName, chatID, playerid), &score) &&
		execDb(fmt.Sprintf("UPDATE %s SET score = ? WHERE chatid = ? AND userid = ?", playersTableName),
			score+scoreUpdate, chatID, playerid)
}

func updatePlayerWords(chatID int64, playerid int, wordsUpdate int) bool {
	var numwords int
	return queryDb(fmt.Sprintf("SELECT numwords FROM %s WHERE chatid = %d AND userid = %d", playersTableName, chatID, playerid), &numwords) &&
		execDb(fmt.Sprintf("UPDATE %s SET numwords = ? WHERE chatid = ? AND userid = ?", playersTableName),
			numwords+wordsUpdate, chatID, playerid)
}

func nickNameInUse(chatID int64, nickName string) bool {
	return queryDb(fmt.Sprintf("SELECT userid FROM %s WHERE chatid = %d and nickname = '%s'", playersTableName, chatID, nickName))
}
