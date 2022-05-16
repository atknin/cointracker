package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type config struct {
	Symbols                  []string `required:"true" envconfig:"SYSBOLS"`
	BotChatID                int      `required:"true" envconfig:"BOT_CHAT_ID"`
	BotToken                 string   `required:"true" envconfig:"BOT_TOKEN"`
	PercentPriceChangedAlert float64  `required:"true" envconfig:"PERCENT_PRICE_CHANGED_ALERT"`
}

var instance = &config{}

func init() {
	if err := envconfig.Process("", instance); err != nil {
		panic(err)
	}
}

func Config() config {
	return *instance
}

func BotAPI() string {
	return fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", instance.BotToken)
}
