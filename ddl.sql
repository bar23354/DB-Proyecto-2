CREATE TABLE Vuelo (
    ID SERIAL PRIMARY KEY,
    numero_vuelo VARCHAR(10) NOT NULL,
    avion VARCHAR(20) NOT NULL,
    aerolinea VARCHAR(50),
    origen VARCHAR(50),
    destino VARCHAR(50),
    fecha_hora TIMESTAMP NOT NULL
);

CREATE TABLE Asiento (
    ID SERIAL PRIMARY KEY,
    vuelo_id INTEGER NOT NULL REFERENCES Vuelo(ID),
    fila INTEGER NOT NULL,
    numero INTEGER NOT NULL,
    clase VARCHAR(20) CHECK (clase IN ('economica', 'business')),
    ubicacion VARCHAR(20) CHECK (ubicacion IN ('ventana', 'pasillo')),
    estado VARCHAR(10) DEFAULT 'libre' CHECK (estado IN ('libre', 'reservado')),
    UNIQUE (vuelo_id, fila, numero)
);

CREATE TABLE Reserva (
    ID SERIAL PRIMARY KEY,
    asiento_id INTEGER NOT NULL REFERENCES Asiento(ID),
    pasajero_id INTEGER REFERENCES Pasajero(ID),
    fecha_reserva TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    estado VARCHAR(10) CHECK (estado IN ('éxito', 'fallido'))
);

CREATE TABLE Pasajero (
    ID SERIAL PRIMARY KEY,
    nombre VARCHAR(100) NOT NULL,
    pasaporte VARCHAR(20) UNIQUE,
    email VARCHAR(100)
);