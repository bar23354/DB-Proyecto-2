FROM postgres:latest

RUN apt-get update && \
    apt-get install -y \
    build-essential \
    libpq-dev \
    postgresql-server-dev-all \
    g++ \
    libpqxx-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY . .

RUN g++ -std=c++17 simulacion_reservas.cpp -lpqxx -lpthread -o reservas

COPY init_db.sh /docker-entrypoint-initdb.d/
RUN chmod +x /docker-entrypoint-initdb.d/init_db.sh

CMD ["tail", "-f", "/dev/null"]