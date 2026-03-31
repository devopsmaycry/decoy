package services

type SMTP struct{}

func (s SMTP) Banner() []byte {
	return []byte("220 mail.corp.local ESMTP Postfix (Debian/GNU)\r\n")
}
