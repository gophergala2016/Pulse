# Pulse
Log pulse learns from your log files. It uses a machine learning algorithm that Michael Dropps came up with. It is a go package that can be consumed and used for use for anyone. The package itself just reads lines of strings and returns what it thinks is out of place. That way when you are trying to find that error in your logs, you don't spend hours searching and looking. We have made a simple application around it to show case it's ability.

The application is simple. If you run it with no flags or arguments it will read the `PulseConfig.toml` file and read those files listed there. If you include arguments but no flags then the arguments must be filepaths to logs you want to read. EX `LogPulse someFile.log anotherFile.log waitHereIsAnother.log`.

LogPulse accepts one flag `-api`. This is what is running our live demo. It accepts a file on an endpoint in the body and runs the algorithm. It will email the user when it is done with all the anomalies it could find (we are using MailGun). If you wanted to run local you could supply an SMTP config file (location is set in `PulseConfig.toml` and must be a toml file). This is were the credentials are so you are able to send emails locally. You could have the SMTP config file setup and run LogPulse without the `-api` flag and it would send emails as well. If no email option is set it will save all emails (subject and body) to the output file that is specified in the `PulseConfig.toml`

## As a package
To use the algorithm just import the package as such!

`import "github.com/gophergala2016/Pulse/pulse"`

This package exposes the `Run(chan string, func(string))` function. You just need to create a channel that you are going to use. It does require that it is passed in line by line as well. The `func(string)` is a function that is called whenever an unusual string comes by. It is highly recommended that if this is being written to a file to buffer a few strings before you write. Then when you have read all strings dump the rest of the buffer in the file.

## Install
Installing is as simple as:

`go get github.com/gophergala2016/Pulse/LogPulse`

### Pulse Config
The `PulseConfig.toml` needs to be located in the same directory as your executable. The file should look similar to this:
```
LogList = [
"demoData/kern.log.1",
"demoData/kern.log.2"
]

EmailList = [
"someuser@example.org",
"AnneConley@example.org",
"WeAreAwesome@example.org"
]

OutputFile = "PulseOut.txt"
SMTPConfig = "SMTP.toml"

Port = 8080
```
`LogList` is a list of strings. This is where the log files are located that you want pulse to read.


`EmailList` is also a list of strings. But this is everyone that you want to email when something is unusual

`OutputFile` is just a string. It is where the emails are sent if you do not setup an SMTP server (don't have SMTPConf file).

`SMTPConfig` is the location of you SMTP credentials (explained below).

`Port` is the port on which the API server will listen on.

### SMTP Config
The `SMTP.toml` can be anywhere you want it as long as the application can read the file. It is where all the required information is to send email to the SMTP server. It should look like:
```
[Server]
Host = "smtp.server.com"
Port = 25

[User]
UserName = "user@server.com"
PassWord = "LovelyPassword"
```
`[Server]` is a table with `Host` and `Port`
- `Host` is the where the server is listening to receive emails to send.
- `Port` is the port on which the server is listening

`[User]` is also a table but with `UserName` and `PassWord`
- `UserName` is the email address at which the email is sending from.
- `PassWord` is the password for the user that is sending the email

## Team
- Michael Dropps [Github](https://github.com/michaeldropps)
- Miguel Espinoza [Github](https://github.com/miguelespinoza)
- Will Dixon [Github](https://github.com/dixonwille)

## TODO
- [ ] Create Algorithm
  - [ ] DumpPattern
  - [ ] LoadPattern
- [x] Read Config values
  - [x] Files
  - [x] Read a log file on Hard Drive
  - [x] Write to a file to save outstanding logs
- [x] Be able to send emails
  - [x] Use Mailgun's API to send emails
  - [x] Validate User's Email using Mailgun
  - [x] Store keys securely
- [ ] Create the API
  - [ ] POST: log string
  - [ ] POST: file
- [x] Create Webpage
  - [ ] Consume the API
  - [ ] Video of how it works
  - [ ] Static content describing the application
  - [ ] Support links
