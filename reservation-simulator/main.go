package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"reservation-simulator/config"
	"reservation-simulator/simulation"
)

func main() {
	var userCount int
	var isolationLevel string

	flag.IntVar(&userCount, "users", 10, "Cantidad de usuarios para la simulación")
	flag.StringVar(&isolationLevel, "isolation", "read_committed", "Nivel de aislamiento (read_uncommitted, read_committed, repeatable_read, serializable)")
	flag.Parse()

	if userCount <= 0 {
		log.Fatalf("La cantidad de usuarios debe ser mayor a 0")
	}

	validIsolationLevels := map[string]bool{
		"read_uncommitted": true,
		"read_committed":   true,
		"repeatable_read":  true,
		"serializable":     true,
	}

	if !validIsolationLevels[isolationLevel] {
		log.Fatalf("Nivel de aislamiento inválido: %s", isolationLevel)
	}

	db, err := config.CreateDatabaseConnection()
	if err != nil {
		log.Fatalf("Error al conectar a la base de datos: %v", err)
	}

	result, err := simulation.RunSimulation(db, userCount, isolationLevel)
	if err != nil {
		log.Fatalf("Error al ejecutar la simulación: %v", err)
	}

	fmt.Println("Simulación completada exitosamente.")

	// Guardar en CSV
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