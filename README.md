# Sistema de Reservas de Vuelos

## Requisitos
- Go
- PostgreSQL (pgAdmin)

## Instalación de la base de datos
Crea una base de datos en pgAdmin con el siguiente nombre
```
reservaciones_vuelos
```

El usuario por defendo en PostgreSQL es postgres, pero si tú tienes otro nombre de usuario, puedes cambiarlo en el .env

Pon la contraseña de tu usuario en el .env

En sql_scripts accede al archivo 00_ddl.sql, copia su contenido y pegalo y ejecutalo en el Query tool de pgAdmin, en la base de datos que acabas de crear.

## Ejecutar el código de simulación
En cmd/powershell accede a la ubicación de la carpeta reservation-simulator

Ejecuta el código con el siguiente comando:
```
go run .
```
