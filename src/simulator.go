package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

func RunSimulation(config SimulationConfig) SimulationResult {
	db := GetDB()
	resultChan := make(chan ReservationResult, config.ConcurrencyLevel)
	var wg sync.WaitGroup
	var deadlockCount int32
	var successCount int32

	// Resetear el estado inicial del asiento
	_, err := db.Exec("UPDATE asiento SET estado = 'disponible' WHERE id = $1", config.SeatID)
	if err != nil {
		log.Fatalf("Error reseteando asiento: %v", err)
	}

	// Limpiar reservas previas para este asiento
	_, err = db.Exec("DELETE FROM detallereserva WHERE asiento_id = $1", config.SeatID)
	if err != nil {
		log.Fatalf("Error limpiando detalles de reserva: %v", err)
	}
	_, err = db.Exec("DELETE FROM reserva WHERE id IN (SELECT reserva_id FROM detallereserva WHERE asiento_id = $1)", config.SeatID)
	if err != nil {
		log.Fatalf("Error limpiando reservas: %v", err)
	}

	for i := 0; i < config.ConcurrencyLevel; i++ {
		wg.Add(1)
		go func(attemptNum int) {
			defer wg.Done()
			passengerID := config.PassengerIDs[attemptNum%len(config.PassengerIDs)]
			start := time.Now()

			// Seleccionar nivel de aislamiento basado en la configuración
			var isolationLevel sql.IsolationLevel
			switch config.IsolationLevel {
			case "READ COMMITTED":
				isolationLevel = sql.LevelReadCommitted
			case "REPEATABLE READ":
				isolationLevel = sql.LevelRepeatableRead
			case "SERIALIZABLE":
				isolationLevel = sql.LevelSerializable
			default:
				isolationLevel = sql.LevelReadCommitted
			}

			result := attemptReservation(db, config, passengerID, start, isolationLevel, attemptNum+1)
			
			if result.Success {
				atomic.AddInt32(&successCount, 1)
			}
			if result.IsDeadlock {
				atomic.AddInt32(&deadlockCount, 1)
			}
			resultChan <- result
		}(i)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var results []ReservationResult
	var totalDuration time.Duration

	for result := range resultChan {
		results = append(results, result)
		totalDuration += time.Duration(result.Duration) * time.Millisecond
	}

	avgDuration := float64(totalDuration.Milliseconds()) / float64(config.ConcurrencyLevel)

	// Verificar estado final del asiento
	var finalState string
	err = db.QueryRow("SELECT estado FROM asiento WHERE id = $1", config.SeatID).Scan(&finalState)
	if err != nil {
		log.Printf("Error verificando estado final del asiento: %v", err)
	}

	return SimulationResult{
		TotalAttempts:   config.ConcurrencyLevel,
		SuccessCount:    int(successCount),
		FailureCount:    config.ConcurrencyLevel - int(successCount),
		AvgDuration:     avgDuration,
		IsolationLevel:  config.IsolationLevel,
		Concurrency:     config.ConcurrencyLevel,
		Deadlocks:       int(deadlockCount),
		FinalSeatState:  finalState,
		DetailedResults: results,
	}
}

func attemptReservation(db *sql.DB, config SimulationConfig, passengerID int, start time.Time, 
    isolationLevel sql.IsolationLevel, attempt int) ReservationResult {
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    tx, err := db.BeginTx(ctx, &sql.TxOptions{
        Isolation: isolationLevel,
    })
    if err != nil {
        log.Printf("Error iniciando transacción (Intento %d): %v", attempt, err)
        return newFailedResult(config, passengerID, start, err, false)
    }
    defer tx.Rollback()

    // Obtener información completa del asiento con bloqueo
    var asiento struct {
        estado     string
        precio     float64
        version    int
        vueloID    int
    }
    
    err = tx.QueryRowContext(ctx, 
        "SELECT estado, precio_base, vuelo_id FROM asiento WHERE id = $1 FOR UPDATE",
        config.SeatID,
    ).Scan(&asiento.estado, &asiento.precio, &asiento.vueloID)
    
    if err != nil {
        log.Printf("Error obteniendo estado del asiento (Intento %d): %v", attempt, err)
        return newFailedResult(config, passengerID, start, err, isDeadlockError(err))
    }

    // Verificar disponibilidad
    if asiento.estado != "disponible" {
        errMsg := fmt.Sprintf("asiento no disponible (estado: %s)", asiento.estado)
        log.Printf("Intento %d: %s", attempt, errMsg)
        return newFailedResult(config, passengerID, start, fmt.Errorf(errMsg), false)
    }

    // Simular procesamiento variable (entre 10-60ms)
    processingTime := time.Duration(10+rand.Intn(50)) * time.Millisecond
    time.Sleep(processingTime)

    // Crear reserva principal
    codigoReserva := fmt.Sprintf("RES-%d-%d-%d", time.Now().Unix(), config.SeatID, attempt)
    var reservaID int
    
    err = tx.QueryRowContext(ctx,
        `INSERT INTO reserva (codigo_reserva, pasajero_id, estado, total_pago, session_id)
         VALUES ($1, $2, 'confirmada', $3, $4) RETURNING id`,
        codigoReserva, passengerID, asiento.precio, fmt.Sprintf("session-%d", attempt),
    ).Scan(&reservaID)
    
    if err != nil {
        log.Printf("Error creando reserva principal (Intento %d): %v", attempt, err)
        return newFailedResult(config, passengerID, start, err, isDeadlockError(err))
    }

    // Crear detalle de reserva
    _, err = tx.ExecContext(ctx,
        `INSERT INTO detallereserva (reserva_id, asiento_id, precio_final)
         VALUES ($1, $2, $3)`,
        reservaID, config.SeatID, asiento.precio,
    )
    if err != nil {
        log.Printf("Error creando detalle de reserva (Intento %d): %v", attempt, err)
        return newFailedResult(config, passengerID, start, err, isDeadlockError(err))
    }

    // Actualizar estado del asiento
    res, err := tx.ExecContext(ctx,
        `UPDATE asiento SET estado = 'reservado' 
         WHERE id = $1 AND estado = 'disponible'`,
        config.SeatID,
    )
    if err != nil {
        log.Printf("Error actualizando asiento (Intento %d): %v", attempt, err)
        return newFailedResult(config, passengerID, start, err, isDeadlockError(err))
    }

    rowsAffected, err := res.RowsAffected()
    if err != nil {
        log.Printf("Error verificando filas afectadas (Intento %d): %v", attempt, err)
        return newFailedResult(config, passengerID, start, err, false)
    }

    if rowsAffected == 0 {
        errMsg := "el asiento ya fue reservado por otro proceso"
        log.Printf("Intento %d: %s", attempt, errMsg)
        return newFailedResult(config, passengerID, start, fmt.Errorf(errMsg), false)
    }

    if err := tx.Commit(); err != nil {
        log.Printf("Error confirmando transacción (Intento %d): %v", attempt, err)
        return newFailedResult(config, passengerID, start, err, isDeadlockError(err))
    }

    return ReservationResult{
        Success:      true,
        SeatID:       config.SeatID,
        PassengerID:  passengerID,
        Duration:     time.Since(start).Milliseconds(),
        Isolation:    config.IsolationLevel,
        Concurrency:  config.ConcurrencyLevel,
        Attempt:      attempt,
        ProcessingMS: processingTime.Milliseconds(),
    }
}

func isDeadlockError(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL deadlock error
	return err.Error() == "pq: deadlock detected" || 
	       err.Error() == "ERROR: deadlock detected (SQLSTATE 40P01)"
}

func newFailedResult(config SimulationConfig, passengerID int, start time.Time, err error, isDeadlock bool) ReservationResult {
    errorMsg := "unknown error"
    if err != nil {
        errorMsg = err.Error()
    }
    
    // Registrar intento fallido en la base de datos
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    
    _, dbErr := db.ExecContext(ctx,
        `INSERT INTO reserva (codigo_reserva, pasajero_id, estado, total_pago, session_id, error) 
         VALUES ($1, $2, 'fallida', 0, $3, $4)`,
        fmt.Sprintf("FAIL-%d-%d", time.Now().Unix(), passengerID),
        passengerID,
        fmt.Sprintf("session-%d", start.UnixNano()),
        errorMsg,
    )
    
    if dbErr != nil {
        log.Printf("Error registrando reserva fallida: %v", dbErr)
    }

    return ReservationResult{
        Success:     false,
        SeatID:      config.SeatID,
        PassengerID: passengerID,
        Duration:    time.Since(start).Milliseconds(),
        Error:       err,
        Isolation:   config.IsolationLevel,
        Concurrency: config.ConcurrencyLevel,
        IsDeadlock:  isDeadlock,
        ErrorDetail: errorMsg,
    }
}