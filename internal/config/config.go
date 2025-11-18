package config

type Config struct {
	DBconfig string
}

func NewCon() Config {
	return Config{
		DBconfig: "host=localhost port=5432 user=postgres password=1234 dbname=data sslmode=disable",
	}
}
