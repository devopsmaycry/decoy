package services

type FTP struct {
	banner string
}

func (f FTP) Banner() []byte {
	return []byte(f.banner + "\r\n")
}
