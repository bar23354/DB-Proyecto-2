package models

import (
	"database/sql"
	"errors"
)

// Seat representa la estructura de un asiento en el sistema.
type Seat struct {
	ID       int
	Number   string
	IsBooked bool
}

// GetSeatByID consulta un asiento por su ID en la base de datos.
func GetSeatByID(db *sql.DB, id int) (Seat, error) {
	var seat Seat
	query := "SELECT id, number, is_booked FROM seats WHERE id = ?"
	err := db.QueryRow(query, id).Scan(&seat.ID, &seat.Number, &seat.IsBooked)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Seat{}, errors.New("seat not found")
		}
		return Seat{}, err
	}
	return seat, nil
}

// ReserveSeat reserva un asiento para un usuario utilizando una transacción.
func ReserveSeat(tx *sql.Tx, seatID int, userID int) error {
	// Verificar si el asiento ya está reservado.
	var isBooked bool
	queryCheck := "SELECT is_booked FROM seats WHERE id = ?"
	err := tx.QueryRow(queryCheck, seatID).Scan(&isBooked)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("seat not found")
		}
		return err
	}
	if isBooked {
		return errors.New("seat is already booked")
	}

	// Reservar el asiento.
	queryUpdate := "UPDATE seats SET is_booked = 1, booked_by = ? WHERE id = ?"
	_, err = tx.Exec(queryUpdate, userID, seatID)
	if err != nil {
		return err
	}

	return nil
}