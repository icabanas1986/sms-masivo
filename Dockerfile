FROM golang:1.23-alpine

WORKDIR /app

# Copia los archivos de dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copia el código fuente
COPY . .

# Compila la aplicación
RUN go build -o main .

# Expone el puerto
EXPOSE 8080

# Ejecuta la aplicación
CMD ["./main"]
