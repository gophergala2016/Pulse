package email

import (
	"Pulse/pulse/config"
	"encoding/json"
	"fmt"
	"net/smtp"
	"os"
	"strconv"

	"github.com/mailgun/mailgun-go"
)

const (
	mailGunSend = iota
	smtpSend    = iota
	jsonSend    = iota
)

// JSONAlert ...
type JSONAlert struct {
	Message string `json:"message"`
	Body    string `json:"body"`
}

var mGun *config.SecretConfig
var smtpConfig *config.SMTPConfig
var emailList []string

var (
	mailGunConfig   = "secret.toml"
	emailOption     = -1
	pulseConfigFail = false
)

/* initialize email service used for notifications
1. MailGun
2. SMTP package
3. Send to JSON
*/
func init() {
	val, err := config.Load()
	if err != nil {
		fmt.Println("Failed  to load Main config file")
		pulseConfigFail = true
	}

	if !pulseConfigFail {
		emailList = val.EmailList
	}

	mGun, err = config.LoadSecret()
	if err != nil {
		// Check smtp server
		smtpConfig, err = config.LoadSMTP()
		if err != nil {
			// Use JSON
			emailOption = jsonSend
			return
		}
		emailOption = smtpSend
		return
	}
	emailOption = mailGunSend
}

// Send : sends email via MailGun, smtp server, or simply a JSON file
func Send(message string) {
	switch emailOption {
	case mailGunSend:
		if pulseConfigFail {
			fmt.Println("MailGun service is dependent of PulseConfig")
			return
		}
		fireMailGun(message)
	case smtpSend:
		if pulseConfigFail {
			fmt.Println("SMTP client is dependent of PulseConfig")
			return
		}
		fireSMTPMessage(message)
	case jsonSend:
		fireJSONOutput(message)
	}
}

// fireMailGun : uses MailGun API: thanks! for your service :)
func fireMailGun(body string) {
	gun := mailgun.NewMailgun(mGun.Domain, mGun.PrivateKey, mGun.PublicKey)

	for _, email := range emailList { // Get Addresses from PulseConfig
		check, _ := gun.ValidateEmail(email)
		if check.IsValid {
			m := mailgun.NewMessage(
				fmt.Sprintf("LogPulse <%s>", mGun.Sender),
				"Alert! Found Anomaly in Log Files via LogPulse",
				body,
				fmt.Sprintf("Recipient <%s>", email))

			response, id, _ := gun.Send(m)
			// TODO: for testing purpose will change later, maybe just fire goroutine
			fmt.Printf("Response ID: %s\n", id)
			fmt.Printf("Message from server: %s\n", response)
		}

	}

}

// fireSMTPMessage : uses smtp client to fire an email based on config file settings
func fireSMTPMessage(body string) {

	auth := smtp.PlainAuth(
		"",
		smtpConfig.User.UserName,
		smtpConfig.User.PassWord,
		smtpConfig.Server.Host,
	)

	for _, email := range emailList { // Get Addresses from PulseConfig

		to := []string{email}
		msg := []byte("To: " + email + ":\r\n" +
			"Subject: Alert! Found Anomaly in Log Files via LogPulse\r\n" +
			"\r\n" +
			body + "\r\n")

		err := smtp.SendMail(
			fmt.Sprintf("%s:%s", smtpConfig.Server.Host, strconv.Itoa(smtpConfig.Server.Port)),
			auth,
			"sender@example.org",
			to,
			msg,
		)
		if err != nil {
			fmt.Printf("Failed to send to %s\n", email)
		}
	}
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
	newLine := true
	var f *os.File
	f, err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		f, err = os.Create(filename)
		if err != nil {
			fmt.Println("Failed to create json alert file")
			return
		}
		newLine = false
	}
	defer f.Close()

	if newLine {
		val = []byte(string(val) + "\n")
	}
	if _, err = f.WriteString(string(val)); err != nil {
		fmt.Println("Failed to write json alert to file")
		return
	}
}
