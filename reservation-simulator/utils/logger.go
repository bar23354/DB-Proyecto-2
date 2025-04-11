package utils

import (
	"log"
	"os"
	"sync"
)

// Logger struct para manejar logs con prefijos y colores.
type Logger struct {
	mu      sync.Mutex
	colors  map[string]string
	enabled bool
}

// NewLogger crea una nueva instancia de Logger.
func NewLogger(enabled bool) *Logger {
	return &Logger{
		colors: map[string]string{
			"red":    "\033[31m",
			"green":  "\033[32m",
			"yellow": "\033[33m",
			"blue":   "\033[34m",
			"purple": "\033[35m",
			"cyan":   "\033[36m",
			"reset":  "\033[0m",
		},
		enabled: enabled,
	}
}

// Log imprime un mensaje con un prefijo de usuario y color.
func (l *Logger) Log(user string, color string, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.enabled {
		colorCode, exists := l.colors[color]
		if !exists {
			colorCode = l.colors["reset"]
		}
		log.Printf("%s[%s]%s %s", colorCode, user, l.colors["reset"], message)
	} else {
		log.Printf("[%s] %s", user, message)
	}
}

// Example usage
func main() {
	// Crear archivo de log
	file, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error al abrir el archivo de log: %v", err)
	}
	defer file.Close()

	log.SetOutput(file)

	logger := NewLogger(true)

	// Logs de ejemplo
	logger.Log("User1", "red", "Este es un mensaje de error.")
	logger.Log("User2", "green", "Este es un mensaje de éxito.")
	logger.Log("User3", "blue", "Este es un mensaje informativo.")
}

func (l *Logger) Info(user string, message string) {
	l.Log(user, "blue", message)
}

func (l *Logger) Success(user string, message string) {
	l.Log(user, "green", message)
}

func (l *Logger) Error(user string, message string) {
	l.Log(user, "red", message)
}
