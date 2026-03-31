package services

// Service defines the behavior of a simulated network service.
// Banner returns the initial bytes sent to a connecting client.
type Service interface {
	Banner() []byte
}

var registry = map[string]Service{
	"smtp":  SMTP{},
	"ftp":   FTP{},
	"redis": Redis{},
	"mysql": MySQL{},
	"mssql": MSSQL{},
}

// Get returns the Service for the given name, or nil if unknown.
func Get(name string) Service {
	return registry[name]
}
