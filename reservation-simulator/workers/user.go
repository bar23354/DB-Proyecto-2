package workers

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

func SimulateUser(db *sql.DB, asientoID int, userID int, level string) (bool, bool, error) {
	tx, err := db.Begin()
	if err != nil {
		return false, false, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			log.Printf("Transacción revertida por pánico para el usuario %d: %v", userID, p)
		}
	}()

	var estado string
	err = tx.QueryRow("SELECT estado FROM Asiento WHERE ID = $1 FOR UPDATE", asientoID).Scan(&estado)
	if err != nil {
		_ = tx.Rollback()
		return false, false, err
	}
	if estado != "disponible" {
		_ = tx.Rollback()
		return false, false, nil
	}

	var reservaID int
	err = tx.QueryRow(`
		INSERT INTO Reserva (codigo_reserva, pasajero_id, total_pago, estado)
		VALUES ($1, $2, $3, 'confirmada') RETURNING ID`,
		fmt.Sprintf("R%04d", userID), userID, 100.00).Scan(&reservaID)
	if err != nil {
		_ = tx.Rollback()
		return false, false, err
	}

	_, err = tx.Exec(`
		INSERT INTO DetalleReserva (reserva_id, asiento_id, precio_final)
		VALUES ($1, $2, $3)`, reservaID, asientoID, 100.00)
	if err != nil {
		_ = tx.Rollback()
		return false, false, err
	}

	_, err = tx.Exec(`UPDATE Asiento SET estado = 'reservado' WHERE ID = $1`, asientoID)
	if err != nil {
		_ = tx.Rollback()
		return false, false, err
	}

	simulateDelay(level)

	if err := tx.Commit(); err != nil {
		if strings.Contains(err.Error(), "deadlock") {
			return false, true, nil
		}
		return false, false, err
	}

	return true, false, nil
}

func simulateDelay(level string) {
	switch level {
	case "READ UNCOMMITTED":
		time.Sleep(100 * time.Millisecond)
	case "READ COMMITTED":
		time.Sleep(200 * time.Millisecond)
	case "REPEATABLE READ":
		time.Sleep(300 * time.Millisecond)
	case "SERIALIZABLE":
		time.Sleep(400 * time.Millisecond)
	default:
		time.Sleep(100 * time.Millisecond)
	}
}

func RunUserSimulation(db *sql.DB, userID int, level string) (bool, bool) {
	asientoID := 1 // todos intentan el mismo asiento para provocar concurrencia
	success, deadlock, _ := SimulateUser(db, asientoID, userID, level)
	return success, deadlock
}
