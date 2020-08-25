package notifier

import (
	"log"
	"net/smtp"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

//EmailAlert - Send email alerts
type EmailAlert struct {
	Body         string
	Subject      string
	ProviderName string
}

//SendViaChannel - For sending email alerts
func (a EmailAlert) SendViaChannel(eCounter chan int) {

	err := godotenv.Load("/home/cloudiff/.env") //Load Environmental Variables
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var (
		gmailAddress  = os.Getenv("GMAIL_ADDRESS")
		gmailPassword = os.Getenv("GMAIL_APP_PASSWORD")
	)

	msg := "From: " + gmailAddress + "\n" +
		"To: " + gmailAddress + "\n" +
		"Subject: Cloudiff " + a.ProviderName + " Alert \n\n" +
		a.Body

	if (a.ProviderName == "Localhost") || (a.ProviderName == "LocalHostNmapResults") { //Don't send localhost alerts
		log.Printf("No changes to SEND")
	} else {
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

//SendIPChangeAlerts - Opens a channel to send IP change alerts
func SendIPChangeAlerts(wg *sync.WaitGroup, alerts chan EmailAlert, aCounter chan int) {
	defer wg.Done()
	for alert := range alerts {
		go alert.SendViaChannel(aCounter)
	}
}

//SendServiceChangeAlerts - Open channel to send service change alerts
func SendServiceChangeAlerts(wg *sync.WaitGroup, alerts chan EmailAlert, aCounter chan int) {
	defer wg.Done()
	for alert := range alerts {
		go alert.SendViaChannel(aCounter)
	}
}

//TrackIPChangeAlerts - numberOfAlerts is equal numberOfProviders
func TrackIPChangeAlerts(numberOfAlerts int, alerts chan EmailAlert, counter chan int) {
	c := 0
	for {
		select {
		case i := <-counter:
			c = c + i
			log.Printf("IP Change Email #%d sent...", c)
			if c == numberOfAlerts {
				close(alerts)
				break
			}
		}
	}
}

//TrackServiceChangeAlerts - numberOfAlerts is equal numberOfProviders
func TrackServiceChangeAlerts(numberOfAlerts int, alerts chan EmailAlert, counter chan int) {
	c := 0
	for {
		select {
		case i := <-counter:
			c = c + i
			log.Printf("Service Change Email #%d sent...", c)
			if c == numberOfAlerts {
				close(alerts)
				break
			}
		}
	}
}
