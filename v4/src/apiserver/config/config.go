package config

import (
	"fmt"

	"github.com/spf13/viper"

	"ds-lab2-bmstu/apiserver/core/ports/reservation"

	"ds-lab2-bmstu/apiserver/core/ports/library"
	"ds-lab2-bmstu/apiserver/core/ports/rating"
)

type Config struct {
	HTTPAddr    string `mapstructure:"http_addr"`
	Library     library.Config
	Rating      rating.Config
	Reservation reservation.Config
}

func ReadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/usr/local/etc/apiserver")
	viper.AddConfigPath(".")

	viper.SetDefault("http_addr", ":8080")
	viper.SetDefault("library.host", "library")
	viper.SetDefault("library.port", "8060")
	viper.SetDefault("rating.host", "rating")
	viper.SetDefault("rating.port", "8050")
	viper.SetDefault("reservation.host", "reservation")
	viper.SetDefault("reservation.port", "8070")

	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}