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

		// Configuración optimizada para pruebas de concurrencia
		db.SetMaxOpenConns(50)
		db.SetMaxIdleConns(30)
		db.SetConnMaxLifetime(10 * time.Minute)
		db.SetConnMaxIdleTime(5 * time.Minute)

		// Verificar conexión
		err = db.Ping()
		if err != nil {
			log.Fatalf("Error al verificar conexión: %v", err)
		}

		log.Println("Conexión exitosa a la base de datos")
	})
	return db
}

func VerifyDatabaseStructure() error {
	db := GetDB()
	
	requiredTables := map[string][]string{
		"modeloavion":    {"id", "nombre", "capacidad_economica", "capacidad_business", "filas_economica", "filas_business", "asientos_por_fila"},
		"avion":          {"id", "modelo_id", "matricula", "estado"},
		"aeropuerto":     {"id", "codigo_iata", "nombre", "ciudad", "pais"},
		"vuelo":          {"id", "numero_vuelo", "avion_id", "aerolinea", "origen_id", "destino_id", "fecha_hora_salida", "fecha_hora_llegada", "estado", "puerta_embarque"},
		"claseservicio":  {"id", "nombre", "multiplicador_precio"},
		"asiento":        {"id", "vuelo_id", "codigo_asiento", "clase_servicio_id", "ubicacion", "estado", "precio_base"},
		"pasajero":       {"id", "nombre", "apellido", "tipo_documento", "numero_documento", "email", "telefono", "programa_millas"},
		"metodopago":     {"id", "tipo", "ultimos_digitos", "pasajero_id"},
		"reserva":        {"id", "codigo_reserva", "pasajero_id", "fecha_reserva", "estado", "metodo_pago_id", "total_pago", "session_id"},
		"detallereserva": {"id", "reserva_id", "asiento_id", "precio_final"},
		"auditoriareservas": {"id", "fecha_hora", "asiento_id", "estado_anterior", "estado_nuevo", "reserva_id", "usuario"},
	}

	for table, columns := range requiredTables {
		var tableExists bool
		err := db.QueryRow(
			"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)",
			table,
		).Scan(&tableExists)
		
		if err != nil {
			return fmt.Errorf("error verificando existencia de tabla %s: %v", table, err)
		}
		
		if !tableExists {
			return fmt.Errorf("tabla requerida no encontrada: %s", table)
		}

		for _, column := range columns {
			var columnExists bool
			err := db.QueryRow(
				"SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = $1 AND column_name = $2)",
				table, column,
			).Scan(&columnExists)
			
			if err != nil {
				return fmt.Errorf("error verificando columna %s en tabla %s: %v", column, table, err)
			}
			
			if !columnExists {
				return fmt.Errorf("columna requerida no encontrada: %s.%s", table, column)
			}
		}
	}

	// Verificar índices para optimizar concurrencia
	requiredIndexes := []string{
		"idx_asiento_vuelo_estado",
		"idx_reserva_pasajero_fecha",
		"idx_detalle_reserva_asiento",
		"idx_vuelo_fechas",
	}

	for _, index := range requiredIndexes {
		var indexExists bool
		err := db.QueryRow(
			"SELECT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = $1)",
			index,
		).Scan(&indexExists)
		
		if err != nil {
			return fmt.Errorf("error verificando índice %s: %v", index, err)
		}
		
		if !indexExists {
			return fmt.Errorf("índice requerido no encontrado: %s", index)
		}
	}

	return nil
}

func ResetTestData() error {
	db := GetDB()
	
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error iniciando transacción: %v", err)
	}
	defer tx.Rollback()

	// Limpiar tablas en orden correcto por dependencias
	_, err = tx.Exec(`
		TRUNCATE TABLE 
			auditoriareservas, detallereserva, reserva, asiento, 
			vuelo, avion, modeloavion, aeropuerto, claseservicio, 
			pasajero, metodopago 
		RESTART IDENTITY CASCADE
	`)
	if err != nil {
		return fmt.Errorf("error truncando tablas: %v", err)
	}

	// Insertar datos de prueba básicos
	_, err = tx.Exec(`
		INSERT INTO modeloavion (nombre, capacidad_economica, capacidad_business, filas_economica, filas_business, asientos_por_fila) 
		VALUES ('Boeing 737-800', 150, 20, 25, 5, 6)
	`)
	if err != nil {
		return fmt.Errorf("error insertando modelo avión: %v", err)
	}

	_, err = tx.Exec(`
		INSERT INTO avion (modelo_id, matricula, estado) 
		VALUES (1, 'N12345', 'operativo')
	`)
	if err != nil {
		return fmt.Errorf("error insertando avión: %v", err)
	}

	// Insertar aeropuertos
	_, err = tx.Exec(`
		INSERT INTO aeropuerto (codigo_iata, nombre, ciudad, pais) 
		VALUES 
			('GUA', 'La Aurora', 'Guatemala City', 'Guatemala'),
			('MIA', 'Miami International Airport', 'Miami', 'Estados Unidos')
	`)
	if err != nil {
		return fmt.Errorf("error insertando aeropuertos: %v", err)
	}

	// Insertar vuelo de prueba
	_, err = tx.Exec(`
		INSERT INTO vuelo (numero_vuelo, avion_id, aerolinea, origen_id, destino_id, fecha_hora_salida, fecha_hora_llegada, estado) 
		VALUES ('AV101', 1, 'Avianca', 1, 2, NOW() + INTERVAL '1 day', NOW() + INTERVAL '1 day 3 hours', 'programado')
	`)
	if err != nil {
		return fmt.Errorf("error insertando vuelo: %v", err)
	}

	// Insertar clases de servicio
	_, err = tx.Exec(`
		INSERT INTO claseservicio (nombre, multiplicador_precio) 
		VALUES 
			('Económica', 1.0),
			('Business', 2.5)
	`)
	if err != nil {
		return fmt.Errorf("error insertando clases de servicio: %v", err)
	}

	// Insertar 30 asientos de prueba (5 business, 25 económica)
	for i := 1; i <= 30; i++ {
		clase := 1 // Económica
		if i <= 5 {
			clase = 2 // Business
		}

		// Calcular código de asiento (ej: 1A, 1B, ..., 2A, etc.)
		fila := (i + 5) / 6
		letra := 'A' + (i-1)%6
		codigo := fmt.Sprintf("%d%c", fila, letra)

		// Determinar ubicación
		var ubicacion string
		switch (i-1) % 6 {
		case 0, 5:
			ubicacion = "ventana"
		case 2, 3:
			ubicacion = "centro"
		default:
			ubicacion = "pasillo"
		}

		_, err = tx.Exec(`
			INSERT INTO asiento (vuelo_id, codigo_asiento, clase_servicio_id, ubicacion, estado, precio_base) 
			VALUES (1, $1, $2, $3, 'disponible', $4)
		`, codigo, clase, ubicacion, 100.0*float64(clase))
		if err != nil {
			return fmt.Errorf("error insertando asiento %d: %v", i, err)
		}
	}

	// Insertar 30 pasajeros de prueba
	for i := 1; i <= 30; i++ {
		_, err = tx.Exec(`
			INSERT INTO pasajero (nombre, apellido, tipo_documento, numero_documento) 
			VALUES ($1, $2, 'pasaporte', $3)
		`, fmt.Sprintf("Pasajero%d", i), fmt.Sprintf("Apellido%d", i), fmt.Sprintf("PAS%04d", i))
		if err != nil {
			return fmt.Errorf("error insertando pasajero %d: %v", i, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error confirmando transacción: %v", err)
	}

	return nil
}

// CheckSeatAvailability verifica si un asiento está disponible (para pruebas)
func CheckSeatAvailability(seatID int) (bool, error) {
	db := GetDB()
	
	var estado string
	err := db.QueryRow(
		"SELECT estado FROM asiento WHERE id = $1",
		seatID,
	).Scan(&estado)
	
	if err != nil {
		return false, fmt.Errorf("error verificando disponibilidad: %v", err)
	}
	
	return estado == "disponible", nil
}

// GetSeatDetails obtiene información detallada de un asiento (para análisis)
func GetSeatDetails(seatID int) (map[string]interface{}, error) {
	db := GetDB()
	
	details := make(map[string]interface{})
	
	var codigo, clase, ubicacion, estado string
	var precio float64
	var vueloID int
	
	err := db.QueryRow(`
		SELECT a.codigo_asiento, a.ubicacion, a.estado, a.precio_base, a.vuelo_id, cs.nombre as clase
		FROM asiento a
		JOIN claseservicio cs ON a.clase_servicio_id = cs.id
		WHERE a.id = $1
	`, seatID).Scan(&codigo, &ubicacion, &estado, &precio, &vueloID, &clase)
	
	if err != nil {
		return nil, fmt.Errorf("error obteniendo detalles del asiento: %v", err)
	}
	
	details["codigo"] = codigo
	details["clase"] = clase
	details["ubicacion"] = ubicacion
	details["estado"] = estado
	details["precio"] = precio
	details["vuelo_id"] = vueloID
	
	return details, nil
}