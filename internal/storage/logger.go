package storage

// Logger for printing out storage events.
type Logger interface {
	Infof(string, ...interface{})
}
