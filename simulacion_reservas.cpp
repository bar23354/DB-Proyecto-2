#include <iostream>
#include <vector>
#include <string>
#include <pqxx/pqxx>
#include <pthread.h>
#include <cstdlib>
#include <ctime>
#include <unistd.h>
#include <libpqxx/pqxx> 

using namespace std;
using namespace pqxx;

// Configuración de la base de datos
const string DB_CONNECTION = "dbname=airline user=postgres password=postgres";

// Estructura para pasar parámetros a los hilos
struct ThreadParams {
    int thread_id;
    int asiento_id;
    int pasajero_id;
    string isolation_level;
};

// Función para conectar a la base de datos
connection* connect_db() {
    try {
        return new connection(DB_CONNECTION);
    } catch (const exception &e) {
        cerr << "Error al conectar a la base de datos: " << e.what() << endl;
        return nullptr;
    }
}

// Función que ejecuta cada hilo para intentar reservar un asiento
void* reserve_seat(void* params) {
    ThreadParams* p = (ThreadParams*)params;
    
    try {
        connection* C = connect_db();
        if (!C) {
            cerr << "Hilo " << p->thread_id << ": Error de conexión" << endl;
            return nullptr;
        }

        // Configurar el nivel de aislamiento
        string isolation_sql;
        if (p->isolation_level == "READ COMMITTED") {
            isolation_sql = "SET TRANSACTION ISOLATION LEVEL READ COMMITTED;";
        } else if (p->isolation_level == "REPEATABLE READ") {
            isolation_sql = "SET TRANSACTION ISOLATION LEVEL REPEATABLE READ;";
        } else { // SERIALIZABLE
            isolation_sql = "SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;";
        }

        // Iniciar transacción
        work W(*C);
        C->exec(isolation_sql);

        // Verificar estado del asiento
        string check_sql = "SELECT estado FROM Asiento WHERE ID = " + to_string(p->asiento_id) + ";";
        result R = W.exec(check_sql);
        
        if (R.empty()) {
            cerr << "Hilo " << p->thread_id << ": Asiento no encontrado" << endl;
            W.abort();
            delete C;
            return nullptr;
        }

        string estado = R[0][0].as<string>();
        
        if (estado == "libre") {
            // Intentar reservar
            string update_sql = "UPDATE Asiento SET estado = 'reservado' WHERE ID = " + to_string(p->asiento_id) + ";";
            W.exec(update_sql);
            
            string insert_sql = "INSERT INTO Reserva (asiento_id, pasajero_id, estado) VALUES (" + 
                                to_string(p->asiento_id) + ", " + to_string(p->pasajero_id) + ", 'éxito');";
            W.exec(insert_sql);
            
            W.commit();
            cout << "Hilo " << p->thread_id << ": Reserva exitosa para asiento " << p->asiento_id << endl;
        } else {
            // Asiento ya reservado
            string insert_sql = "INSERT INTO Reserva (asiento_id, pasajero_id, estado) VALUES (" + 
                                to_string(p->asiento_id) + ", " + to_string(p->pasajero_id) + ", 'fallido');";
            W.exec(insert_sql);
            
            W.commit();
            cout << "Hilo " << p->thread_id << ": Reserva fallida - asiento " << p->asiento_id << " ya ocupado" << endl;
        }
        
        delete C;
    } catch (const exception &e) {
        cerr << "Hilo " << p->thread_id << ": Error en transacción - " << e.what() << endl;
    }
    
    return nullptr;
}

// Función principal
int main(int argc, char* argv[]) {
    if (argc != 4) {
        cerr << "Uso: " << argv[0] << " <num_usuarios> <asiento_id> <isolation_level>" << endl;
        cerr << "Niveles de aislamiento: READ COMMITTED, REPEATABLE READ, SERIALIZABLE" << endl;
        return 1;
    }

    int num_usuarios = atoi(argv[1]);
    int asiento_id = atoi(argv[2]);
    string isolation_level = argv[3];
    
    // Validar nivel de aislamiento
    if (isolation_level != "READ COMMITTED" && isolation_level != "REPEATABLE READ" && isolation_level != "SERIALIZABLE") {
        cerr << "Nivel de aislamiento no válido. Opciones: READ COMMITTED, REPEATABLE READ, SERIALIZABLE" << endl;
        return 1;
    }

    cout << "Iniciando simulación con " << num_usuarios << " usuarios para asiento " << asiento_id 
         << " con nivel de aislamiento " << isolation_level << endl;

    // Crear hilos
    vector<pthread_t> threads(num_usuarios);
    vector<ThreadParams> params(num_usuarios);
    
    srand(time(nullptr));
    
    for (int i = 0; i < num_usuarios; ++i) {
        params[i].thread_id = i + 1;
        params[i].asiento_id = asiento_id;
        params[i].pasajero_id = (rand() % 3) + 1; // IDs de pasajero entre 1 y 3
        params[i].isolation_level = isolation_level;
        
        if (pthread_create(&threads[i], nullptr, reserve_seat, &params[i])) {
            cerr << "Error al crear hilo " << i << endl;
            return 1;
        }
        
        // Pequeña pausa para variar el timing
        usleep(rand() % 1000);
    }

    // Esperar a que todos los hilos terminen
    for (int i = 0; i < num_usuarios; ++i) {
        pthread_join(threads[i], nullptr);
    }

    cout << "Simulación completada" << endl;
    return 0;
}