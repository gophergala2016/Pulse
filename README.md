# Pulse
Log pulse learns from your log files. It uses a machine learning algorithm that Michael Dropps came up with. It is a go package that can be consumed and used for use for anyone that wants to use it. The package itself just reads lines of strings and returns what it thinks is out of place. That way when you are trying to find that error in your logs, you don't spend hours searching and looking. We have made a simple application around it to show case it's ability.

The application is simple. If you run it with no commands it will listen on whatever port is specified in the `PulseConfig.toml` file. It is listening for any log server willing to give us log lines that are then passed to the algorithm. You can setup an SMTP server to be able to let our application email you if we find anything out of the unusual. But if you use the `-d` flag, for default, it will read what ever log files that already exist and are mapped to using the `PulseConfig.toml`.

Not enough what if you don't want to edit the config file every time. Then just pass in file names as arguments like so `LogPulse somefile.log yetAnotherLog.log`. It will read from both and let you know something is up.

You don't have an SMTP server? Log Pulse will output all unusual entries into an output file that you specified in the config.

## Install
Installing is as simple as:

`go get http://github.com/gophergala2016/Pulse/pulse/cmd/pulse`

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

## As a package
To use the algorithm just import the package as such!

`import "github.com/gophergala2016/Pulse/pulse"`

This package exposes the `Run(chan string, func(string))` function. You just need to create a channel that you are going to use. It does require that it is passed in line by line as well. The `func(string)` is a function that is called whenever an unusual string comes by. It is highly recommended that if this is being written to a file to buffer a few strings before you write. Then when you have read all strings dump the rest of the buffer in the file.
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
