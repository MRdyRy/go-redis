package config

import "github.com/spf13/viper"

type ConfigureEnv struct {
	RedisUrl  string `mapstructure:"URL"`
	RedisPort string `mapstructure:"PORT"`
	RedisPass string `mapstructure:"PASS"`
}

func LoadConfig(path string) (config ConfigureEnv, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("dev")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)

	return
}
