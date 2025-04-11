package simulation

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"reservation-simulator/utils"
	"reservation-simulator/workers"
)

type SimResult struct {
	Success  bool
	Deadlock bool
	Duration float64 // en milisegundos
}

type SummaryResult struct {
	Concurrency int
	Isolation   string
	Successes   int
	Failures    int
	Deadlocks   int
	AvgTime     float64
	FinalState  string
}

func RunSimulation(db *sql.DB, numUsers int, level string) (SummaryResult, error) {
	var wg sync.WaitGroup
	logger := utils.NewLogger(true)
	results := make([]SimResult, numUsers)

	logger.Log("SYSTEM", "blue", fmt.Sprintf("Iniciando simulación con nivel de aislamiento: %s", level))

	for i := 0; i < numUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			start := time.Now()
			success, deadlock := workers.RunUserSimulation(db, userID+1, level)
			duration := time.Since(start).Seconds() * 1000

			results[userID] = SimResult{
				Success:  success,
				Deadlock: deadlock,
				Duration: duration,
			}
		}(i)
	}

	wg.Wait()

	successes := 0
	failures := 0
	deadlocks := 0
	var totalTime float64

	for _, r := range results {
		if r.Success {
			successes++
		} else {
			failures++
			if r.Deadlock {
				deadlocks++
			}
		}
		totalTime += r.Duration
	}

	avgTime := totalTime / float64(numUsers)

	var estadoFinal string
	err := db.QueryRow("SELECT estado FROM Asiento WHERE ID = $1", 1).Scan(&estadoFinal)
	if err != nil {
		estadoFinal = "desconocido"
	}

	fmt.Println("Concurrencia  Nivel Aislamiento    Éxitos  Fallos  Deadlocks  Tiempo Avg (ms)  Estado Final")
	fmt.Printf("%-12d %-19s %-7d %-7d %-10d %-17.2f %s\n", numUsers, formatIsolation(level), successes, failures, deadlocks, avgTime, estadoFinal)

	return SummaryResult{
		Concurrency: numUsers,
		Isolation:   formatIsolation(level),
		Successes:   successes,
		Failures:    failures,
		Deadlocks:   deadlocks,
		AvgTime:     avgTime,
		FinalState:  estadoFinal,
	}, nil
}

func formatIsolation(level string) string {
	switch level {
	case "read_uncommitted":
		return "READ UNCOMMITTED"
	case "read_committed":
		return "READ COMMITTED"
	case "repeatable_read":
		return "REPEATABLE READ"
	case "serializable":
		return "SERIALIZABLE"
	default:
		return level
	}
}
