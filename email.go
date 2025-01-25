package luckydip

import (
	"bytes"
	"log"
	"os"

	"gopkg.in/mail.v2"
)

func sendEmail(to, subject, body string, pic []byte) {
	m := mail.NewMessage()

	// Set the email sender and recipient.
	m.SetHeader("From", "andrewpcfield@gmail.com")
	m.SetHeader("To", to)

	// Set the email subject.
	m.SetHeader("Subject", subject)

	// Set the email body.
	m.SetBody("text/plain", body)

	// Add the attachment if pic is provided. This means there was an error.
	if pic != nil {
		m.AttachReader("Error picture.jpg", bytes.NewReader(pic))
	}

	// Set up authentication information.
	password := os.Getenv("GOOGLEAPPPASSWORD")
	if password == "" {
		log.Fatal("GOOGLEAPPPASSWORD environment variable is not set")
	}

	// Create a new SMTP dialer.
	d := mail.NewDialer("smtp.gmail.com", 587, "andrewpcfield@gmail.com", password)

	// Send the email.
	if err := d.DialAndSend(m); err != nil {
		log.Printf("Error sending email: %v", err)
	} else {
		log.Println("Email sent successfully!")
	}
}
