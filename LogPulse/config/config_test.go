package config_test

import (
	"testing"

	. "github.com/gophergala2016/Pulse/LogPulse/config"
)

func TestLoad(t *testing.T) {
	expectedCfg := Configuration{}
	expectedCfg.LogList = []string{"demoData/kern.log.1", "demoData/kern.log.2"}
	expectedCfg.EmailList = []string{
		"someuser@example.org",
		"AnneConley@example.org",
		"WeAreAwesome@example.org",
		"espinozamiguel349@gmail.com",
	}
	expectedCfg.OutputFile = "PulseOut.txt"
	expectedCfg.SMTPConfig = "s.toml"
	expectedCfg.Port = 8080
	cfg, err := Load()
	if err != nil {
		t.Errorf("Could not load config. %s", err)
	}

	if len(expectedCfg.LogList) != len(cfg.LogList) {
		t.Errorf("Loglist lengths are wrong")
	}
	for i, l := range expectedCfg.LogList {
		if cfg.LogList[i] != l {
			t.Errorf("Loglist does not match")
		}
	}

	if len(expectedCfg.EmailList) != len(cfg.EmailList) {
		t.Errorf("Emaillist lengths are wrong")
	}
	for i, l := range expectedCfg.EmailList {
		if cfg.EmailList[i] != l {
			t.Errorf("Emaillist does not match")
		}
	}

	if expectedCfg.OutputFile != cfg.OutputFile {
		t.Errorf("Output file does not match")
	}
	if expectedCfg.SMTPConfig != cfg.SMTPConfig {
		t.Errorf("SMTP File does not match")
	}
	if expectedCfg.Port != cfg.Port {
		t.Errorf("Prot numbers does not match")
	}
}

func TestLoadSMTP(t *testing.T) {
	expectedCfg := SMTPConfig{}
	expectedCfg.Server.Host = "smtp.mailgun.org"
	expectedCfg.Server.Port = 25
	expectedCfg.User.PassWord = "Password"
	expectedCfg.User.UserName = "postmaster@clemsonopoly.com"
	cfg, err := LoadSMTP()
	if err != nil {
		t.Errorf("Could not load SMTP file")
	}
	if expectedCfg.Server.Host != cfg.Server.Host {
		t.Errorf("Host does not match")
	}
	if expectedCfg.Server.Port != cfg.Server.Port {
		t.Errorf("Prot numbers does not match")
	}
	if expectedCfg.User.UserName != cfg.User.UserName {
		t.Errorf("Username does not match")
	}
	if expectedCfg.User.PassWord != cfg.User.PassWord {
		t.Errorf("Password does not match")
	}
}

func TestLoadSecret(t *testing.T) {
	expectedCfg := SecretConfig{}
	expectedCfg.Domain = "clemsonopoly.com"
	expectedCfg.PrivateKey = "SECRET"
	expectedCfg.PublicKey = "PUBLIC"
	expectedCfg.Sender = "postmaster@clemsonopoly.com"
	cfg, err := LoadSecret()
	if err != nil {
		t.Errorf("Could not load secret file")
	}
	if expectedCfg.Domain != cfg.Domain {
		t.Errorf("Domain does not match")
	}
	if expectedCfg.PrivateKey != cfg.PrivateKey {
		t.Errorf("Privatekey does not match")
	}
	if expectedCfg.PublicKey != cfg.PublicKey {
		t.Errorf("Publickey does not match")
	}
	if expectedCfg.Sender != cfg.Sender {
		t.Errorf("Sender does not match")
	}
}
