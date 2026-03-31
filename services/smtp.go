package services

type SMTP struct {
	banner string
}

func (s SMTP) Banner() []byte {
	return []byte(s.banner + "\r\n")
}
