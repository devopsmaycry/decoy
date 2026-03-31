package services

type FTP struct{}

func (f FTP) Banner() []byte {
	return []byte("220 Microsoft FTP Service\r\n")
}
