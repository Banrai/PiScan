// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

// Package emailer provides functions for sending outgoing email,
// inspired by https://gist.github.com/rmulley/6603544 and
// https://code.google.com/p/go-wiki/wiki/SendingMail

package emailer

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/smtp"
	"os"
	"text/template"
)

const (
	MAIL_SERVER = "localhost"
	MAIL_PORT   = 25

	LINE_MAX_LEN = 500 // for splitting encoded attachment data

	// templates for generating the message components
	ADDRESS    = "\"{{.DisplayName}}\" <{{.Address}}>"
	HEADERS    = "From: {{.Sender}}\r\nTo: {{.Recipient}}\r\nSubject: {{.Subject}}\r\nMIME-Version: 1.0\r\nContent-Type: multipart/alternative; boundary=\"{{.Boundary}}\"\r\n"
	BODY       = "\r\n--{{.Boundary}}\r\nContent-Type: {{.ContentType}}\r\n\r\n{{.MessageBody}}"
	ATTACHMENT = "\r\n--{{.Boundary}}\r\nContent-Type: {{.ContentType}}; name=\"{{.FileLocation}}\"\r\nContent-Transfer-Encoding:base64\r\nContent-Disposition: attachment; filename=\"{{.FileName}}\"\r\n\r\n{{.EncodedFileData}}"

	// message body mime types
	TEXT_MIME = "text/plain"
	HTML_MIME = "text/html"
)

type EmailAddress struct {
	DisplayName string
	Address     string
}

type EmailHeaders struct {
	Sender    string
	Recipient string
	Subject   string
	Boundary  string
}

type EmailBody struct {
	ContentType string
	MessageBody string
	Boundary    string
}

type EmailAttachment struct {
	ContentType     string
	FileLocation    string
	FileName        string
	EncodedFileData string
	Boundary        string
}

// GenerateBoundary produces a random string that can be used for the email
// multipart boundary marker
func GenerateBoundary() string {
	f, e := os.OpenFile("/dev/urandom", os.O_RDONLY, 0)
	defer f.Close()

	if e != nil {
		return ""
	} else {
		b := make([]byte, 16)
		f.Read(b)
		return fmt.Sprintf("%x", b)
	}
}

func GenerateAddress(context *EmailAddress) (string, error) {
	var doc bytes.Buffer
	t := template.Must(template.New("ADDRESS").Parse(ADDRESS))
	err := t.Execute(&doc, context)
	return doc.String(), err
}

func GenerateHeaders(sender, recipient, subject, boundary string) (string, error) {
	var doc bytes.Buffer
	context := &EmailHeaders{sender, recipient, subject, boundary}
	t := template.Must(template.New("HEADERS").Parse(HEADERS))
	err := t.Execute(&doc, context)
	return doc.String(), err
}

func GenerateBody(message, contentType, boundary string) (string, error) {
	var doc bytes.Buffer
	context := &EmailBody{contentType, message, boundary}
	t := template.Must(template.New("BODY").Parse(BODY))
	err := t.Execute(&doc, context)
	return doc.String(), err
}

func GenerateAttachment(attachment *EmailAttachment) (string, error) {
	var doc, buf bytes.Buffer

	// read and encode the file attachment
	content, contentErr := ioutil.ReadFile(attachment.FileLocation)
	if contentErr != nil {
		return "", contentErr
	}
	encoded := base64.StdEncoding.EncodeToString(content)

	// split the encoded data into individual lines
	// and append them to the byte buffer
	lines := len(encoded) / LINE_MAX_LEN
	for i := 0; i < lines; i++ {
		buf.WriteString(encoded[i*LINE_MAX_LEN:(i+1)*LINE_MAX_LEN] + "\n")
	}
	// don't forget the last line in the buffer
	buf.WriteString(encoded[lines*LINE_MAX_LEN:])
	attachment.EncodedFileData = buf.String()

	// can now process the template
	t := template.Must(template.New("ATTACHMENT").Parse(ATTACHMENT))
	err := t.Execute(&doc, attachment)
	return doc.String(), err
}

// SendFromServer transmits the given message, with optional attachments,
// via the defined mail server and port
func SendFromServer(subject, message, messageType, server string, sender, recipient *EmailAddress, attachments []*EmailAttachment, port int) error {
	var buf bytes.Buffer
	boundary := GenerateBoundary()

	from, fromErr := GenerateAddress(sender)
	if fromErr != nil {
		return fromErr
	}

	to, toErr := GenerateAddress(recipient)
	if toErr != nil {
		return toErr
	}

	hdr, hdrErr := GenerateHeaders(from, to, subject, boundary)
	if hdrErr != nil {
		return hdrErr
	}
	buf.WriteString(hdr)

	body, bodyErr := GenerateBody(message, messageType, boundary)
	if bodyErr != nil {
		return bodyErr
	}
	buf.WriteString(body)

	for _, a := range attachments {
		a.Boundary = boundary
		attach, attachErr := GenerateAttachment(a)
		if attachErr != nil {
			return attachErr
		}
		buf.WriteString(attach)
	}

	// add the closing boundary marker
	buf.WriteString("\r\n--")
	buf.WriteString(boundary)
	buf.WriteString("--")
	fmt.Println(buf.String())

	// connect to the mail server + port
	c, cErr := smtp.Dial(fmt.Sprintf("%s:%d", server, port))
	if cErr != nil {
		return cErr
	}

	// set the sender and recipient (raw email address strings)
	c.Mail(sender.Address)
	c.Rcpt(recipient.Address)

	// stream the full email data
	wc, err := c.Data()
	if err != nil {
		return err
	}
	defer wc.Close()

	_, err = buf.WriteTo(wc)
	return err
}

// Send transmits the given message, with optional attachments, via the
// default mail server (localhost) and port (25)
func Send(subject, message, messageType string, sender, recipient *EmailAddress, attachments []*EmailAttachment) error {
	return SendFromServer(subject, message, messageType, MAIL_SERVER, sender, recipient, attachments, MAIL_PORT)
}
