package application

import (
	"time"

	"github.com/spf13/viper"
)

type ScrapperConfig struct {
	Address           string
	BotBaseURL        string
	Interval          time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	BotClientTimeout  time.Duration
	LogsPath          string
	CheckLinksWorkers int
	SizeLinksPage     int64
	DBAccessType      string
}

type BotConfig struct {
	TgToken               string
	Address               string
	ScrapperBaseURL       string
	ReadTimeout           time.Duration
	WriteTimeout          time.Duration
	ScrapperClientTimeout time.Duration
	LogsPath              string
}

type DBConfig struct {
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
}

type Config struct {
	ScrapConfig ScrapperConfig
	BotConfig   BotConfig
	DBConfig    DBConfig
}

func ReadYAMLConfig() (*Config, error) {
	viper.AutomaticEnv()
	config := Config{
		ScrapConfig: ScrapperConfig{
			Address:           viper.GetString("SCRAPPER_ADDRESS"),
			BotBaseURL:        viper.GetString("BOT_BASEURL"),
			Interval:          viper.GetDuration("CHECK_LINKS_INTERVAL"),
			ReadTimeout:       viper.GetDuration("SCRAPPER_READ_TIMEOUT"),
			WriteTimeout:      viper.GetDuration("SCRAPPER_WRITE_TIMEOUT"),
			BotClientTimeout:  viper.GetDuration("BOT_CLIENT_TIMEOUT"),
			CheckLinksWorkers: viper.GetInt("CHECK_LINKS_WORKERS"),
			SizeLinksPage:     viper.GetInt64("SIZE_LINKS_PAGE"),
			DBAccessType:      viper.GetString("DB_ACCESS_TYPE"),
		},
		BotConfig: BotConfig{
			TgToken:               viper.GetString("TG_TOKEN"),
			Address:               viper.GetString("BOT_ADDRESS"),
			ScrapperBaseURL:       viper.GetString("SCRAPPER_BASEURL"),
			ReadTimeout:           viper.GetDuration("BOT_READ_TIMEOUT"),
			WriteTimeout:          viper.GetDuration("BOT_WRITE_TIMEOUT"),
			ScrapperClientTimeout: viper.GetDuration("SCRAPPER_CLIENT_TIMEOUT"),
		},
		DBConfig: DBConfig{
			PostgresUser:     viper.GetString("POSTGRES_USER"),
			PostgresPassword: viper.GetString("POSTGRES_PASSWORD"),
			PostgresDB:       viper.GetString("POSTGRES_DB"),
		},
	}

	return &config, nil
}
