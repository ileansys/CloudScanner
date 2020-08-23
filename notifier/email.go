package notifier

import (
	"log"
	"net/smtp"
	"os"
)

//EmailAlert - Send email alerts
type EmailAlert struct {
	Body         string
	ProviderName string
}

var (
	gmailAddress  = os.Getenv("GMAIL_ADDRESS")
	gmailPassword = os.Getenv("GMAIL_APP_PASSWORD")
)

//SendViaChannel - For sending email alerts
func (a EmailAlert) SendViaChannel(eCounter chan int) {

	msg := "From: " + gmailAddress + "\n" +
		"To: " + gmailAddress + "\n" +
		"Subject: Cloudiff " + a.ProviderName + " Alert \n\n" +
		a.Body

	if a.ProviderName != "Localhost" { //Don't send localhost alerts
		err := smtp.SendMail("smtp.gmail.com:587",
			smtp.PlainAuth("", gmailAddress, gmailPassword, "smtp.gmail.com"),
			gmailAddress, []string{gmailAddress}, []byte(msg))
		if err != nil {
			log.Printf("smtp error: %s", err)
			return
		}

		log.Printf("SENT %s changes", a.ProviderName)
	}

	eCounter <- 1
}

//Send - For sending email alerts
func (a EmailAlert) Send() {

	msg := "From: " + gmailAddress + "\n" +
		"To: " + gmailAddress + "\n" +
		"Subject: Cloudiff " + a.ProviderName + " Alert \n\n" +
		a.Body

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", gmailAddress, gmailPassword, "smtp.gmail.com"),
		gmailAddress, []string{gmailAddress}, []byte(msg))
	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}

	log.Printf("SENT %s changes", a.ProviderName)

}

//TrackEmailAlerts - numberOfAlerts is equal numberOfProviders
func TrackEmailAlerts(numberOfAlerts int, alerts chan EmailAlert, counter chan int) {
	c := 0
	for {
		select {
		case i := <-counter:
			c = c + i
			log.Printf("Email #%d sent...", c)
			if c == numberOfAlerts {
				close(alerts)
				break
			}
		}
	}
}
