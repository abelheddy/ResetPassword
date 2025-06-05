package email

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"net"
	"net/smtp"
	"time"
)

type EmailService struct {
	db *sql.DB
}

func NewEmailService(db *sql.DB) *EmailService {
	return &EmailService{db: db}
}

func (s *EmailService) SendEmail(to, subject, body string) error {
	var config struct {
		Host      string
		Port      int
		Username  string
		Password  string
		FromEmail string
	}

	err := s.db.QueryRow(`
		SELECT host, port, username, password, from_email 
		FROM smtp_config 
		WHERE is_active = TRUE LIMIT 1`).Scan(
		&config.Host,
		&config.Port,
		&config.Username,
		&config.Password,
		&config.FromEmail,
	)

	if err != nil {
		return fmt.Errorf("no se pudo obtener la configuración SMTP: %v", err)
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		config.FromEmail, to, subject, body)

	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)

	if config.Port == 587 {
		return s.sendWithStartTLS(config, auth, msg, to)
	} else if config.Port == 465 {
		return s.sendWithTLS(config, auth, msg, to)
	}

	return fmt.Errorf("puerto SMTP no soportado: %d", config.Port)
}

func (s *EmailService) sendWithStartTLS(config struct {
	Host      string
	Port      int
	Username  string
	Password  string
	FromEmail string
}, auth smtp.Auth, msg, to string) error {
	client, err := smtp.Dial(fmt.Sprintf("%s:%d", config.Host, config.Port))
	if err != nil {
		return fmt.Errorf("error al conectar al servidor SMTP: %v", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{
			ServerName:         config.Host,
			InsecureSkipVerify: true,
		}
		if err = client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("error al iniciar TLS: %v", err)
		}
	}

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("error de autenticación: %v", err)
	}

	if err = client.Mail(config.FromEmail); err != nil {
		return fmt.Errorf("error al establecer remitente: %v", err)
	}

	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("error al establecer destinatario: %v", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("error al preparar cuerpo del mensaje: %v", err)
	}
	defer w.Close()

	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("error al escribir mensaje: %v", err)
	}

	return nil
}

func (s *EmailService) sendWithTLS(config struct {
	Host      string
	Port      int
	Username  string
	Password  string
	FromEmail string
}, auth smtp.Auth, msg, to string) error {
	tlsConfig := &tls.Config{
		ServerName:         config.Host,
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port), tlsConfig)
	if err != nil {
		return fmt.Errorf("error al conectar al servidor SMTP: %v", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, config.Host)
	if err != nil {
		return fmt.Errorf("error al crear cliente SMTP: %v", err)
	}
	defer client.Close()

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("error de autenticación: %v", err)
	}

	if err = client.Mail(config.FromEmail); err != nil {
		return fmt.Errorf("error al establecer remitente: %v", err)
	}

	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("error al establecer destinatario: %v", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("error al preparar cuerpo del mensaje: %v", err)
	}
	defer w.Close()

	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("error al escribir mensaje: %v", err)
	}

	return nil
}