package main

import (
	"database/sql"
	"fmt"
	"log"

	tg "github.com/semog/go-bot-api/v5"
)

const playersTableName = "players"

func getPlayer(chatID int64, user *tg.User) *playerEntry {
	var update = false
	player, err := getPlayerByID(chatID, user.ID)
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

	if err != nil {
		log.Printf("Adding %s to game [%d].", formatPlayerName(player), chatID)
		createPlayer(player)
	} else if update {
		log.Printf("Updating %s player info.", formatPlayerName(player))
		savePlayer(player)
	}
	return player
}

func getPlayerByID(chatID int64, userid int64) (*playerEntry, error) {
	player := &playerEntry{
		chatid: chatID,
		userid: userid,
	}
	err := gamedb.SingleQuery(fmt.Sprintf("SELECT firstname, lastname, username, nickname, score, numwords FROM %s WHERE chatid = %d AND userid = %d",
		playersTableName, player.chatid, player.userid),
		&player.firstname, &player.lastname, &player.username, &player.nickname, &player.score, &player.numWords)
	return player, err
}

func getPlayers(chatID int64) []*playerEntry {
	players := make(playerList, 0)
	// Sort by score ranking.
	gamedb.MultiQuery(fmt.Sprintf("SELECT userid, firstname, lastname, username, nickname, score, numwords FROM %s WHERE chatid = %d ORDER BY score DESC", playersTableName, chatID),
		func(rows *sql.Rows) error {
			player := &playerEntry{
				chatid: chatID,
			}
			rows.Scan(&player.userid, &player.firstname, &player.lastname, &player.username, &player.nickname, &player.score, &player.numWords)
			players = append(players, player)
			return nil
		})
	return players
}

func createPlayer(player *playerEntry) error {
	return gamedb.Exec(fmt.Sprintf("INSERT INTO %s (chatid, userid, firstname, lastname, username, nickname, score, numwords) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", playersTableName),
		player.chatid, player.userid, player.firstname, player.lastname, player.username, player.nickname, player.score, player.numWords)
}

func savePlayer(player *playerEntry) error {
	return gamedb.Exec(fmt.Sprintf("UPDATE %s SET firstname = ?, lastname = ?, username = ?, nickname = ?, score = ?, numwords = ? WHERE chatid = ? AND userid = ?", playersTableName),
		player.firstname, player.lastname, player.username, player.nickname, player.score, player.numWords, player.chatid, player.userid)
}

func updatePlayerScore(chatID int64, userid int64, scoreUpdate int) error {
	return gamedb.Exec(fmt.Sprintf("UPDATE %s SET score = (SELECT score+%d FROM %s WHERE chatid = %d AND userid = %d) WHERE chatid = %d AND userid = %d",
		playersTableName, scoreUpdate, playersTableName, chatID, userid, chatID, userid))
}

func updatePlayerWords(chatID int64, userid int64, wordsUpdate int) error {
	return gamedb.Exec(fmt.Sprintf("UPDATE %s SET numwords = (SELECT numwords+%d FROM %s WHERE chatid = %d AND userid = %d) WHERE chatid = %d AND userid = %d",
		playersTableName, wordsUpdate, playersTableName, chatID, userid, chatID, userid))
}

func nickNameInUse(chatID int64, nickName string) bool {
	return nil == gamedb.SingleQuery(fmt.Sprintf("SELECT userid FROM %s WHERE chatid = %d and nickname = '%s'", playersTableName, chatID, nickName))
}
