-- Insertar modelos de avión
INSERT INTO ModeloAvion (nombre, capacidad_economica, capacidad_business, filas_economica, filas_business, asientos_por_fila)
VALUES 
('Boeing 737-800', 162, 24, 27, 4, 6),
('Airbus A320', 150, 30, 25, 5, 6),
('Embraer E190', 94, 20, 18, 4, 4);

-- Insertar aviones
INSERT INTO Avion (modelo_id, matricula, estado)
VALUES
(1, 'N12345', 'operativo'),
(1, 'N67890', 'operativo'),
(2, 'A1B2C3', 'operativo'),
(3, 'E19001', 'operativo');

-- Insertar aeropuertos
INSERT INTO Aeropuerto (codigo_iata, nombre, ciudad, pais)
VALUES
('GUA', 'La Aurora', 'Ciudad de Guatemala', 'Guatemala'),
('MIA', 'Miami International', 'Miami', 'Estados Unidos'),
('JFK', 'John F. Kennedy', 'Nueva York', 'Estados Unidos'),
('SAL', 'El Salvador International', 'San Salvador', 'El Salvador');

-- Insertar clases de servicio
INSERT INTO ClaseServicio (nombre, multiplicador_precio)
VALUES
('Business', 2.5),
('Económica', 1.0);

-- Insertar vuelos
INSERT INTO Vuelo (numero_vuelo, avion_id, aerolinea, origen_id, destino_id, fecha_hora_salida, fecha_hora_llegada, estado, puerta_embarque)
VALUES
('AV101', 1, 'Avianca', 1, 2, '2024-06-01 08:00:00', '2024-06-01 12:30:00', 'programado', 'B12'),
('AV102', 2, 'Avianca', 2, 1, '2024-06-02 14:00:00', '2024-06-02 16:45:00', 'programado', 'D45'),
('TA202', 3, 'TACA', 4, 1, '2024-06-05 09:45:00', '2024-06-05 10:30:00', 'programado', 'A3');

-- Función para insertar asientos automáticamente según el modelo de avión
CREATE OR REPLACE FUNCTION insertar_asientos_vuelo() RETURNS void AS $$
DECLARE
    vuelo_record RECORD;
    modelo_record RECORD;
    fila INT;
    asiento_num INT;
    letra CHAR;
    clase_id INT;
    ubicacion TEXT;
    precio_base DECIMAL(10,2);
    codigo_asiento VARCHAR(5);
BEGIN
    FOR vuelo_record IN SELECT * FROM Vuelo LOOP
        SELECT * INTO modelo_record FROM ModeloAvion WHERE ID = (
            SELECT modelo_id FROM Avion WHERE ID = vuelo_record.avion_id
        );
        
        -- Asientos Business (primera clase)
        FOR fila IN 1..modelo_record.filas_business LOOP
            FOR asiento_num IN 1..modelo_record.asientos_por_fila LOOP
                letra := CHR(64 + asiento_num); -- A, B, C, etc.
                codigo_asiento := fila || letra;
                
                ubicacion := CASE 
                    WHEN asiento_num = 1 OR asiento_num = modelo_record.asientos_por_fila THEN 'ventana'
                    WHEN asiento_num = 2 OR asiento_num = modelo_record.asientos_por_fila-1 THEN 'pasillo'
                    ELSE 'centro' 
                END;
                
                INSERT INTO Asiento (vuelo_id, codigo_asiento, clase_servicio_id, ubicacion, estado, precio_base)
                VALUES (vuelo_record.ID, codigo_asiento, 1, ubicacion, 'disponible', 500.00);
            END LOOP;
        END LOOP;
        
        -- Asientos Económicos
        FOR fila IN (modelo_record.filas_business+1)..(modelo_record.filas_business+modelo_record.filas_economica) LOOP
            FOR asiento_num IN 1..modelo_record.asientos_por_fila LOOP
                letra := CHR(64 + asiento_num); -- A, B, C, etc.
                codigo_asiento := fila || letra;
                
                ubicacion := CASE 
                    WHEN asiento_num = 1 OR asiento_num = modelo_record.asientos_por_fila THEN 'ventana'
                    WHEN asiento_num = 2 OR asiento_num = modelo_record.asientos_por_fila-1 THEN 'pasillo'
                    ELSE 'centro' 
                END;
                
                -- Filas con más espacio (plus)
                IF fila <= modelo_record.filas_business + 5 THEN
                    precio_base := 300.00;
                ELSE
                    precio_base := 200.00;
                END IF;
                
                INSERT INTO Asiento (vuelo_id, codigo_asiento, clase_servicio_id, ubicacion, estado, precio_base)
                VALUES (vuelo_record.ID, codigo_asiento, 2, ubicacion, 'disponible', precio_base);
            END LOOP;
        END LOOP;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Ejecutar la función para insertar asientos
SELECT insertar_asientos_vuelo();

-- Insertar pasajeros
INSERT INTO Pasajero (nombre, apellido, tipo_documento, numero_documento, email, telefono, programa_millas)
VALUES
('Juan', 'Pérez', 'pasaporte', 'P12345678', 'juan.perez@email.com', '+502 12345678', 'Avianca LifeMiles'),
('María', 'González', 'pasaporte', 'G87654321', 'maria.gonzalez@email.com', '+52 55 98765432', NULL),
('Ana', 'Martínez', 'dpi', '2876543210101', 'ana.martinez@email.com', '+502 87654321', NULL),
('Carlos', 'Ramírez', 'pasaporte', 'P98765433', 'carlos.ramirez@email.com', '+502 23456789', 'Aeromexico Club Premier'),
('Luisa', 'Hernández', 'dpi', '1234567890102', 'luisa.hernandez@email.com', '+502 34567890', 'Copa ConnectMiles'),
('Pedro', 'López', 'pasaporte', 'P11223345', 'pedro.lopez@email.com', '+505 12345678', 'American Airlines AAdvantage'),
('Sofía', 'Gutiérrez', 'dpi', '4556789100102', 'sofia.gutierrez@email.com', '+502 23456781', NULL),
('Ricardo', 'Martínez', 'pasaporte', 'P22334456', 'ricardo.martinez@email.com', '+503 34567890', 'Delta SkyMiles'),
('Verónica', 'Sánchez', 'dpi', '1230987654322', 'veronica.sanchez@email.com', '+502 45678901', NULL),
('Juanita', 'Flores', 'pasaporte', 'P66778890', 'juanita.flores@email.com', '+503 56789012', 'United MileagePlus'),
('Eduardo', 'Torres', 'dpi', '5566778899002', 'eduardo.torres@email.com', '+502 67890123', 'Iberia Plus'),
('Gabriela', 'Vargas', 'pasaporte', 'P99887767', 'gabriela.vargas@email.com', '+502 78901234', 'LATAM Pass'),
('Raúl', 'Jiménez', 'dpi', '1122334455668', 'raul.jimenez@email.com', '+504 89012345', 'Aerolineas Argentinas Plus'),
('Carmen', 'Delgado', 'pasaporte', 'P55664434', 'carmen.delgado@email.com', '+502 11112222', 'Volaris v.club'),
('Fernando', 'Mejía', 'dpi', '3216549870124', 'fernando.mejia@email.com', '+502 33334444', NULL),
('Laura', 'Pineda', 'pasaporte', 'P99882212', 'laura.pineda@email.com', '+505 55556666', 'Sky Airline Cloud'),
('Diego', 'Alvarado', 'dpi', '7894561237891', 'diego.alvarado@email.com', '+506 77778888', NULL),
('Tatiana', 'Ríos', 'pasaporte', 'P44556678', 'tatiana.rios@email.com', '+502 99990000', 'Avianca LifeMiles'),
('Marcelo', 'Ortiz', 'dpi', '1203984756013', 'marcelo.ortiz@email.com', '+502 12344321', NULL),
('Patricia', 'Salazar', 'pasaporte', 'P77889901', 'patricia.salazar@email.com', '+503 65432100', 'Delta SkyMiles'),
('Jorge', 'Castro', 'dpi', '3012457698124', 'jorge.castro@email.com', '+504 56789012', NULL),
('Isabel', 'Fernández', 'pasaporte', 'P33445567', 'isabel.fernandez@email.com', '+502 87654321', 'American Airlines AAdvantage'),
('Esteban', 'Morales', 'dpi', '9876543214568', 'esteban.morales@email.com', '+502 43211234', NULL),
('Andrea', 'Navarro', 'pasaporte', 'P11224455', 'andrea.navarro@email.com', '+502 10101010', 'Aeromexico Club Premier'),
('Santiago', 'Peralta', 'dpi', '1298765432101', 'santiago.peralta@email.com', '+502 20202020', NULL),
('Natalia', 'Córdoba', 'pasaporte', 'P66778822', 'natalia.cordoba@email.com', '+502 30303030', 'Viva Air Miles'),
('Tomás', 'González', 'dpi', '9876543210123', 'tomas.gonzalez@email.com', '+502 40404040', 'Sky Airline Cloud'),
('Camila', 'Herrera', 'pasaporte', 'P55667788', 'camila.herrera@email.com', '+502 50505050', 'LATAM Pass'),
('Emilio', 'Reyes', 'dpi', '1234432198765', 'emilio.reyes@email.com', '+502 60606060', NULL),
('Ana', 'Domínguez', 'pasaporte', 'P33446677', 'ana.dominguez@email.com', '+502 70707070', 'Copa ConnectMiles'),
('Julián', 'Valdez', 'dpi', '3217896543210', 'julian.valdez@email.com', '+502 80808080', 'Iberia Plus'),
('Paola', 'Mendoza', 'pasaporte', 'P77882233', 'paola.mendoza@email.com', '+502 90909090', 'Delta SkyMiles'),
('Mateo', 'Luna', 'dpi', '4567891234567', 'mateo.luna@email.com', '+502 00001111', NULL);

-- Insertar métodos de pago
INSERT INTO MetodoPago (tipo, ultimos_digitos, pasajero_id)
VALUES
('tarjeta', '1234', 1),
('tarjeta', '5678', 2),
('efectivo', NULL, 3);

-- Insertar reservas de prueba
INSERT INTO Reserva (codigo_reserva, pasajero_id, estado, metodo_pago_id, total_pago, session_id)
VALUES
('RES001', 1, 'confirmada', 1, 500.00, 'sess123'),
('RES002', 2, 'confirmada', 2, 600.00, 'sess456'),
('RES003', 3, 'pendiente', 3, 200.00, 'sess789');

-- Insertar detalles de reserva (simulando conflictos de concurrencia)
INSERT INTO DetalleReserva (reserva_id, asiento_id, precio_final)
VALUES
(1, 5, 500.00), -- Business
(2, 18, 300.00), -- Económica plus
(2, 19, 300.00), -- Económica plus
(3, 42, 200.00); -- Económica

-- Actualizar estado de asientos reservados
UPDATE Asiento SET estado = 'reservado' WHERE ID IN (5, 18, 19, 42);

-- Insertar auditoría de cambios
INSERT INTO AuditoriaReservas (asiento_id, estado_anterior, estado_nuevo, reserva_id, usuario)
VALUES
(5, 'disponible', 'reservado', 1, 'sistema'),
(18, 'disponible', 'reservado', 2, 'sistema'),
(19, 'disponible', 'reservado', 2, 'sistema'),
(42, 'disponible', 'reservado', 3, 'sistema');

-- Insertar promociones
INSERT INTO Promocion (codigo, descuento, fecha_inicio, fecha_fin)
VALUES
('VERANO24', 15.00, '2024-06-01', '2024-08-31'),
('MILLAS15', 15.00, '2024-05-01', '2024-12-31');