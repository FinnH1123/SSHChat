package server

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/FinnH1123/SSHChat/config"
	"github.com/charmbracelet/bubbles/list"
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
	err := db.QueryRow(sqlStatement, message.content, message.sender, roomid, time.Now()).Scan(&msg.Id, &msg.Content, &msg.RoomId, &msg.Sender)
	return msg, err
}

func GetRoomMessages(db *sql.DB, roomid string) ([]list.Item, error) {
	var msgs []list.Item
	sqlStatement := `
	Select * FROM messages 
	WHERE room = $1
	ORDER BY sent_on ASC
	`

	rows, err := db.Query(sqlStatement, roomid)
	msg := DBMessage{}
	defer rows.Close()
	for rows.Next() { // creating new struct for every row
		err = rows.Scan(&msg.Id, &msg.Content, msg.Sender, msg.RoomId, msg.Sent)
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
	conninfo := "user=postgres password=yourpassword host=127.0.0.1 sslmode=disable"
	db, err := sql.Open("postgres", conninfo)

	if err != nil {
		//handle the error
		return nil, err
	}

	_, err = db.Exec("create database IF NOT EXISTS " + config.Database.Dbname)
	if err != nil {
		//handle the error
		return nil, err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS messages (id serial PRIMARY KEY, username VARCHAR ( 200 ) UNIQUE NOT NULL, password VARCHAR ( 200 ) NOT NULL)")

	if err != nil {
		//handle the error
		return nil, err
	}
	db.Close()
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config.Database.Host, config.Database.Port, config.Database.Username, config.Database.Password, config.Database.Dbname)

	db, err = sql.Open(config.Database.Driver, psqlInfo)
	if err != nil {
		//handle the error
		return nil, err
	}
	return db, err
}
