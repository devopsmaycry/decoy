package services

type Redis struct{}

func (r Redis) Banner() []byte {
	return []byte("-NOAUTH Authentication required.\r\n")
}
