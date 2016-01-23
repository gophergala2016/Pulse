package email

import (
	"encoding/json"
	"fmt"
	"net/smtp"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/mailgun/mailgun-go"
)

const (
	mailGunSend = iota
	smtpSend    = iota
	jsonSend    = iota
)

// MailGun ...
type MailGun struct {
	Sender string `toml:"Sender"`
	Domain string `toml:"Domain"`
	Secret string `toml:"Secret"`
	Public string `toml:"Public"`
}

// JSONAlert ...
type JSONAlert struct {
	Message string `json:"message"`
	Body    string `json:"body"`
}

var mGun *MailGun

var (
	mailGunConfig = "MailGun.toml"
	emailOption   = -1
)

/* initialize email service used for notifications
1. MailGun
2. SMTP package
3. Send to JSON
*/
func init() {
	var err error
	mGun, err = LoadConfig(mailGunConfig)
	if err != nil {
		// Check smtp server
	}
	emailOption = mailGunSend
}

// Send : sends email via MailGun, smtp server, or simply a JSON file
func Send(message string) {
	switch emailOption {
	case mailGunSend:
		fireMailGun(message)
	case smtpSend:
		fireSMTPMessage(message)
	case jsonSend:
	}
}

// fireMailGun : uses MailGun API: thanks! for your service :)
func fireMailGun(body string) {
	gun := mailgun.NewMailgun(mGun.Domain, mGun.Secret, mGun.Public)

	email := "recipient@example.com"
	// for _, val := range .. { // Get Addresses from PulseConfig
	check, _ := gun.ValidateEmail(email)
	if check.IsValid {
		m := mailgun.NewMessage(
			fmt.Sprintf("Sender <%s>", mGun.Sender),
			"Alert! Found Anomaly in Log Files via LogPulse",
			body,
			fmt.Sprintf("Recipient <%s>", email))
		response, id, _ := gun.Send(m)
		fmt.Printf("Response ID: %s\n", id)
		fmt.Printf("Message from server: %s\n", response)
	}

	// }

}

// fireSMTPMessage : uses smtp client to fire an email based on config file settings
func fireSMTPMessage(body string) {
	auth := smtp.PlainAuth(
		"",
		"user@example.com",
		"password",
		"mail.example.com",
	)

	email := "recipient@example.com"
	// for _, val := range .. { // Get Addresses from PulseConfig

	to := []string{email}
	msg := []byte("To: " + email + ":\r\n" +
		"Subject: Alert! Found Anomaly in Log Files via LogPulse\r\n" +
		"\r\n" +
		body + "\r\n")

	err := smtp.SendMail(
		"mail.example.com:25",
		auth,
		"sender@example.org",
		to,
		msg,
	)
	if err != nil {
		fmt.Printf("Failed to send to %s\n", email)
	}
	// }
}

// fireJSONOutput : when all else fails... output body to JSON
func fireJSONOutput(body string) {
	output := JSONAlert{"Alert! Found Anomaly in Log Files via LogPulse", body}
	val, err := json.Marshal(output)
	if err != nil {
		fmt.Println("Failed to create JSON Alert")
		return
	}
	var filename = "./log/alert.json"
	var f *os.File
	f, err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		f, err = os.Create(filename)
		if err != nil {
			fmt.Println("Failed to create json alert file")
			return
		}
	}
	defer f.Close()

	if _, err = f.WriteString(string(val)); err != nil {
		fmt.Println("Failed to write json alert to file")
		return
	}
}

// LoadConfig ...
func LoadConfig(filename string) (*MailGun, error) {
	cfg := &MailGun{}
	if _, err := toml.DecodeFile(filename, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
