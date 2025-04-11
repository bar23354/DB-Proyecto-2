package simulation

import "fmt"

// Definimos los distintos niveles de aislamiento como constantes
const (
	ReadUncommitted = "READ UNCOMMITTED"
	ReadCommitted   = "READ COMMITTED"
	RepeatableRead  = "REPEATABLE READ"
	Serializable    = "SERIALIZABLE"
)

// Mapeamos los nombres al código SQL correspondiente
func GetIsolationQuery(level string) string {
	switch level {
	case ReadUncommitted:
		return "SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;"
	case ReadCommitted:
		return "SET TRANSACTION ISOLATION LEVEL READ COMMITTED;"
	case RepeatableRead:
		return "SET TRANSACTION ISOLATION LEVEL REPEATABLE READ;"
	case Serializable:
		return "SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;"
	default:
		return fmt.Sprintf("ERROR: Isolation level '%s' not recognized", level)
	}
}