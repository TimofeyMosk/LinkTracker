package domain

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

var BotCommands = []tgbotapi.BotCommand{
	{
		Command:     "start",
		Description: "Начало работы с ботом",
	},
	{
		Command:     "help",
		Description: "Помощь по командам",
	},
	{
		Command:     "track",
		Description: "Начать отслеживание ссылки",
	},
	{
		Command:     "untrack",
		Description: "Прекратить отслеживание",
	},
	{
		Command:     "list",
		Description: "Список отслеживаемых ссылок",
	},
}
