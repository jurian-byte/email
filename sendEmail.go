package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"os"

	"github.com/joho/godotenv"
	"github.com/rs/cors"
	mail "github.com/xhit/go-simple-mail/v2"
)

type EmailRequest struct {
	Name    string `json:"name"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

var (
	host           = "smtp-mail.outlook.com"
	port           = 587
	username       = ""
	password       = ""
	encryptionType = mail.EncryptionSTARTTLS
	connectTimeout = 10 * time.Second
	sendTimeout    = 10 * time.Second
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	username = os.Getenv("USERNAME_")
	password = os.Getenv("PASSWORD_")
	if username == "" || password == "" {
		log.Fatal("Error loading username or password from environment variables")
	}

	c := cors.AllowAll()

	router := http.NewServeMux()
	router.HandleFunc("/send-email", handleSendEmail)

	handler := c.Handler(router)
	log.Fatal(http.ListenAndServe(":8080", handler))

}

func handleSendEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	var emailRequest EmailRequest
	err := json.NewDecoder(r.Body).Decode(&emailRequest)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Error decoding request body: %v", err)})
		return
	}

	client := mail.NewSMTPClient()
	client.Host = host
	client.Port = port
	client.Username = username
	client.Password = password
	client.Encryption = encryptionType
	client.ConnectTimeout = connectTimeout
	client.SendTimeout = sendTimeout
	client.Authentication = mail.AuthLogin

	smtpClient, err := client.Connect()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Error connecting to SMTP server: %v", err)})
		return
	}

	email := mail.NewMSG()
	email.SetFrom(username).
		AddTo(username).
		SetSubject(emailRequest.Subject).
		SetBody(mail.TextHTML, fmt.Sprintf("Nombre: %s<br>Message: %s", emailRequest.Name, emailRequest.Body))

	err = email.Send(smtpClient)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Error sending email: %v", err)})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Email sent successfully"})
}
