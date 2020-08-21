package notifier

import (
	"log"
	"net/smtp"
	"os"
)

var (
	gmailAddress  = os.Getenv("GMAIL_ADDRESS")
	gmailPassword = os.Getenv("GMAIL_APP_PASSWORD")
)

//Send email alerts
func Send(body string, providerName string) {

	msg := "From: " + gmailAddress + "\n" +
		"To: " + gmailAddress + "\n" +
		"Subject: Cloudiff " + providerName + " Changes \n\n" +
		body

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", gmailAddress, gmailPassword, "smtp.gmail.com"),
		gmailAddress, []string{gmailAddress}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}

	log.Printf("SENT %s changes", providerName)
}
