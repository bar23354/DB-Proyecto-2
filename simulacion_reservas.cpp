#include <iostream>
#include <vector>
#include <string>
#include <pqxx/pqxx>
#include <pthread.h>
#include <cstdlib>
#include <ctime>
#include <unistd.h>

using namespace std;
using namespace pqxx;

// g++ -std=c++17 simulacion_reservas.cpp -lpqxx -lpthread -o reservas
// ./reservas 10 5 "SERIALIZABLE"
const string DB_CONNECTION = "dbname=airline user=postgres password=postgres host=db client_encoding=UTF8";

struct ThreadParams {
    int thread_id;
    int asiento_id;
    int pasajero_id;
    string isolation_level;
};

connection* connect_db() {
    try {
        return new connection(DB_CONNECTION);
    } catch (const exception &e) {
        cerr << "Error de conexión: " << e.what() << endl;
        return nullptr;
    }
}

void* reserve_seat(void* params) {
    ThreadParams* p = static_cast<ThreadParams*>(params);
    
    try {
        connection* conn = connect_db();
        if (!conn) {
            cerr << "Hilo " << p->thread_id << ": Error de conexión" << endl;
            return nullptr;
        }

        string isolation_sql = "SET TRANSACTION ISOLATION LEVEL " + p->isolation_level + ";";
        
        int max_retries = (p->isolation_level == "SERIALIZABLE") ? 3 : 1;
        bool success = false;

        for (int retry = 0; retry < max_retries; ++retry) {
            try {
                work tx(*conn);
                conn->exec(isolation_sql);

                // Verificar estado del asiento y reservar si está libre
                string check_sql = 
                    "SELECT estado FROM Asiento WHERE ID = " + to_string(p->asiento_id) + ";";
                result res_check = tx.exec(check_sql);
                
                if (!res_check.empty() && res_check[0][0].as<string>() == "libre") {
                    // Actualizar estado del asiento
                    string update_sql = 
                        "UPDATE Asiento SET estado = 'reservado' "
                        "WHERE ID = " + to_string(p->asiento_id) + ";";
                    tx.exec(update_sql);

                    // Crear reserva
                    string insert_sql = 
                        "INSERT INTO Reserva(asiento_id, pasajero_id, estado, fecha_reserva) "
                        "VALUES(" + to_string(p->asiento_id) + ", " 
                        + to_string(p->pasajero_id) + ", 'éxito', NOW());";
                    tx.exec(insert_sql);
                    
                    tx.commit();
                    cout << "Hilo " << p->thread_id << ": Reserva EXITOSA para asiento " 
                         << p->asiento_id << endl;
                    success = true;
                    break;
                } else {
                    // Registrar intento fallido
                    string insert_sql = 
                        "INSERT INTO Reserva(asiento_id, pasajero_id, estado, fecha_reserva) "
                        "VALUES(" + to_string(p->asiento_id) + ", " 
                        + to_string(p->pasajero_id) + ", 'fallido', NOW());";
                    tx.exec(insert_sql);
                    tx.commit();
                    cout << "Hilo " << p->thread_id << ": Reserva FALLIDA (asiento " 
                         << p->asiento_id << " ocupado)" << endl;
                    break;
                }
            } catch (const serialization_failure& e) {
                if (retry == max_retries - 1) throw;
                cerr << "Hilo " << p->thread_id << ": Reintento " << (retry + 1) << endl;
                usleep(100000); // 100ms de espera entre reintentos
            }
        }

        delete conn;
    } catch (const exception &e) {
        cerr << "Hilo " << p->thread_id << ": Error crítico - " << e.what() << endl;
    }
    
    return nullptr;
}

int main(int argc, char* argv[]) {
    if (argc != 4) {
        cerr << "Uso: " << argv[0] << " <num_usuarios> <asiento_id> <isolation_level>\n"
             << "Niveles de aislamiento válidos: READ COMMITTED, REPEATABLE READ, SERIALIZABLE" 
             << endl;
        return 1;
    }

    int num_usuarios = atoi(argv[1]);
    int asiento_id = atoi(argv[2]);
    string isolation_level = argv[3];
    
    const vector<string> valid_levels = {
        "READ COMMITTED", "REPEATABLE READ", "SERIALIZABLE"
    };
    
    if (find(valid_levels.begin(), valid_levels.end(), isolation_level) == valid_levels.end()) {
        cerr << "Error: Nivel de aislamiento no válido" << endl;
        return 1;
    }

    cout << "=== Simulación iniciada ===" << endl
         << "Usuarios: " << num_usuarios << endl
         << "Asiento objetivo: " << asiento_id << endl
         << "Nivel de aislamiento: " << isolation_level << "\n\n";

    vector<pthread_t> threads(num_usuarios);
    vector<ThreadParams> params(num_usuarios);
    
    srand(time(nullptr));
    
    for (int i = 0; i < num_usuarios; ++i) {
        params[i] = {
            .thread_id = i + 1,
            .asiento_id = asiento_id,
            .pasajero_id = (rand() % 3) + 1,  // IDs de pasajero entre 1 y 3
            .isolation_level = isolation_level
        };
        
        if (pthread_create(&threads[i], nullptr, reserve_seat, &params[i])) {
            cerr << "Error fatal: No se pudo crear el hilo " << i << endl;
            return 1;
        }
        
        usleep(rand() % 1000);  // Pequeña variación en el inicio de los hilos
    }

    for (auto& thread : threads) {
        pthread_join(thread, nullptr);
    }

    cout << "\n=== Simulación completada ===" << endl;
    return 0;
}