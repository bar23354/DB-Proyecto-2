FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY src/go.mod src/go.sum ./
RUN go mod download

# Copia todos los archivos fuente
COPY src/ ./

# Compila especificando todos los archivos .go necesarios
RUN CGO_ENABLED=0 GOOS=linux go build -o reservas-app main.go simulator.go database.go models.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/reservas-app .
COPY sql_scripts/ /sql_scripts/
CMD ["./reservas-app"]