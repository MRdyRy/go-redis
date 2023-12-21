package config

import "github.com/spf13/viper"

type ConfigureEnv struct {
	RedisUrl  string `mapstructure:"REDIS.URL"`
	RedisPort string `mapstructure:"REDIS.PORT"`
	RedisPass string `mapstructure:"REDIS.PASS"`
	DbUrl     string `mapstructure:"DB.URL"`
	DbPort    string `mapstructure:"DB.PORT"`
	DbUser    string `mapstructure:"DB.USER"`
	DbPass    string `mapstructure:"DB.PASS"`
	DbName    string `mapstructure:"DB.NAME"`
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
