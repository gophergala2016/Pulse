// Package email will try and send an email using MailGun.
// If we don't have the config for MailGun use the SMTP confg.
// If we don't have that either save to the output file specified in config
package email

import (
	"encoding/json"
	"fmt"
	"net/smtp"
	"os"
	"strconv"

	"github.com/gophergala2016/Pulse/LogPulse/config"
	"github.com/gophergala2016/Pulse/LogPulse/file"
	"github.com/mailgun/mailgun-go"
)

const (
	mailGunSend = iota
	smtpSend    = iota
	jsonSend    = iota
)

// JSONAlert holds the message and body to send through email.
type JSONAlert struct {
	Message string `json:"message"`
	Body    string `json:"body"`
}

var (
	emailOption = -1

	// ByPassMail is Whether or not we are using the email system.
	ByPassMail = false
	mGun       *config.SecretConfig
	smtpConfig *config.SMTPConfig

	// EmailList is a list of emails to send messages to.
	EmailList []string

	// OutputFile is the output file specified in the main config
	OutputFile   string
	stringBuffer []string
)

// initialize email service used for notifications
// 1. MailGun
// 2. SMTP package
// 3. Send to JSON
func init() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			os.Exit(0)
		}
	}()

	if ByPassMail {
		emailOption = jsonSend
		return
	}
	val, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("email.init: Failed to load Main config file"))
	}

	// Get values from the main config.
	EmailList = val.EmailList
	OutputFile = val.OutputFile

	mGun, err = config.LoadSecret()
	if err != nil {
		// Check smtp server
		smtpConfig, err = config.LoadSMTP()
		if err != nil {
			// Use JSON
			emailOption = jsonSend
			return
		}
		// Use SMTP
		emailOption = smtpSend
		return
	}
	// Use MailGun
	emailOption = mailGunSend
}

// SendFromCache sends email via MailGun, smtp server, or simply a JSON file but loads body from cache file.
// Filename is the location of the cache file
func SendFromCache(filename string) {
	fmt.Println("email.SendFromCache: Sending from Cache")
	var body string

	line := make(chan string)
	file.Read(filename, line)
	for l := range line {
		body += l + "\n"
	}

	Send(body)
}

// Send sends email via MailGun, smtp server, or simply a JSON file.
func Send(message string) {
	fmt.Println("email.Send: Sending")
	switch emailOption {
	case mailGunSend:
		go fireMailGun(message)
	case smtpSend:
		go fireSMTPMessage(message)
	case jsonSend:
		fireJSONOutput(message) // We want lines sent in saved in the order they were sent in.
	}
}

//SaveToCache takes a string and saves it to file.
func SaveToCache(message string) {
	fireJSONOutput(message) // We want lines sent in saved in the order they were sent in.
}

// IsValid checks to see if the email that is passed in is a valid email or not.
func IsValid(email string) bool {
	gun := mailgun.NewMailgun(mGun.Domain, mGun.PrivateKey, mGun.PublicKey)

	check, _ := gun.ValidateEmail(email)
	return check.IsValid
}

// fireMailGun uses MailGun API: thanks! for your service :)
func fireMailGun(body string) {
	gun := mailgun.NewMailgun(mGun.Domain, mGun.PrivateKey, mGun.PublicKey)

	for _, email := range EmailList { // Get Addresses from PulseConfig
		if IsValid(email) {
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

// fireSMTPMessage uses smtp client to fire an email based on config file settings.
func fireSMTPMessage(body string) {

	auth := smtp.PlainAuth(
		"", // identity left blank because it will use UserName instead
		smtpConfig.User.UserName,
		smtpConfig.User.PassWord,
		smtpConfig.Server.Host,
	)

	for _, email := range EmailList { // Get Addresses from PulseConfig

		to := []string{email}
		msg := []byte("To: " + email + ":\r\n" +
			"Subject: Alert! Found Anomaly in Log Files via LogPulse\r\n" +
			"\r\n" +
			body + "\r\n")

		err := smtp.SendMail(
			fmt.Sprintf("%s:%s", smtpConfig.Server.Host, strconv.Itoa(smtpConfig.Server.Port)),
			auth,
			smtpConfig.User.UserName,
			to,
			msg,
		)
		if err != nil {
			fmt.Printf("fireSMTPMessage: Failed to send to %s\n", email)
		}
	}
}

// fireJSONOutput when all else fails... output body to JSON
// Also used by chaching system.
func fireJSONOutput(body string) {

	output := JSONAlert{"Alert! Found Anomaly in Log Files via LogPulse", body}
	val, err := json.Marshal(output)
	if err != nil {
		fmt.Println("email.fireJSONOutput: Failed to create JSON Alert")
		return
	}

	// Create a buffer of strings so we are not constantly opening and closing the file
	file.Write(OutputFile, string(val))
}
