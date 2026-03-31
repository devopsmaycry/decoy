package services

// MySQL sends a Protocol v10 Initial Handshake Packet (MySQL 5.7.42).
// The server speaks first, so this works as a banner without any client input.
type MySQL struct{}

func (m MySQL) Banner() []byte {
	serverVersion := []byte("5.7.42-log\x00")

	payload := []byte{0x0a} // protocol version 10
	payload = append(payload, serverVersion...)
	payload = append(payload, 0x08, 0x00, 0x00, 0x00) // connection_id = 8
	payload = append(payload,
		0x52, 0x7b, 0x3d, 0x49, 0x5a, 0x61, 0x34, 0x76, // auth-plugin-data-1 (8 bytes)
		0x00,       // filler
		0xff, 0xf7, // capability flags lower
		0x21,       // character set: utf8
		0x02, 0x00, // status flags: SERVER_STATUS_AUTOCOMMIT
		0xff, 0x81, // capability flags upper
		0x15,                                           // auth-plugin-data length = 21
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // reserved (10 bytes)
		0x21, 0x4f, 0x5a, 0x33, 0x44, 0x38, 0x61, 0x62, 0x43, 0x39, 0x76, 0x57, 0x00, // auth-plugin-data-2 (13 bytes)
	)
	payload = append(payload, []byte("mysql_native_password\x00")...)

	length := len(payload)
	header := []byte{byte(length), byte(length >> 8), byte(length >> 16), 0x00}
	return append(header, payload...)
}
