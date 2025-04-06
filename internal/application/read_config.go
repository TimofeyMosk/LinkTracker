package application

import (
	"log/slog"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type ScrapperConfig struct {
	Address          string        `yaml:"address"`
	BotBaseURL       string        `yaml:"bot_baseurl"`
	Interval         time.Duration `yaml:"scrap_interval" `
	ReadTimeout      time.Duration `yaml:"read_timeout" `
	WriteTimeout     time.Duration `yaml:"write_timeout" `
	BotClientTimeout time.Duration `yaml:"bot_client_timeout" `
	LogsPath         string        `yaml:"logger_path" `
}

type BotConfig struct {
	TgToken               string        `yaml:"tg_token" `
	Address               string        `yaml:"address"`
	ScrapperBaseURL       string        `yaml:"scrapper_baseurl"`
	ReadTimeout           time.Duration `yaml:"read_timeout" `
	WriteTimeout          time.Duration `yaml:"write_timeout" `
	ScrapperClientTimeout time.Duration `yaml:"scrapper_client_timeout" `
	LogsPath              string        `yaml:"logger_path" `
}

type DBConfig struct {
	PostgresUser     string `yaml:"postgres_user"`
	PostgresPassword string `yaml:"postgres_password"`
	PostgresDB       string `yaml:"postgres_db"`
}

type Config struct {
	ScrapConfig ScrapperConfig `yaml:"scrapper" `
	BotConfig   BotConfig      `yaml:"bot" `
	DBConfig    DBConfig       `yaml:"db"`
}

func ReadYAMLConfig(filePath string) (*Config, error) {
	if err := godotenv.Load(); err != nil {
		slog.Error("Failed to load the .env file, continue without it")
	}

	viper.SetConfigFile(filePath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		slog.Error("Error reading configuration file", "error", err)
		return nil, err
	}

	viper.AutomaticEnv()
	config := Config{
		ScrapConfig: ScrapperConfig{
			Address:          viper.GetString("scrapper.address"),
			BotBaseURL:       viper.GetString("scrapper.bot_baseurl"),
			Interval:         viper.GetDuration("scrapper.scrap_interval"),
			ReadTimeout:      viper.GetDuration("scrapper.read_timeout"),
			WriteTimeout:     viper.GetDuration("scrapper.write_timeout"),
			BotClientTimeout: viper.GetDuration("scrapper.bot_client_timeout"),
			LogsPath:         viper.GetString("scrapper.logger_path"),
		},
		BotConfig: BotConfig{
			TgToken:               viper.GetString("bot.tg_token"),
			Address:               viper.GetString("bot.address"),
			ScrapperBaseURL:       viper.GetString("bot.scrapper_baseurl"),
			ReadTimeout:           viper.GetDuration("bot.read_timeout"),
			WriteTimeout:          viper.GetDuration("bot.write_timeout"),
			ScrapperClientTimeout: viper.GetDuration("bot.scrapper_client_timeout"),
			LogsPath:              viper.GetString("bot.logger_path"),
		},
		DBConfig: DBConfig{
			PostgresUser:     viper.GetString("db.postgres_user"),
			PostgresPassword: viper.GetString("db.postgres_password"),
			PostgresDB:       viper.GetString("db.postgres_db"),
		},
	}

	// Replace with a token from an environment variable if one exists
	if viper.GetString("tg_token") != "" {
		config.BotConfig.TgToken = viper.GetString("tg_token")
	}

	if viper.GetString("postgres_user") != "" {
		config.DBConfig.PostgresUser = viper.GetString("postgres_user")
	}

	if viper.GetString("postgres_password") != "" {
		config.DBConfig.PostgresPassword = viper.GetString("postgres_password")
	}

	if viper.GetString("postgres_db") != "" {
		config.DBConfig.PostgresDB = viper.GetString("postgres_db")
	}

	return &config, nil
}
