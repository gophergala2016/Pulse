package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

//Configuration is the main configurations for the application
type Configuration struct {
}

//SMTPConfig is the configurations for a personal SMTP server a user would like to use
type SMTPConfig struct {
	Server Server
	User   User
}

//Server is the SMTP Server
type Server struct {
	Host string `toml:"Host"`
	Port int    `toml:"Port"`
}

//User is the User/Pass combination for the SMTP Server
type User struct {
	UserName string `toml:"UserName"`
	PassWord string `toml:"PassWord"`
}

//SecretConfig is the configurations to hold the keys for MailGun
type SecretConfig struct {
}

//Load returns the main configuration
func Load(filename string) (*Configuration, error) {
	return nil, fmt.Errorf("Something went wrong")
}

//LoadSMTP loads the settings for the smtp server
func LoadSMTP(filename string) (*SMTPConfig, error) {
	cfg := &SMTPConfig{}
	if _, err := toml.DecodeFile(filename, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

//LoadSecret loads the keys for Mailgun
func LoadSecret(filename string) (*SecretConfig, error) {
	return nil, fmt.Errorf("Something went wrong")
}
