// Package config reads config files and returns the proper config structure.
package config

import (
	"fmt"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/mitchellh/go-homedir"
)

// Configuration is the main configurations for the application.
type Configuration struct {
	// LogList is a list of log file locations that should be read if no arguments are uncompressed.
	LogList []string `toml:"LogList"`

	// EmailList is the list of recipients that will get an email when algorithm is done.
	EmailList []string `toml:"EmailList"`

	// OutputFile is the location of a file to output the emails if an SMTP server is not present.
	OutputFile string `toml:"OutputFile"`

	// SMTPConfig is the locationn of the SMTP config file with credentials in it.
	SMTPConfig string `toml:"SMTPConfig"`

	// Port is the port at which the API is to listen on.
	Port int `toml:"Port"`
}

// SMTPConfig is the configurations for a personal SMTP server a user would like to use.
type SMTPConfig struct {
	// Server has the information about the where the SMTP server is hosted and what port it is listening on.
	Server Server

	// User is the person who is going to be the person who is sending the emails.
	User User
}

// Server is the SMTP Server credentials.
type Server struct {
	// Host is where the SMTP server is hosted.
	Host string `toml:"Host"`

	// Port is the Prot on which the SMTP server is listening on.
	Port int `toml:"Port"`
}

// User has the credentials for the person who is sending the email.
type User struct {
	// UserName is the username of the person sending the email.
	UserName string `toml:"UserName"`

	// PassWord is the password of the user.
	PassWord string `toml:"PassWord"`
}

// SecretConfig is the configurations to hold the keys for MailGun.
type SecretConfig struct {
	// Sender is the user who is sending the email.
	Sender string `toml:"Sender"`

	// Domain is the domain name of which we want to use.
	Domain string `toml:"Domain"`

	// PrivateKey is the private key to access MailGun's API.
	PrivateKey string `toml:"PrivateKey"`

	// PublicKey is the public key to access MailGun's API.
	PublicKey string `toml:"PublicKey"`
}

var (
	mailGunConfig = "MailGun.toml"
	pulseConfig   = "PulseConfig.toml"
	smtpConfig    string
)

//Load returns the main configuration file.
func Load() (*Configuration, error) {
	cfg := &Configuration{}
	// Search in the same directory as the binary first.
	if _, err := toml.DecodeFile(pulseConfig, cfg); err != nil {
		// If we couldn't find it ther keep looking.

		// Find the home directory for user.
		home, err := homedir.Dir()
		if err != nil {
			return nil, fmt.Errorf("config.Load: Could not find %s in the executable directory and could not find home directory", pulseConfig)
		}
		// Look in the home directory of the user for the main config.
		if _, err := toml.DecodeFile(filepath.Join(home, pulseConfig), cfg); err != nil {
			return nil, fmt.Errorf("config.Load: Could not find %s in the %s or executable directory", pulseConfig, home)
		}
	}
	return cfg, nil
}

//LoadSMTP loads the settings for the smtp server.
func LoadSMTP() (*SMTPConfig, error) {
	//SMTP file location is in the main config.

	// Try and load it. If we can't return an error
	maincfg, err := Load()
	if err != nil {
		return nil, fmt.Errorf("config.LoadSMTP: %s", err)
	}

	// Load the SMTP config and return if we can.
	cfg := &SMTPConfig{}
	if _, err := toml.DecodeFile(maincfg.SMTPConfig, cfg); err != nil {
		return nil, fmt.Errorf("config.LoadSMTP: %s", err)
	}
	return cfg, nil
}

//LoadSecret loads the keys for Mailgun.
func LoadSecret() (*SecretConfig, error) {
	//Only search in directory of binary since we are the only ones with access to our MailGun client.
	cfg := &SecretConfig{}
	if _, err := toml.DecodeFile(mailGunConfig, cfg); err != nil {
		return nil, fmt.Errorf("config.LoadSecret: %s", err)
	}
	return cfg, nil
}
