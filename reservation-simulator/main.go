package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"reservation-simulator/config"
	"reservation-simulator/simulation"
)

func main() {
	isolationLevels := []string{"read_committed", "repeatable_read", "serializable"}
	userCounts := []int{5, 10, 20, 30}

	db, err := config.CreateDatabaseConnection()
	if err != nil {
		log.Fatalf("Error al conectar a la base de datos: %v", err)
	}

	// Crear o abrir archivo CSV
	file, err := os.OpenFile("resultados_simulacion.csv", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("No se pudo abrir el archivo CSV: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Escribir encabezado si el archivo está vacío
	info, _ := file.Stat()
	if info.Size() == 0 {
		writer.Write([]string{"Concurrencia", "Nivel Aislamiento", "Exitos", "Fallos", "Deadlocks", "Tiempo Avg (ms)", "Estado Final"})
	}

	for _, userCount := range userCounts {
		for _, isolationLevel := range isolationLevels {
			fmt.Printf("Ejecutando simulación con %d usuarios y nivel de aislamiento %s...\n", userCount, isolationLevel)

			result, err := simulation.RunSimulation(db, userCount, isolationLevel)
			if err != nil {
				log.Printf("Error al ejecutar la simulación con %d usuarios y nivel %s: %v", userCount, isolationLevel, err)
				continue
			}

			// Guardar resultados en el archivo CSV
			writer.Write([]string{
				strconv.Itoa(result.Concurrency),
				result.Isolation,
				strconv.Itoa(result.Successes),
				strconv.Itoa(result.Failures),
				strconv.Itoa(result.Deadlocks),
				fmt.Sprintf("%.2f", result.AvgTime),
				result.FinalState,
			})
		}
	}

	fmt.Println("Todas las simulaciones se han completado.")
}

