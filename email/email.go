package email

import (
	"html/template"
	"log"
	"net/smtp"
	"os"
	"strings"
)

type EmailPayload struct {
	From     string
	To       string
	Password string
	Code     string
	Message  string
}

func SendVerificationCode(params EmailPayload) (string, error) {
	htmlFile, err := os.ReadFile("./email/format.html")
	if err != nil {
		log.Println("Cannot read from file", htmlFile)
		return "", err
	}

	temp, err := template.New("email").Parse(string(htmlFile))
	if err != nil {
		log.Println("Cannot parse file", err)
		return "", err
	}

	var builder strings.Builder
	err = temp.Execute(&builder, params)
	if err != nil {
		log.Println("Cannot execute file", err)
		return "", err
	}

	message := "From: " + params.From + "\n" +
		"To: " + params.To + "\n" +
		"Subject: Super-clinic app\n" +
		"MIME-Version: 1.0\n" +
		"Content-type: text/html; charset=\"UTF-8\"\n" +
		"\n" +
		builder.String()

	err = smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("smtp", params.From, params.Password, "smtp.gmail.com"),
		params.From, []string{params.To}, []byte(message),
	)

	if err != nil {
		log.Println("cannot send an email verification", err)
		return "", err
	}

	return "a verification code was sent to your email, please check it", nil
}
