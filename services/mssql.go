package services

// MSSQL sends a TDS Pre-Login Response simulating SQL Server 2019 (15.0.2000.5).
// Note: In real TDS the client speaks first. This banner is sent unconditionally
// on connect and is sufficient for scanners that probe port 1433.
type MSSQL struct{}

func (m MSSQL) Banner() []byte {
	// Pre-Login option list (relative offsets from start of Pre-Login data):
	//   VERSION    type=0x00, offset=11 (0x000B), length=6
	//   ENCRYPTION type=0x01, offset=17 (0x0011), length=1
	//   Terminator 0xFF
	// Followed by:
	//   VERSION data:    SQL Server 2019 RTM → 15.0.2000.5 → 0x0F 0x00 0x07 0xD0 0x00 0x05
	//   ENCRYPTION data: ENCRYPT_NOT_SUP (0x02)
	//
	// Total Pre-Login payload: 11 + 6 + 1 = 18 bytes
	// Total TDS packet:         8 (header) + 18 = 26 = 0x1A
	return []byte{
		// TDS header (8 bytes)
		// Type 0x04 = TABULAR_RESULT — correct packet type for a Pre-Login Response.
		// The client Pre-Login Request uses 0x12; the server response uses 0x04.
		// nmap matches: \x04\x01..\x00\x00..\x00\x00\x15\x00\x06
		0x04,       // Type: TABULAR_RESULT (Pre-Login Response)
		0x01,       // Status: End of message
		0x00, 0x1A, // Length: 26
		0x00, 0x00, // SPID
		0x01,       // Packet ID
		0x00,       // Window
		// Pre-Login options
		0x00, 0x00, 0x0B, 0x00, 0x06, // VERSION:    offset=11, length=6
		0x01, 0x00, 0x11, 0x00, 0x01, // ENCRYPTION: offset=17, length=1
		0xFF, // Terminator
		// VERSION data: SQL Server 2019 RTM (15.0.2000.5)
		0x0F, 0x00, 0x07, 0xD0, 0x00, 0x05,
		// ENCRYPTION: ENCRYPT_NOT_SUP
		0x02,
	}
}
