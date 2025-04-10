package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "sync"
    "sync/atomic"
    "time"
    "math/rand"
)

func RunSimulation(config SimulationConfig) SimulationResult {
    db := GetDB()
    resultChan := make(chan ReservationResult, config.ConcurrencyLevel)
    var wg sync.WaitGroup
    var deadlockCount int32
    var successCount int32

    // Resetear el estado inicial
    _, err := db.Exec("UPDATE asiento SET estado = 'disponible' WHERE id = $1", config.SeatID)
    if err != nil {
        log.Fatalf("Error reseteando asiento: %v", err)
    }

    for i := 0; i < config.ConcurrencyLevel; i++ {
        wg.Add(1)
        go func(passengerID int) {
            defer wg.Done()
            start := time.Now()

            // Usar patrones de concurrencia optimizados
            result := attemptOptimizedReservation(db, config, passengerID, start)
            
            if result.Success {
                atomic.AddInt32(&successCount, 1)
            }
            resultChan <- result
        }(config.PassengerIDs[i%len(config.PassengerIDs)])
    }

    go func() {
        wg.Wait()
        close(resultChan)
    }()

    var failureCount int
    var totalDuration int64

    for result := range resultChan {
        if !result.Success {
            failureCount++
        }
        totalDuration += result.Duration
    }

    avgDuration := float64(totalDuration) / float64(config.ConcurrencyLevel)

    return SimulationResult{
        TotalAttempts:   config.ConcurrencyLevel,
        SuccessCount:    int(successCount),
        FailureCount:    failureCount,
        AvgDuration:     avgDuration,
        IsolationLevel:  config.IsolationLevel,
        Concurrency:     config.ConcurrencyLevel,
        Deadlocks:       int(deadlockCount),
    }
}

func attemptOptimizedReservation(db *sql.DB, config SimulationConfig, passengerID int, start time.Time) ReservationResult {
    // Usar una transacción corta y optimizada
    tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
        Isolation: sql.LevelReadCommitted,
    })
    if err != nil {
        return newFailedResult(config, passengerID, start, err)
    }
    defer tx.Rollback()

    // Patrón OPTIMISTIC CONCURRENCY CONTROL
    var currentVersion int
    err = tx.QueryRow(
        "SELECT version FROM asiento WHERE id = $1 FOR UPDATE",
        config.SeatID).Scan(&currentVersion)
    if err != nil {
        return newFailedResult(config, passengerID, start, err)
    }

    // Verificar disponibilidad
    var estado string
    err = tx.QueryRow(
        "SELECT estado FROM asiento WHERE id = $1",
        config.SeatID).Scan(&estado)
    if err != nil || estado != "disponible" {
        return newFailedResult(config, passengerID, start, fmt.Errorf("asiento no disponible"))
    }

    // Simular procesamiento breve
    time.Sleep(time.Duration(10+rand.Intn(20)) * time.Millisecond)

    // Actualizar con verificación de versión
    res, err := tx.Exec(
        "UPDATE asiento SET estado = 'reservado', version = version + 1 WHERE id = $1 AND version = $2",
        config.SeatID, currentVersion)
    if err != nil {
        return newFailedResult(config, passengerID, start, err)
    }

    rowsAffected, _ := res.RowsAffected()
    if rowsAffected == 0 {
        return newFailedResult(config, passengerID, start, fmt.Errorf("conflicto de concurrencia"))
    }

    // Crear reserva
    _, err = tx.Exec(
        "INSERT INTO reserva (asiento_id, pasajero_id, estado) VALUES ($1, $2, 'confirmada')",
        config.SeatID, passengerID)
    if err != nil {
        return newFailedResult(config, passengerID, start, err)
    }

    if err := tx.Commit(); err != nil {
        return newFailedResult(config, passengerID, start, err)
    }

    return ReservationResult{
        Success:     true,
        SeatID:      config.SeatID,
        PassengerID: passengerID,
        Duration:    time.Since(start).Milliseconds(),
        Isolation:   config.IsolationLevel,
        Concurrency: config.ConcurrencyLevel,
    }
}

func newFailedResult(config SimulationConfig, passengerID int, start time.Time, err error) ReservationResult {
    // Registrar fallo en la base de datos
    _, _ = db.Exec(
        "INSERT INTO reserva (asiento_id, pasajero_id, estado) VALUES ($1, $2, 'fallida') ON CONFLICT DO NOTHING",
        config.SeatID, passengerID)
    
    return ReservationResult{
        Success:     false,
        SeatID:      config.SeatID,
        PassengerID: passengerID,
        Duration:    time.Since(start).Milliseconds(),
        Error:       err,
        Isolation:   config.IsolationLevel,
        Concurrency: config.ConcurrencyLevel,
    }
}