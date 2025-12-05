package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

// Configuración de Twilio
type TwilioConfig struct {
	AccountSID string
	AuthToken  string
	FromPhone  string
}

// Estructura para el request de envío
type SMSRequest struct {
	To      []string `json:"to"`      // Lista de números destino
	Message string   `json:"message"` // Mensaje a enviar
}

// Estructura para la respuesta
type SMSResponse struct {
	Sent   []string `json:"sent"`
	Failed []string `json:"failed"`
	Total  int      `json:"total"`
}

var twilioConfig TwilioConfig
var twilioClient *twilio.RestClient

func main() {
	// Carga las variables de entorno desde .env
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  No se encontró archivo .env, usando variables de entorno del sistema")
	}

	// Configura Twilio desde variables de entorno
	twilioConfig = TwilioConfig{
		AccountSID: os.Getenv("TWILIO_ACCOUNT_SID"),
		AuthToken:  os.Getenv("TWILIO_AUTH_TOKEN"),
		FromPhone:  os.Getenv("TWILIO_PHONE_NUMBER"),
	}

	// Valida que las credenciales estén configuradas
	if twilioConfig.AccountSID == "" || twilioConfig.AuthToken == "" || twilioConfig.FromPhone == "" {
		log.Fatal("❌ Error: Configura TWILIO_ACCOUNT_SID, TWILIO_AUTH_TOKEN y TWILIO_PHONE_NUMBER en tu archivo .env")
	}

	// Inicializa el cliente de Twilio
	twilioClient = twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: twilioConfig.AccountSID,
		Password: twilioConfig.AuthToken,
	})

	log.Println("✅ Configuración de Twilio cargada correctamente")

	// Rutas
	http.HandleFunc("/send-sms", sendSMSHandler)
	http.HandleFunc("/send-bulk-sms", sendBulkSMSHandler)
	http.HandleFunc("/health", healthCheckHandler)

	// Puerto del servidor (también configurable desde .env)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Inicia el servidor
	fmt.Printf("🚀 Servidor corriendo en http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// Handler para enviar SMS a un solo destinatario
func sendSMSHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		To      string `json:"to"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Request inválido", http.StatusBadRequest)
		return
	}

	err := sendSMS(req.To, req.Message)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error al enviar SMS: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "SMS enviado correctamente",
		"to":      req.To,
	})
}

// Handler para envío masivo de SMS
func sendBulkSMSHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req SMSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Request inválido", http.StatusBadRequest)
		return
	}

	// Envía SMS de forma concurrente
	response := sendBulkSMS(req.To, req.Message)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Función para enviar un SMS individual usando el SDK de Twilio
func sendSMS(to, message string) error {
	params := &openapi.CreateMessageParams{}
	params.SetTo(to)
	params.SetFrom(twilioConfig.FromPhone)
	params.SetBody(message)

	_, err := twilioClient.Api.CreateMessage(params)
	if err != nil {
		return fmt.Errorf("error al enviar SMS: %v", err)
	}

	return nil
}

// Función para envío masivo con concurrencia
func sendBulkSMS(recipients []string, message string) SMSResponse {
	var wg sync.WaitGroup
	var mu sync.Mutex

	response := SMSResponse{
		Sent:   []string{},
		Failed: []string{},
		Total:  len(recipients),
	}

	// Limita la concurrencia a 10 goroutines simultáneas
	semaphore := make(chan struct{}, 10)

	for _, recipient := range recipients {
		wg.Add(1)
		go func(to string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Adquiere
			defer func() { <-semaphore }() // Libera

			err := sendSMS(to, message)

			mu.Lock()
			if err != nil {
				response.Failed = append(response.Failed, to)
				log.Printf("❌ Error enviando a %s: %v", to, err)
			} else {
				response.Sent = append(response.Sent, to)
				log.Printf("✅ SMS enviado a %s", to)
			}
			mu.Unlock()
		}(recipient)
	}

	wg.Wait()
	return response
}

// Health check
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "twilio-sms-service",
	})
}
