# Simulación de reservas de vuelos

## Requisitos
- Go  
- PostgreSQL (pgAdmin)

## Instalación de la base de datos  
Crea una base de datos en pgAdmin con el siguiente nombre  
```
reservaciones_vuelos
```

El usuario por defecto en PostgreSQL es postgres, pero si tú tienes otro nombre de usuario, puedes cambiarlo en el .env

Pon la contraseña de tu usuario en el .env

En sql_scripts accede al archivo ddl.sql, copia su contenido, pégalo y ejecútalo en el Query tool de pgAdmin, en la base de datos que acabas de crear.

Después, haz lo mismo con el archivo data.sql para insertar los datos necesarios.

## Ejecutar el código de simulación  
En la terminal accede a la ubicación de la carpeta reservation-simulator

Ejecuta el código con el siguiente comando:  
```
go run .
```

## Ver los datos de simulación  
En la carpeta reservation-simulator se creará un archivo csv en el cual están los datos generados por la simulación.


