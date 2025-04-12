package workers

import (
	"context"
	"database/sql"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
)

func mapIsolationLevel(level string) sql.IsolationLevel {
	switch level {
	case "read_uncommitted":
		return sql.LevelReadUncommitted
	case "read_committed":
		return sql.LevelReadCommitted
	case "repeatable_read":
		return sql.LevelRepeatableRead
	case "serializable":
		return sql.LevelSerializable
	default:
		return sql.LevelDefault
	}
}

func SimulateUser(db *sql.DB, asientoID int, userID int, level string) (bool, bool, error) {

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: mapIsolationLevel(level),
	})
	if err != nil {
		log.Printf("Usuario %d: Error al iniciar transacción: %v", userID, err)
		return false, false, err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			log.Printf("Usuario %d: Transacción revertida por pánico: %v", userID, p)
		}
	}()

	var estado string
	err = tx.QueryRow("SELECT estado FROM Asiento WHERE ID = $1 FOR UPDATE", asientoID).Scan(&estado)
	if err != nil {
		_ = tx.Rollback()
		log.Printf("Usuario %d: Error al consultar estado del asiento %d: %v", userID, asientoID, err)
		return false, false, err
	}

	log.Printf("Usuario %d: Estado actual del asiento %d: %s", userID, asientoID, estado)

	if estado != "disponible" {
		_ = tx.Rollback()
		log.Printf("Usuario %d: Asiento %d no está disponible", userID, asientoID)
		return false, false, nil
	}

	var reservaID int
	codigoReserva := uuid.New().String()
	err = tx.QueryRow(`
		INSERT INTO Reserva (codigo_reserva, pasajero_id, total_pago, estado)
		VALUES ($1, $2, $3, 'confirmada') RETURNING ID`,
		codigoReserva, userID, 100.00).Scan(&reservaID)
	if err != nil {
		_ = tx.Rollback()
		log.Printf("Usuario %d: Error al crear reserva: %v", userID, err)
		return false, false, err
	}
	log.Printf("Usuario %d: Reserva creada con ID %d", userID, reservaID)

	_, err = tx.Exec(`
		INSERT INTO DetalleReserva (reserva_id, asiento_id, precio_final)
		VALUES ($1, $2, $3)`, reservaID, asientoID, 100.00)
	if err != nil {
		_ = tx.Rollback()
		log.Printf("Usuario %d: Error al insertar detalle de reserva: %v", userID, err)
		return false, false, err
	}

	_, err = tx.Exec(`UPDATE Asiento SET estado = 'reservado' WHERE ID = $1`, asientoID)
	if err != nil {
		_ = tx.Rollback()
		log.Printf("Usuario %d: Error al actualizar estado del asiento: %v", userID, err)
		return false, false, err
	}

	simulateDelay(level)

	if err := tx.Commit(); err != nil {
		if strings.Contains(err.Error(), "deadlock") {
			log.Printf("Usuario %d: Transacción abortada por deadlock", userID)
			return false, true, nil
		}
		log.Printf("Usuario %d: Error al hacer commit: %v", userID, err)
		return false, false, err
	}

	log.Printf("Usuario %d: Transacción completada con éxito", userID)
	return true, false, nil
}

func simulateDelay(level string) {
	switch strings.ToLower(level) {
	case "read_uncommitted":
		time.Sleep(100 * time.Millisecond)
	case "read_committed":
		time.Sleep(200 * time.Millisecond)
	case "repeatable_read":
		time.Sleep(300 * time.Millisecond)
	case "serializable":
		time.Sleep(400 * time.Millisecond)
	default:
		time.Sleep(100 * time.Millisecond)
	}
}

func RunUserSimulation(db *sql.DB, userID int, level string) (bool, bool) {
	asientoID := rand.Intn(552) + 1
	success, deadlock, _ := SimulateUser(db, asientoID, userID, level)
	return success, deadlock
}
