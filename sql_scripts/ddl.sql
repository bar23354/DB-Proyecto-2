-- Tabla de modelos de avión (simplificada)
CREATE TABLE ModeloAvion (
    ID SERIAL PRIMARY KEY,
    nombre VARCHAR(50) NOT NULL,
    capacidad_economica INT NOT NULL,
    capacidad_business INT NOT NULL,
    filas_economica INT NOT NULL,
    filas_business INT NOT NULL,
    asientos_por_fila INT NOT NULL
);

-- Tabla de aviones (simplificada)
CREATE TABLE Avion (
    ID SERIAL PRIMARY KEY,
    modelo_id INTEGER NOT NULL REFERENCES ModeloAvion(ID),
    matricula VARCHAR(20) UNIQUE NOT NULL,
    estado VARCHAR(20) DEFAULT 'operativo' CHECK (estado IN ('operativo', 'mantenimiento', 'inactivo'))
);

-- Tabla de aeropuertos (esencial)
CREATE TABLE Aeropuerto (
    ID SERIAL PRIMARY KEY,
    codigo_iata VARCHAR(3) UNIQUE NOT NULL,
    nombre VARCHAR(100) NOT NULL,
    ciudad VARCHAR(50) NOT NULL,
    pais VARCHAR(50) NOT NULL
);

-- Tabla de vuelos (con validaciones mejoradas)
CREATE TABLE Vuelo (
    ID SERIAL PRIMARY KEY,
    numero_vuelo VARCHAR(10) NOT NULL,
    avion_id INTEGER NOT NULL REFERENCES Avion(ID),
    aerolinea VARCHAR(50) NOT NULL,
    origen_id INTEGER NOT NULL REFERENCES Aeropuerto(ID),
    destino_id INTEGER NOT NULL REFERENCES Aeropuerto(ID),
    fecha_hora_salida TIMESTAMP NOT NULL,
    fecha_hora_llegada TIMESTAMP NOT NULL,
    estado VARCHAR(20) DEFAULT 'programado' CHECK (estado IN ('programado', 'abordando', 'en vuelo', 'aterrizado', 'cancelado', 'retrasado')),
    puerta_embarque VARCHAR(10),
    CHECK (destino_id != origen_id),
    CHECK (fecha_hora_llegada > fecha_hora_salida)
);

-- Tabla de clases de servicio (simplificada)
CREATE TABLE ClaseServicio (
    ID SERIAL PRIMARY KEY,
    nombre VARCHAR(20) NOT NULL UNIQUE,
    multiplicador_precio DECIMAL(3,2) DEFAULT 1.0
);

-- Tabla de asientos (optimizada para concurrencia)
CREATE TABLE Asiento (
    ID SERIAL PRIMARY KEY,
    vuelo_id INTEGER NOT NULL REFERENCES Vuelo(ID),
    codigo_asiento VARCHAR(5) NOT NULL, -- Ej: "12A", "3B"
    clase_servicio_id INTEGER NOT NULL REFERENCES ClaseServicio(ID),
    ubicacion VARCHAR(10) CHECK (ubicacion IN ('ventana', 'pasillo', 'centro')),
    estado VARCHAR(20) DEFAULT 'disponible' CHECK (estado IN ('disponible', 'reservado', 'ocupado', 'no disponible')),
    precio_base DECIMAL(10,2) NOT NULL,
    UNIQUE (vuelo_id, codigo_asiento)
);

-- Tabla de pasajeros (esencial)
CREATE TABLE Pasajero (
    ID SERIAL PRIMARY KEY,
    nombre VARCHAR(100) NOT NULL,
    apellido VARCHAR(100) NOT NULL,
    tipo_documento VARCHAR(20) CHECK (tipo_documento IN ('pasaporte', 'dpi', 'licencia')),
    numero_documento VARCHAR(20) UNIQUE NOT NULL,
    email VARCHAR(100),
    telefono VARCHAR(20),
    programa_millas VARCHAR(50)
);

-- Tabla de métodos de pago (simplificada)
CREATE TABLE MetodoPago (
    ID SERIAL PRIMARY KEY,
    tipo VARCHAR(20) CHECK (tipo IN ('tarjeta', 'transferencia', 'efectivo', 'millas')),
    ultimos_digitos VARCHAR(4),
    pasajero_id INTEGER REFERENCES Pasajero(ID)
);

-- Tabla de reservaciones (con campos para concurrencia)
CREATE TABLE Reserva (
    ID SERIAL PRIMARY KEY,
    codigo_reserva VARCHAR(50) UNIQUE NOT NULL,
    pasajero_id INTEGER NOT NULL REFERENCES Pasajero(ID),
    fecha_reserva TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    estado VARCHAR(20) CHECK (estado IN ('pendiente', 'confirmada', 'cancelada', 'completada')),
    metodo_pago_id INTEGER REFERENCES MetodoPago(ID),
    total_pago DECIMAL(10,2) NOT NULL,
    session_id VARCHAR(100) -- Para manejo de concurrencia
);

-- Tabla de detalles de reservación (transaccional)
CREATE TABLE DetalleReserva (
    ID SERIAL PRIMARY KEY,
    reserva_id INTEGER NOT NULL REFERENCES Reserva(ID),
    asiento_id INTEGER NOT NULL REFERENCES Asiento(ID),
    precio_final DECIMAL(10,2) NOT NULL,
    UNIQUE (asiento_id, reserva_id)
);

-- Tabla de promociones (opcional)
CREATE TABLE Promocion (
    ID SERIAL PRIMARY KEY,
    codigo VARCHAR(20) UNIQUE NOT NULL,
    descuento DECIMAL(5,2) NOT NULL,
    fecha_inicio DATE NOT NULL,
    fecha_fin DATE NOT NULL
);

-- Tabla para auditoría de cambios (importante para concurrencia)
CREATE TABLE AuditoriaReservas (
    ID SERIAL PRIMARY KEY,
    fecha_hora TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    asiento_id INTEGER NOT NULL REFERENCES Asiento(ID),
    estado_anterior VARCHAR(20) NOT NULL,
    estado_nuevo VARCHAR(20) NOT NULL,
    reserva_id INTEGER REFERENCES Reserva(ID),
    usuario VARCHAR(50)
);

-- Índices para mejorar el rendimiento en operaciones concurrentes
CREATE INDEX idx_asiento_vuelo_estado ON Asiento(vuelo_id, estado);
CREATE INDEX idx_reserva_pasajero_fecha ON Reserva(pasajero_id, fecha_reserva);
CREATE INDEX idx_detalle_reserva_asiento ON DetalleReserva(asiento_id);
CREATE INDEX idx_vuelo_fechas ON Vuelo(fecha_hora_salida, fecha_hora_llegada);