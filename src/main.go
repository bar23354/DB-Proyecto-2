package main

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"
	"time"
)

func main() {
	// Inicializar base de datos
	db := GetDB()
	defer db.Close()

	// Verificar estructura de la base de datos
	if err := VerifyDatabaseStructure(); err != nil {
		log.Fatalf("Error en estructura de BD: %v", err)
	}

	// Configuración de pruebas
	concurrencyLevels := []int{5, 10, 20, 30}
	isolationLevels := []string{"READ COMMITTED", "REPEATABLE READ", "SERIALIZABLE"}
	passengerIDs := generatePassengerIDs(30) // Generar 30 IDs de pasajero
	seatID := 1 // Todos competirán por el mismo asiento

	// Crear archivo de resultados
	file, err := os.Create("resultados_simulacion.txt")
	if err != nil {
		log.Fatalf("Error al crear archivo de resultados: %v", err)
	}
	defer file.Close()

	// Escribir cabecera en archivo
	file.WriteString("Resultados de Simulación de Reservaciones - " + time.Now().Format("2006-01-02 15:04:05") + "\n\n")
	file.WriteString("Concurrencia\tNivel Aislamiento\tReservas Exitosas\tReservas Fallidas\tDeadlocks\tTiempo Promedio (ms)\tEstado Final\n")

	// Configurar tabulador para consola
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Concurrencia\tNivel Aislamiento\tÉxitos\tFallos\tDeadlocks\tTiempo Avg (ms)\tEstado Final")

	var results []SimulationResult

	// Ejecutar simulaciones
	for _, concurrency := range concurrencyLevels {
		for _, isolation := range isolationLevels {
			// Reiniciar datos de prueba para cada simulación
			if err := ResetTestData(); err != nil {
				log.Printf("Error reiniciando datos: %v", err)
				continue
			}

			config := SimulationConfig{
				ConcurrencyLevel: concurrency,
				IsolationLevel:   isolation,
				PassengerIDs:     passengerIDs,
				SeatID:          seatID,
			}

			result := RunSimulation(config)
			results = append(results, result)

			// Escribir resultados
			row := fmt.Sprintf("%d\t%s\t%d\t%d\t%d\t%.2f\t%s\n",
				result.Concurrency,
				result.IsolationLevel,
				result.SuccessCount,
				result.FailureCount,
				result.Deadlocks,
				result.AvgDuration,
				result.FinalSeatState,
			)

			file.WriteString(row)
			fmt.Fprint(w, row)
		}
	}

	w.Flush()

	// Generar resumen estadístico
	generateSummary(results, file)
	fmt.Println("\nResultados guardados en resultados_simulacion.txt")
}

func generatePassengerIDs(count int) []int {
	ids := make([]int, count)
	for i := 0; i < count; i++ {
		ids[i] = 1000 + i + 1
	}
	return ids
}

func generateSummary(results []SimulationResult, file *os.File) {
	file.WriteString("\nResumen Estadístico:\n")
	fmt.Println("\nResumen Estadístico:")

	for _, isolation := range []string{"READ COMMITTED", "REPEATABLE READ", "SERIALIZABLE"} {
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
			successRate := float64(totalSuccess) / float64(totalSuccess+totalFailures) * 100
			deadlockRate := float64(totalDeadlocks) / float64(totalSuccess+totalFailures) * 100

			summary := fmt.Sprintf("%s - Tasa Éxito: %.1f%%, Fallos: %d, Deadlocks: %d (%.1f%%), Tiempo avg: %.2fms\n",
				isolation,
				successRate,
				totalFailures,
				totalDeadlocks,
				deadlockRate,
				avgDuration,
			)

			fmt.Print(summary)
			file.WriteString(summary)
		}
	}
}