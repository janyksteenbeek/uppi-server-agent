package config

const (
	Version         = "1.0.0"
	DefaultInstance = "https://uppi.dev"
	DefaultInterval = 60 // seconds
)

type Config struct {
	Secret          string
	Instance        string
	ServerId        string
	AutoUpdate      bool
	IntervalMinutes int
}
