# Twilio SMS Backend

Backend en Go para envío masivo de SMS usando Twilio.

## Configuración

1. Copia .env.example a .env
2. Agrega tus credenciales de Twilio
3. Ejecuta: `go run main.go`

## Endpoints

- POST /send-sms - Enviar SMS individual
- POST /send-bulk-sms - Enviar SMS masivo
- GET /health - Health check

## Variables de Entorno

- TWILIO_ACCOUNT_SID - Tu Account SID de Twilio
- TWILIO_AUTH_TOKEN - Tu Auth Token de Twilio
- TWILIO_PHONE_NUMBER - Tu número de Twilio
- PORT - Puerto del servidor (default: 8080)
