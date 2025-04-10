package main

import (
	"fmt"
	"os"
	"log"
	"sync"
	"text/tabwriter"
	"time"
)

func main() {
	// Inicializar base de datos
	InitializeDatabase()

	// Configuración de pruebas
	concurrencyLevels := []int{5, 10, 20, 30}
	isolationLevels := []string{"READ COMMITTED", "REPEATABLE READ", "SERIALIZABLE"}
	passengerIDs := []int{1001, 1002, 1003, 1004, 1005}
	seatID := 1 // Todos competirán por el mismo asiento

	// Crear archivo de resultados
	file, err := os.Create("resultados.txt")
	if err != nil {
		log.Fatalf("Error al crear archivo de resultados: %v", err)
	}
	defer file.Close()

	// Escribir cabecera en archivo
	file.WriteString("Resultados de Simulación - " + time.Now().Format("2006-01-02 15:04:05") + "\n\n")
	file.WriteString("Concurrencia\tNivel Aislamiento\tReservas Exitosas\tReservas Fallidas\tDeadlocks\tTiempo Promedio (ms)\n")

	// Configurar tabulador para consola
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Concurrencia\tNivel Aislamiento\tReservas Exitosas\tReservas Fallidas\tDeadlocks\tTiempo Promedio (ms)")

	var results []SimulationResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, concurrency := range concurrencyLevels {
		for _, isolation := range isolationLevels {
			wg.Add(1)
			go func(concurrency int, isolation string) {
				defer wg.Done()

				ResetDatabase()

				config := SimulationConfig{
					TotalSeats:       30,
					ConcurrencyLevel: concurrency,
					IsolationLevel:   isolation,
					PassengerIDs:     passengerIDs,
					SeatID:          seatID,
				}

				result := RunSimulation(config)

				mu.Lock()
				results = append(results, result)
				mu.Unlock()
			}(concurrency, isolation)
		}
	}

	wg.Wait()

	// Escribir resultados en consola y archivo
	for _, concurrency := range concurrencyLevels {
		for _, isolation := range isolationLevels {
			for _, result := range results {
				if result.Concurrency == concurrency && result.IsolationLevel == isolation {
					// Escribir en consola
					fmt.Fprintf(w, "%d\t%s\t%d\t%d\t%d\t%.2f\n",
						result.Concurrency,
						result.IsolationLevel,
						result.SuccessCount,
						result.FailureCount,
						result.Deadlocks,
						result.AvgDuration,
					)

					// Escribir en archivo
					file.WriteString(fmt.Sprintf("%d\t%s\t%d\t%d\t%d\t%.2f\n",
						result.Concurrency,
						result.IsolationLevel,
						result.SuccessCount,
						result.FailureCount,
						result.Deadlocks,
						result.AvgDuration,
					))
				}
			}
		}
	}

	w.Flush()

	// Escribir resumen estadístico
	file.WriteString("\nResumen Estadístico:\n")
	fmt.Println("\nResumen Estadístico:")
	for _, isolation := range isolationLevels {
		var totalSuccess, totalFailures, totalDeadlocks int
		var totalDuration float64
		var count int

		for _, result := range results {
			if result.IsolationLevel == isolation {
				totalSuccess += result.SuccessCount
				totalFailures += result.FailureCount
				totalDeadlocks += result.Deadlocks
				totalDuration += result.AvgDuration
				count++
			}
		}

		if count > 0 {
			avgDuration := totalDuration / float64(count)
			summary := fmt.Sprintf("%s - Éxito: %.1f%%, Fallos: %.1f%%, Deadlocks: %d, Tiempo avg: %.2fms\n",
				isolation,
				float64(totalSuccess)/float64(totalSuccess+totalFailures)*100,
				float64(totalFailures)/float64(totalSuccess+totalFailures)*100,
				totalDeadlocks,
				avgDuration,
			)

			fmt.Print(summary)
			file.WriteString(summary)
		}
	}

	fmt.Println("\nResultados guardados en resultados.txt")
}