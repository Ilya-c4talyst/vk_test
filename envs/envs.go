package envs

import (
	"os"

	"github.com/joho/godotenv"
)

// Глобальная структура для хранения данных из env
var ServerEnvs Envs

type Envs struct {
	PORT string
}

// Инициализация значений ENV
func LoadEnvs() error {
	// Если файл .env не найден, то выводим сообщение
	if err := godotenv.Load(); err != nil {
		return err
	}
	// Инициализация значений ENV
	ServerEnvs.PORT = os.Getenv("PORT")
	return nil
}
