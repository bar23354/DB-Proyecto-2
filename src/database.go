package main

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"
	
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "R20e59e01111"
	dbname   = "reservaciones_vuelos"
)

var db *sql.DB
var once sync.Once

func GetDB() *sql.DB {
	once.Do(func() {
		psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname)

		var err error
		db, err = sql.Open("postgres", psqlInfo)
		if err != nil {
			log.Fatalf("Error al conectar a la base de datos: %v", err)
		}

		// Configurar parámetros de conexión
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(25)
		db.SetConnMaxLifetime(5 * time.Minute)

		err = db.Ping()
		if err != nil {
			log.Fatalf("Error al hacer ping a la base de datos: %v", err)
		}

		fmt.Println("Conexión exitosa a la base de datos")
	})
	return db
}

func InitializeDatabase() {
	db := GetDB()

	// Eliminar tablas existentes para pruebas
	_, _ = db.Exec("DROP TABLE IF EXISTS reserva, asiento, vuelo")

	// Crear tablas con estructura optimizada
	_, err := db.Exec(`CREATE TABLE vuelo (
		id SERIAL PRIMARY KEY,
		numero_vuelo VARCHAR(10) NOT NULL,
		origen VARCHAR(50) NOT NULL,
		destino VARCHAR(50) NOT NULL,
		fecha_hora_salida TIMESTAMP NOT NULL
	)`)
	if err != nil {
		log.Fatalf("Error al crear tabla vuelo: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE asiento (
		id SERIAL PRIMARY KEY,
		vuelo_id INTEGER NOT NULL REFERENCES vuelo(id),
		codigo VARCHAR(5) NOT NULL,
		clase VARCHAR(20) NOT NULL,
		estado VARCHAR(20) NOT NULL DEFAULT 'disponible',
		version INTEGER NOT NULL DEFAULT 0,
		UNIQUE(vuelo_id, codigo)
	)`)
	if err != nil {
		log.Fatalf("Error al crear tabla asiento: %v", err)
	}

	// Tabla reserva con control de concurrencia optimista
	_, err = db.Exec(`CREATE TABLE reserva (
		id SERIAL PRIMARY KEY,
		asiento_id INTEGER NOT NULL REFERENCES asiento(id),
		pasajero_id INTEGER NOT NULL,
		fecha_reserva TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		estado VARCHAR(20) NOT NULL CHECK (estado IN ('confirmada', 'fallida')),
		session_id VARCHAR(100),
		CONSTRAINT unique_asiento_confirmado UNIQUE (asiento_id)
	)`)
	if err != nil {
		log.Fatalf("Error al crear tabla reserva: %v", err)
	}

	// Crear índices para mejorar el rendimiento
	_, err = db.Exec("CREATE INDEX idx_asiento_estado ON asiento(estado) WHERE estado = 'disponible'")
	if err != nil {
		log.Printf("Error creando índice: %v", err)
	}

	// Insertar datos de prueba
	_, err = db.Exec(`INSERT INTO vuelo (numero_vuelo, origen, destino, fecha_hora_salida) 
		VALUES ('AV101', 'GUA', 'MIA', NOW() + INTERVAL '1 day')
		ON CONFLICT (numero_vuelo) DO NOTHING`)
	if err != nil {
		log.Printf("Error al insertar vuelo: %v", err)
	}

	// Insertar 30 asientos para pruebas
	for i := 1; i <= 30; i++ {
		_, err = db.Exec(fmt.Sprintf(`INSERT INTO asiento (vuelo_id, codigo, clase, estado) 
			VALUES (1, '%dA', 'Economica', 'disponible')
			ON CONFLICT (vuelo_id, codigo) DO NOTHING`, i))
		if err != nil {
			log.Printf("Error al insertar asiento %dA: %v", i, err)
		}
	}
}

func ResetDatabase() {
	db := GetDB()
	
	// Usar transacción para reset atómico
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error iniciando transacción: %v", err)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("TRUNCATE TABLE reserva")
	if err != nil {
		log.Printf("Error limpiando reservas: %v", err)
		return
	}

	_, err = tx.Exec("UPDATE asiento SET estado = 'disponible', version = 0")
	if err != nil {
		log.Printf("Error reseteando asientos: %v", err)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error confirmando transacción: %v", err)
	}
}

// Función optimizada para reservas
func ReserveSeat(db *sql.DB, seatID, passengerID int) (bool, error) {
	tx, err := db.Begin()
	if err != nil {
		return false, fmt.Errorf("error iniciando transacción: %v", err)
	}
	defer tx.Rollback()

	// Obtener versión actual del asiento
	var currentVersion int
	var estado string
	err = tx.QueryRow(
		"SELECT version, estado FROM asiento WHERE id = $1 FOR UPDATE",
		seatID).Scan(&currentVersion, &estado)
	if err != nil {
		return false, fmt.Errorf("error consultando asiento: %v", err)
	}

	if estado != "disponible" {
		return false, nil
	}

	// Actualizar con control de concurrencia optimista
	res, err := tx.Exec(
		"UPDATE asiento SET estado = 'reservado', version = version + 1 WHERE id = $1 AND version = $2",
		seatID, currentVersion)
	if err != nil {
		return false, fmt.Errorf("error actualizando asiento: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return false, nil
	}

	// Crear registro de reserva
	_, err = tx.Exec(
		"INSERT INTO reserva (asiento_id, pasajero_id, estado) VALUES ($1, $2, 'confirmada')",
		seatID, passengerID)
	if err != nil {
		return false, fmt.Errorf("error creando reserva: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("error confirmando reserva: %v", err)
	}

	return true, nil
}