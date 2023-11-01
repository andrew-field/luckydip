package main

import (
	"bytes"
	"encoding/base64"
	"net/smtp"
	"os"
)

func sendEmail(to, subject, body string, pic []byte) {
	from := "andrewpcfield@gmail.com"

	// Set up authentication information.
	auth := smtp.PlainAuth("", from, os.Getenv("GOOGLEAPPPASSWORD"), "smtp.gmail.com")

	// Message
	msg := bytes.NewBuffer(nil)
	msg.WriteString("From: " + from + "\n")
	msg.WriteString("To: " + to + "\n")
	msg.WriteString("Subject: " + subject + "\n")

	if pic != nil { // There was an error
		msg.WriteString("MIME-Version: 1.0\n")
		msg.WriteString(`Content-Type: multipart/related; boundary="myboundary"` + "\n\n")
		msg.WriteString("--myboundary\n")

		// This is the body
		msg.WriteString(`Content-Type: text/plain; charset="utf-8"` + "\n")
		msg.WriteString("Content-Transfer-Encoding: quoted-printable" + "\n\n")
		msg.WriteString(body + "\n\n")
		msg.WriteString("--myboundary\n")

		// This is the attachment
		encodedImage := base64.StdEncoding.EncodeToString(pic)
		msg.WriteString(`Content-Type: image/jpeg;name="image.jpg"` + "\n")
		msg.WriteString("Content-Transfer-Encoding: base64" + "\n")
		msg.WriteString("Content-Disposition: attachment;filename=\"image.jpg\"" + "\n\n")
		msg.WriteString(encodedImage + "\n\n")
		msg.WriteString("--myboundary--")
	} else {
		msg.WriteString(body + "\n\n")
	}

	// Send the message.
	err := smtp.SendMail("smtp.gmail.com:587", auth, from, []string{to}, msg.Bytes())
	if err != nil {
		panic(err)
	}
}
