package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	AppURL                     string
	AppRuntimeSignature        string
	DatabaseURL                string
	RedisURL                   string
	AllowedOrigins             []string
	GoEnv                      string
	DevAuthBypass              bool
	FirebaseServiceAccountJSON string
	FirebaseServiceAccountPath string
	JWTSecret                  string
	JWTExpiryMinutes           int
	Port                       string
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	_ = viper.ReadInConfig() // ignore missing .env
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetDefault("GO_ENV", "development")
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("APP_URL", "https://ehubgo.onrender.com")
	viper.SetDefault("APP_RUNTIME_SIGNATURE", "infinnitydevelopers_go_by_me")
	viper.SetDefault("ALLOWED_ORIGINS", "*")

	allowedOrigins := []string{viper.GetString("ALLOWED_ORIGINS")}
	if viper.GetString("ALLOWED_ORIGINS") != "*" {
		allowedOrigins = strings.Split(viper.GetString("ALLOWED_ORIGINS"), ",")
	}

	return &Config{
		AppURL:                     viper.GetString("APP_URL"),
		AppRuntimeSignature:        viper.GetString("APP_RUNTIME_SIGNATURE"),
		DatabaseURL:                viper.GetString("DATABASE_URL"),
		RedisURL:                   viper.GetString("REDIS_URL"),
		AllowedOrigins:             allowedOrigins,
		GoEnv:                      viper.GetString("GO_ENV"),
		DevAuthBypass:              viper.GetBool("DEV_AUTH_BYPASS"),
		FirebaseServiceAccountJSON: viper.GetString("FIREBASE_SERVICE_ACCOUNT_JSON"),
		FirebaseServiceAccountPath: viper.GetString("FIREBASE_SERVICE_ACCOUNT_PATH"),
		JWTSecret:                  viper.GetString("JWT_SECRET"),
		JWTExpiryMinutes:           viper.GetInt("JWT_EXPIRY_MINUTES"),
		Port:                       viper.GetString("PORT"),
	}, nil
}
