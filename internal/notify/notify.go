// Package notify fans an event out to the configured channels: Unraid's own
// notification system (through the GraphQL API), an SMTP mailbox and a Telegram
// chat. Every channel is optional; failures are logged, never fatal.
package notify

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"time"
)

// Importance mirrors Unraid's levels.
type Importance string

const (
	Info    Importance = "INFO"
	Warning Importance = "WARNING"
	Alert   Importance = "ALERT"
)

// UnraidSender runs the createNotification mutation with an API key.
type UnraidSender interface {
	Query(ctx context.Context, apiKey, query string, variables map[string]any) (json.RawMessage, error)
}

// Config selects the channels.
type Config struct {
	UnraidEnabled bool
	UnraidAPIKey  string // optional fixed key; otherwise the key of the session that caused the event

	SMTPHost, SMTPPort, SMTPUser, SMTPPassword, SMTPFrom, SMTPTo string
	SMTPTLS                                                      string // starttls (default) | tls | none

	TelegramToken, TelegramChatID string
}

// Notifier sends events.
type Notifier struct {
	cfg    Config
	unraid UnraidSender
	log    *slog.Logger
	http   *http.Client
}

// New creates a notifier; unraid may be nil.
func New(cfg Config, unraid UnraidSender, log *slog.Logger) *Notifier {
	return &Notifier{cfg: cfg, unraid: unraid, log: log, http: &http.Client{Timeout: 15 * time.Second}}
}

// Channels lists what is configured, for the console and the UI.
func (n *Notifier) Channels() []string {
	var out []string
	if n.cfg.UnraidEnabled {
		out = append(out, "unraid")
	}
	if n.cfg.SMTPHost != "" && n.cfg.SMTPTo != "" {
		out = append(out, "smtp")
	}
	if n.cfg.TelegramToken != "" && n.cfg.TelegramChatID != "" {
		out = append(out, "telegram")
	}
	return out
}

// Send delivers to every channel. sessionKey is used for Unraid when no fixed key is configured.
// It returns one error per failed channel (nil when everything went through).
func (n *Notifier) Send(ctx context.Context, sessionKey, title, body string, imp Importance) []error {
	var errs []error
	if n.cfg.UnraidEnabled && n.unraid != nil {
		key := n.cfg.UnraidAPIKey
		if key == "" {
			key = sessionKey
		}
		if key == "" {
			errs = append(errs, errors.New("unraid: no API key available for the notification"))
		} else if err := n.sendUnraid(ctx, key, title, body, imp); err != nil {
			// Unraid answers "Forbidden resource" when the key lacks the notification permission
			// (a VIEWER key): say what to do instead of echoing the GraphQL error.
			if strings.Contains(strings.ToLower(err.Error()), "forbidden") {
				which := "the signing-in key"
				if n.cfg.UnraidAPIKey != "" {
					which = "NOTIFY_UNRAID_API_KEY"
				}
				errs = append(errs, fmt.Errorf("unraid: %s is not allowed to create notifications (VIEWER role). Create an ADMIN key in Unraid (Settings › Management Access › API Keys) and set it as NOTIFY_UNRAID_API_KEY in the container settings (Show more settings…)", which))
			} else {
				errs = append(errs, fmt.Errorf("unraid: %w", err))
			}
		}
	}
	if n.cfg.SMTPHost != "" && n.cfg.SMTPTo != "" {
		if err := n.sendMail(title, body); err != nil {
			errs = append(errs, fmt.Errorf("smtp: %w", err))
		}
	}
	if n.cfg.TelegramToken != "" && n.cfg.TelegramChatID != "" {
		if err := n.sendTelegram(ctx, title, body); err != nil {
			errs = append(errs, fmt.Errorf("telegram: %w", err))
		}
	}
	for _, e := range errs {
		n.log.Warn("notification failed", "err", e, "title", title)
	}
	return errs
}

func (n *Notifier) sendUnraid(ctx context.Context, key, title, body string, imp Importance) error {
	const q = `mutation($input: NotificationData!) { createNotification(input: $input) { id } }`
	_, err := n.unraid.Query(ctx, key, q, map[string]any{"input": map[string]any{
		"title": "unraid-gateway", "subject": title, "description": body, "importance": string(imp),
	}})
	return err
}

func (n *Notifier) sendMail(title, body string) error {
	port := n.cfg.SMTPPort
	if port == "" {
		port = "587"
	}
	addr := net.JoinHostPort(n.cfg.SMTPHost, port)
	from := n.cfg.SMTPFrom
	if from == "" {
		from = n.cfg.SMTPUser
	}
	to := strings.Split(n.cfg.SMTPTo, ",")
	for i := range to {
		to[i] = strings.TrimSpace(to[i])
	}
	msg := fmt.Sprintf("From: unraid-gateway <%s>\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n", from, strings.Join(to, ", "), title, body)
	var auth smtp.Auth
	if n.cfg.SMTPUser != "" {
		auth = smtp.PlainAuth("", n.cfg.SMTPUser, n.cfg.SMTPPassword, n.cfg.SMTPHost)
	}
	switch strings.ToLower(n.cfg.SMTPTLS) {
	case "tls", "ssl", "implicit":
		conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 15 * time.Second}, "tcp", addr, &tls.Config{ServerName: n.cfg.SMTPHost, MinVersion: tls.VersionTLS12})
		if err != nil {
			return err
		}
		defer conn.Close()
		c, err := smtp.NewClient(conn, n.cfg.SMTPHost)
		if err != nil {
			return err
		}
		defer c.Quit()
		return sendWith(c, auth, from, to, msg)
	case "none":
		c, err := smtp.Dial(addr)
		if err != nil {
			return err
		}
		defer c.Quit()
		return sendWith(c, auth, from, to, msg)
	default: // STARTTLS
		c, err := smtp.Dial(addr)
		if err != nil {
			return err
		}
		defer c.Quit()
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err := c.StartTLS(&tls.Config{ServerName: n.cfg.SMTPHost, MinVersion: tls.VersionTLS12}); err != nil {
				return err
			}
		}
		return sendWith(c, auth, from, to, msg)
	}
}

func sendWith(c *smtp.Client, auth smtp.Auth, from string, to []string, msg string) error {
	if auth != nil {
		if err := c.Auth(auth); err != nil {
			return err
		}
	}
	if err := c.Mail(from); err != nil {
		return err
	}
	for _, t := range to {
		if err := c.Rcpt(t); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		return err
	}
	return w.Close()
}

func (n *Notifier) sendTelegram(ctx context.Context, title, body string) error {
	payload, _ := json.Marshal(map[string]any{"chat_id": n.cfg.TelegramChatID, "text": "🟠 " + title + "\n" + body})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.telegram.org/bot"+n.cfg.TelegramToken+"/sendMessage", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := n.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		var b bytes.Buffer
		_, _ = b.ReadFrom(resp.Body)
		return fmt.Errorf("telegram answered HTTP %d: %s", resp.StatusCode, strings.TrimSpace(b.String()))
	}
	return nil
}
