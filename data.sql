INSERT INTO Vuelo (numero_vuelo, aerolinea, origen, destino, fecha_hora)
VALUES (
    'AA123', 
    'American Airlines', 
    'Nueva York', 
    'Los Ángeles', 
    '2025-06-01 08:00:00'
);

INSERT INTO Asiento (vuelo_id, fila, numero, clase, ubicacion, estado) 
VALUES
(1, 1, 1, 'economica', 'ventana', 'libre'),
(1, 1, 2, 'economica', 'pasillo', 'libre'),
(1, 1, 3, 'economica', 'ventana', 'libre'),
(1, 1, 4, 'economica', 'pasillo', 'libre'),
(1, 1, 5, 'economica', 'ventana', 'libre'),
(1, 1, 6, 'economica', 'pasillo', 'libre'),

(1, 2, 1, 'economica', 'ventana', 'libre'),
(1, 2, 2, 'economica', 'pasillo', 'libre'),
(1, 2, 3, 'economica', 'ventana', 'libre'),
(1, 2, 4, 'economica', 'pasillo', 'libre'),
(1, 2, 5, 'economica', 'ventana', 'libre'),
(1, 2, 6, 'economica', 'pasillo', 'libre'),

(1, 3, 1, 'economica', 'ventana', 'libre'),
(1, 3, 2, 'economica', 'pasillo', 'libre'),
(1, 3, 3, 'economica', 'ventana', 'libre'),
(1, 3, 4, 'economica', 'pasillo', 'libre'),
(1, 3, 5, 'economica', 'ventana', 'libre'),
(1, 3, 6, 'economica', 'pasillo', 'libre'),

(1, 4, 1, 'economica', 'ventana', 'libre'),
(1, 4, 2, 'economica', 'pasillo', 'libre'),
(1, 4, 3, 'economica', 'ventana', 'libre'),
(1, 4, 4, 'economica', 'pasillo', 'libre'),
(1, 4, 5, 'economica', 'ventana', 'libre'),
(1, 4, 6, 'economica', 'pasillo', 'libre'),

(1, 5, 1, 'business', 'ventana', 'libre'),
(1, 5, 2, 'business', 'pasillo', 'libre'),
(1, 5, 3, 'business', 'ventana', 'libre'),
(1, 5, 4, 'business', 'pasillo', 'libre'),
(1, 5, 5, 'business', 'ventana', 'libre'),
(1, 5, 6, 'business', 'pasillo', 'libre');

INSERT INTO Pasajero (nombre, pasaporte, email)
VALUES
('Ana Torres', 'T112233', 'ana@example.com'),
('Luis Ramírez', 'R445566', 'luis@example.com'),
('Marta Sánchez', 'S778899', 'marta@example.com');

INSERT INTO Reserva (asiento_id, pasajero_id, estado, fecha_reserva)
VALUES
(5, 1, 'éxito', '2025-05-20 09:00:00'),
(12, 2, 'éxito', '2025-05-20 09:15:00'),
(5, 3, 'fallido', '2025-05-20 09:30:00');

UPDATE Asiento SET estado = 'reservado' WHERE ID IN (5, 12);