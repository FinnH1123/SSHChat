package server

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/FinnH1123/SSHChat/config"
)

type DBMessage struct {
	Id      int
	Content string
	RoomId  string
	Sender  string
	Sent    time.Time
}

func CreateMessage(db *sql.DB, message Message, roomid string) (DBMessage, error) {
	var msg DBMessage
	sqlStatement := `
	INSERT INTO messages (content, sender, room, sent_on)
	VALUES ($1, $2, $3, $4)
	returning * 
	`
	err := db.QueryRow(sqlStatement, message.content, message.sender, roomid, time.Now()).Scan(&msg.Id, &msg.Content, &msg.RoomId, &msg.Sender, &msg.Sent)
	return msg, err
}

func GetRoomMessages(db *sql.DB, roomid string) ([]Message, error) {
	var msgs []Message
	sqlStatement := `
	Select * FROM messages 
	WHERE room = $1
	ORDER BY sent_on ASC
	`

	rows, err := db.Query(sqlStatement, roomid)
	msg := DBMessage{}
	if rows == nil {
		return []Message{}, nil
	}
	defer rows.Close()
	for rows.Next() { // creating new struct for every row
		err = rows.Scan(&msg.Id, &msg.Content, &msg.Sender, &msg.RoomId, &msg.Sent)
		if err != nil {
			log.Println(err)
		}
		msgs = append(msgs, Message{content: msg.Content, sender: msg.Sender}) // add new row information
	}

	if err != nil {
		fmt.Println(err)
	}
	return msgs, err
}

func InitialiseDB() (*sql.DB, error) {
	//initialiseDB
	config := config.GetConfig()
	conninfo := fmt.Sprintf("user=%s password=%s host=%s sslmode=disable", config.Database.Username, config.Database.Password, config.Database.Host)
	db, err := sql.Open("postgres", conninfo)

	if err != nil || db == nil {
		fmt.Println(err)
		return nil, err
	}

	_, err = db.Exec("create database " + config.Database.Dbname)
	if (err != nil || db == nil) && !strings.Contains(err.Error(), "already exists") {
		fmt.Println(err)
		return nil, err
	}

	db.Close()
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config.Database.Host, config.Database.Port, config.Database.Username, config.Database.Password, config.Database.Dbname)

	db, err = sql.Open(config.Database.Driver, psqlInfo)
	if err != nil || db == nil {
		fmt.Println(err)
		return nil, err
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS messages (id serial PRIMARY KEY, content VARCHAR ( 200 ) NOT NULL, sender VARCHAR ( 200 ) NOT NULL, room VARCHAR (50) NOT NULL, sent_on TIMESTAMP NOT NULL)")

	if err != nil || db == nil {
		fmt.Println(err)
		return nil, err
	}
	return db, err
}
