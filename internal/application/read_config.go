package application

import (
	"log/slog"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type ScrapperConfig struct {
	Addr         string        `yaml:"addr"`
	BotBaseURL   string        `yaml:"bot_baseurl"`
	Interval     time.Duration `yaml:"scrap_interval" `
	ReadTimeout  time.Duration `yaml:"read_timeout" `
	WriteTimeout time.Duration `yaml:"write_timeout" `
}

type BotConfig struct {
	TgToken         string        `yaml:"tg_token" `
	Addr            string        `yaml:"addr"`
	ScrapperBaseURL string        `yaml:"scrapper_baseurl"`
	ReadTimeout     time.Duration `yaml:"read_timeout" `
	WriteTimeout    time.Duration `yaml:"write_timeout" `
}

type Config struct {
	ScrapConfig ScrapperConfig `yaml:"scrapper" `
	BotConfig   BotConfig      `yaml:"bot" `
	LogsPath    string         `yaml:"logger_path" `
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
			Addr:         viper.GetString("scrapper.addr"),
			BotBaseURL:   viper.GetString("scrapper.bot_baseurl"),
			Interval:     viper.GetDuration("scrapper.scrap_interval"),
			ReadTimeout:  viper.GetDuration("scrapper.read_timeout"),
			WriteTimeout: viper.GetDuration("scrapper.write_timeout"),
		},
		BotConfig: BotConfig{
			TgToken:         viper.GetString("bot.tg_token"),
			Addr:            viper.GetString("bot.addr"),
			ScrapperBaseURL: viper.GetString("bot.scrapper_baseurl"),
			ReadTimeout:     viper.GetDuration("bot.read_timeout"),
			WriteTimeout:    viper.GetDuration("bot.write_timeout"),
		},
		LogsPath: viper.GetString("logger_path"),
	}

	// Replace with a token from an environment variable if one exists
	if viper.GetString("tg_token") != "" {
		config.BotConfig.TgToken = viper.GetString("tg_token")
	}

	return &config, nil
}
