package main

import (
	"fmt"
	"net/mail"
	"net/smtp"
	"sync"
	"time"

	"database/sql"

	_ "github.com/lib/pq"
	"github.com/scorredoira/email"
)

const (
	host     = "db"
	port     = 5432
	user     = "postgres"
	password = "1234"
	dbname   = "dbname"
)

type emailContext struct {
	id       int
	subject  string
	body     string
	receiver string
}

func main() {
	fmt.Println("\n\n\n************************")
	currentTime := time.Now()
	fmt.Println("LOG DATE : ", currentTime.Format("01-02-2006 15:04:05"))

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nSuccessfully connected to %s database!\n", dbname)
	defer db.Close()

	sendEmails(db)
	fmt.Println("\n************************")
	return
}

func sendEmails(db *sql.DB) {
	var id int
	var message string
	var subject string
	var body string
	var receiver string
	var EmailData []emailContext

	ResultsRow, err := db.Query("select toemail_table.id,toemail_table.subject,toemail_table.body,toemail_table.receiver where toemail_table.is_sent=FALSE or toemail_table.is_sent=NULL;")
	if err != nil {
		fmt.Printf("Query failed %s", err)
	}
	for ResultsRow.Next() {
		ResultsRow.Scan(&id, &subject, &body, &receiver)
		fmt.Printf("Id:%d , subject: %s , body: %s , receiver: %s\n", id, subject, body, receiver)

		EmailData = append(EmailData, emailContext{
			id:       id,
			subject:  subject,
			body:     message,
			receiver: receiver,
		})
	}

	var wg sync.WaitGroup
	wg.Add(len(EmailData))
	for _, emaildata := range EmailData {
		go func(emaildata emailContext) {
			defer wg.Done()

			m := email.NewHTMLMessage("New Email", emaildata.body)
			m.From = mail.Address{Name: "Name", Address: "Address"}
			m.To = []string{emaildata.receiver}
			m.AddHeader("key", "value")

			// send mail
			auth := smtp.PlainAuth("identity", "username", "password", "host")
			if err := email.Send("addr", auth, m); err != nil {
				fmt.Printf("SMTP Error: %s\n", err)
				return
			}
			updateQuery := "update toemail_table set is_sent=True where id = $1"
			_, err = db.Exec(updateQuery, emaildata.id)
			if err != nil {
				fmt.Printf("Update Query Error : %s\n", err)
				return
			}
			fmt.Println("Email is sent!")
		}(emaildata)
	}
	wg.Wait()
}
