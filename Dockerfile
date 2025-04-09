# Use an official PostgreSQL image as the base
FROM postgres:latest

# Install build dependencies
RUN apt-get update && \
    apt-get install -y \
    build-essential \
    libpq-dev \
    postgresql-server-dev-all \
    g++ \
    libpqxx-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy the application files
WORKDIR /app
COPY ddl.sql .
COPY data.sql .
COPY simulacion_reservas.cpp .

# Compile the application
RUN g++ -std=c++17 simulacion_reservas.cpp -lpqxx -lpthread -o reservas

# Copy the initialization script
COPY init_db.sh /docker-entrypoint-initdb.d/

# Make the script executable
RUN chmod +x /docker-entrypoint-initdb.d/init_db.sh

# Command to keep the container running
CMD ["tail", "-f", "/dev/null"]