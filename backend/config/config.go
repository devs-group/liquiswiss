package config

import (
	"liquiswiss/pkg/utils"
	"os"
	"strconv"
	"time"
)

type Config struct {
	WebHost               string
	DBUser                string
	DBPassword            string
	DBHost                string
	DBPort                string
	DBName                string
	JWTKey                []byte
	SMTPHost              string
	SMTPPort              int
	SMTPUser              string
	SMTPPass              string
	SMTPFromAddress       string
	SMTPFromName          string
	SMTPTLS               string
	FixerIOURl            string
	FixerIOKey            string
	ResetPasswordDelay    time.Duration
	ResetPasswordValidity time.Duration
	InvitationResendDelay time.Duration
	InvitationValidity    time.Duration
}

func GetConfig() Config {
	return Config{
		WebHost: getEnv("WEB_HOST", "http://localhost:3000"),

		DBUser:     getEnv("DB_USER", ""),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBHost:     getEnv("DB_HOST", ""),
		DBPort:     getEnv("DB_PORT", ""),
		DBName:     getEnv("DB_NAME", ""),

		JWTKey: []byte(getEnv("JWT_KEY", "")),

		SMTPHost:        getEnv("SMTP_HOST", ""),
		SMTPPort:        getEnvInt("SMTP_PORT", 1025),
		SMTPUser:        getEnv("SMTP_USER", ""),
		SMTPPass:        getEnv("SMTP_PASS", ""),
		SMTPFromAddress: getEnv("SMTP_FROM_ADDRESS", "no-reply@liquiswiss.local"),
		SMTPFromName:    getEnv("SMTP_FROM_NAME", "LiquiSwiss"),
		SMTPTLS:         getEnv("SMTP_TLS", ""),

		FixerIOURl: getEnv("FIXER_IO_URL", ""),
		FixerIOKey: getEnv("FIXER_IO_KEY", ""),

		ResetPasswordDelay:    getEnvDurationMinutes("RESET_PASSWORD_DELAY_MINUTES", utils.ResetPasswordDelay),
		ResetPasswordValidity: getEnvDurationMinutes("RESET_PASSWORD_VALIDITY_MINUTES", utils.ResetPasswordValidity),
		InvitationResendDelay: getEnvDurationMinutes("INVITATION_RESEND_DELAY_MINUTES", utils.InvitationResendDelay),
		InvitationValidity:    getEnvDurationMinutes("INVITATION_VALIDITY_MINUTES", utils.InvitationValidity),
	}
}

func getEnvDurationMinutes(key string, fallback time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if n, err := strconv.Atoi(value); err == nil && n > 0 {
			return time.Duration(n) * time.Minute
		}
	}
	return fallback
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if n, err := strconv.Atoi(value); err == nil {
			return n
		}
	}
	return fallback
}
