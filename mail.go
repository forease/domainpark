package main

import (
    "bytes"
    "errors"
    "fmt"
    "net/mail"
    "net/smtp"
    "time"
    "strings"
    "log"
)

type Mailer struct {
    Subject string
    Content string
    To string
}

var mailChan = make(chan *Mailer)

// 运行一个监听发送邮件任务
func MailSendServer() {
//    mailChan = make(chan *Mailer)
    log.Println( "Running Mail Send Server..." )

    for {

        mail := <-mailChan

        if mail == nil {
            continue
        }

        m := NewMailMessage( mail.Subject, mail.Content, mail.To )
        if err := m.Send(); err != nil {
            log.Println( "send mail to " + mail.To + " error %s", err )
        }
    }

}

func MailSender( subject, content, to string ) ( err error ) {

    if subject != "" && content != "" && to != "" {
        if AppConfig.smtpDaemon {
            mailChan <- &Mailer{ Subject: subject, Content: content, To: to }
        } else {
            m := NewMailMessage( subject, content, to )
            if err = m.Send(); err != nil {
                log.Println( "send mail to " + to + " error %s", err )
                return err
            }
        }

        return nil
    }

    return errors.New("input is null")
}


func NewMailMessage(subject, content, to string) *MailMessage {
    tos := strings.Split( to, "," )
    message := &MailMessage{Subject: subject, Content: content,
        To: make([]mail.Address, len(tos))}

    for k, v := range tos {
        message.To[k].Address = v
    }

    //fmt.Println( message.To )
    return message
}

func NewMailMessageFrom(subject, content, from, to string) *MailMessage {
    message := NewMailMessage(subject, content, to)
    message.From.Address = from
    return message
}

const crlf = "\r\n"

type MailMessage struct {
    From    mail.Address // if From.Address is empty, Config.DefaultFrom will be used
    To      []mail.Address
    Cc      []mail.Address
    Bcc     []mail.Address
    Subject string
    Content string
}

// http://tools.ietf.org/html/rfc822
// http://tools.ietf.org/html/rfc2821
func (self *MailMessage) String() string {
    var buf bytes.Buffer

    write := func(what string, recipients []mail.Address) {
        if len(recipients) == 0 {
            return
        }
        for i := range recipients {
            if i == 0 {
                buf.WriteString(what)
            } else {
                buf.WriteString(", ")
            }
            buf.WriteString(recipients[i].String())
        }
        buf.WriteString(crlf)
    }

    from := &self.From
    if from.Address == "" {
        from = &mail.Address{ "domainpark", AppConfig.smtpUser }//&Config.From
    }

    /*if AppConfig.adminMail != "" {
        self.Bcc = make([]mail.Address, 1)
        self.Bcc[0] = mail.Address{ "adminer", AppConfig.adminMail }
    }*/


    fmt.Fprintf(&buf, "From: %s%s", from.String(), crlf)
    write("To: ", self.To)
    write("Cc: ", self.Cc)
    write("Bcc: ", self.Bcc)
    fmt.Fprintf(&buf, "Date: %s%s", time.Now().UTC().Format(time.RFC822), crlf)
    fmt.Fprintf(&buf, "Subject: %s%s%s", self.Subject, crlf, self.Content)
    return buf.String()
}

// Returns the first error
func (self *MailMessage) Validate() error {
    if len(self.To) == 0 {
        return errors.New("Missing email recipient (email.Message.To)")
    }
    return nil
}

type fakeAuth struct {
    smtp.Auth
}

func (a fakeAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
    server.TLS = true
    return a.Auth.Start(server)
}


func (self *MailMessage) Send() error {
    var auth smtp.Auth

    if err := self.Validate(); err != nil {
        return err
    }

    to := make([]string, len(self.To))
    for i := range self.To {
        to[i] = self.To[i].Address
    }

    from := self.From.Address
    if from == "" {
        from = AppConfig.smtpUser// Config.From.Address
    }

    addr := fmt.Sprintf("%s:%d", AppConfig.smtpHost, AppConfig.smtpPort)

    if AppConfig.smtpTLS {
        auth = fakeAuth{smtp.PlainAuth("", AppConfig.smtpUser,
            AppConfig.smtpPassword, AppConfig.smtpHost)}
    } else {
        auth = smtp.PlainAuth("", AppConfig.smtpUser, AppConfig.smtpPassword,
            AppConfig.smtpHost)
    }

    return smtp.SendMail(addr, auth, from, to, []byte(self.String()))
}
