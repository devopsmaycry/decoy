package services

type Redis struct {
	banner string
}

func (r Redis) Banner() []byte {
	return []byte(r.banner + "\r\n")
}
