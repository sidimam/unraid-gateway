// Package config loads the gateway configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds every tunable of the gateway.
type Config struct {
	// ListenAddr is the address the HTTP server binds to (default ":8484").
	ListenAddr string
	// UnraidURL is the base URL of the Unraid WebGUI, e.g. "http://192.168.1.10".
	UnraidURL string
	// UnraidInsecureTLS skips certificate verification when UnraidURL is https
	// with a self-signed certificate.
	UnraidInsecureTLS bool
	// ValidateQuery is the GraphQL query used to check an API key.
	ValidateQuery string
	// DataRoot is the directory exposed by the file API (default "/data").
	DataRoot string
	// ReadOnly disables every mutating file operation.
	ReadOnly bool
	// SessionTTL is the lifetime of a session token.
	SessionTTL time.Duration
	// MaxLoginAttempts is the number of failed logins per IP before lockout.
	MaxLoginAttempts int
	// LoginLockout is how long an IP stays locked after too many failures.
	LoginLockout time.Duration
	// TLSCert and TLSKey enable HTTPS directly on the gateway when both are set.
	TLSCert string
	TLSKey  string
	// TrustProxy makes the gateway honour X-Forwarded-For for client IPs.
	TrustProxy bool
	// UploadTTL is how long an unfinished resumable upload is kept.
	UploadTTL time.Duration
	// MaxJSONBody caps the size of JSON request bodies.
	MaxJSONBody int64
	// ChangesWalkLimit bounds the number of entries scanned by /fs/changes.
	ChangesWalkLimit int
	// ChangesDeadline bounds the wall time of one /fs/changes scan.
	ChangesDeadline time.Duration
	// LogRequests enables per-request access logging.
	LogRequests bool
	// LogFormat is "text" (readable, default) or "json".
	LogFormat string
	// LogHealthchecks also logs /healthz requests (noisy: Docker probes every 30s).
	LogHealthchecks bool
	// LogLevel is debug, info, warn or error.
	LogLevel string
	// UserAuth controls Unraid user authentication: "off", "optional" (default) or "required".
	UserAuth string
	// SMBAddr is the Unraid SMB endpoint used to validate user passwords (host:445).
	SMBAddr string
	// DeviceRegistration "on" requires apps to register their installation (DEVICE_REGISTRATION); "off" disables the registry.
	DeviceRegistration string
	// DevicesFile persists the registry (DEVICES_FILE).
	DevicesFile string
	// Notifications.
	NotifyUnraid                                                          bool
	NotifyUnraidAPIKey                                                    string
	SMTPHost, SMTPPort, SMTPUser, SMTPPassword, SMTPFrom, SMTPTo, SMTPTLS string
	TelegramToken, TelegramChatID                                         string
	// WebUIKey is an Unraid API key the web UI may use with one click (WEBUI_API_KEY); empty = none.
	WebUIKey string
	// WebUIKeyFile is where "Remember this key on the gateway" stores the key (WEBUI_API_KEY_FILE).
	WebUIKeyFile string
	// IndexDB is the SQLite file of the item index ("" disables the index, ":memory:" keeps it in RAM).
	IndexDB string
	// IndexDirScan is how often directories are checked for changes made behind the gateway's back.
	IndexDirScan time.Duration
	// IndexFullScan is how often every entry is re-stat'ed.
	IndexFullScan time.Duration
	// SharesConfig is Unraid's share configuration: the generated
	// /etc/samba/smb-shares.conf (file, recommended) or /boot/config/shares (directory).
	SharesConfig string
}

// Load reads the configuration from the environment.
func Load() (Config, error) {
	c := Config{
		ListenAddr:         env("LISTEN_ADDR", ":8484"),
		UnraidURL:          strings.TrimRight(env("UNRAID_URL", ""), "/"),
		ValidateQuery:      env("UNRAID_VALIDATE_QUERY", "query { me { id name roles } }"),
		DataRoot:           env("DATA_ROOT", "/data"),
		TLSCert:            env("TLS_CERT", ""),
		TLSKey:             env("TLS_KEY", ""),
		SessionTTL:         dur("SESSION_TTL", 12*time.Hour),
		LoginLockout:       dur("LOGIN_LOCKOUT", 15*time.Minute),
		UploadTTL:          dur("UPLOAD_TTL", 24*time.Hour),
		ChangesDeadline:    dur("CHANGES_DEADLINE", 20*time.Second),
		MaxLoginAttempts:   num("MAX_LOGIN_ATTEMPTS", 5),
		ChangesWalkLimit:   num("CHANGES_WALK_LIMIT", 250000),
		MaxJSONBody:        int64(num("MAX_JSON_BODY", 1<<20)),
		UserAuth:           strings.ToLower(env("USER_AUTH", "optional")),
		LogFormat:          strings.ToLower(env("LOG_FORMAT", "text")),
		LogLevel:           strings.ToLower(env("LOG_LEVEL", "info")),
		SMBAddr:            env("UNRAID_SMB_ADDR", ""),
		SharesConfig:       env("SHARES_CONFIG", env("SHARES_CONFIG_DIR", "/unraid-shares/smb-shares.conf")),
		IndexDB:            env("INDEX_DB", "/config/index.db"),
		WebUIKey:           env("WEBUI_API_KEY", ""),
		DeviceRegistration: env("DEVICE_REGISTRATION", "on"),
		DevicesFile:        env("DEVICES_FILE", "/config/devices.json"),
		NotifyUnraidAPIKey: env("NOTIFY_UNRAID_API_KEY", ""),
		SMTPHost:           env("SMTP_HOST", ""),
		SMTPPort:           env("SMTP_PORT", "587"),
		SMTPUser:           env("SMTP_USER", ""),
		SMTPPassword:       env("SMTP_PASSWORD", ""),
		SMTPFrom:           env("SMTP_FROM", ""),
		SMTPTo:             env("SMTP_TO", ""),
		SMTPTLS:            env("SMTP_TLS", "starttls"),
		TelegramToken:      env("TELEGRAM_BOT_TOKEN", ""),
		TelegramChatID:     env("TELEGRAM_CHAT_ID", ""),
		WebUIKeyFile:       env("WEBUI_API_KEY_FILE", "/config/webui.key"),
		IndexDirScan:       dur("INDEX_DIR_SCAN", 5*time.Minute),
		IndexFullScan:      dur("INDEX_FULL_SCAN", 6*time.Hour),
	}
	var err error
	if c.UnraidInsecureTLS, err = boolean("UNRAID_INSECURE_TLS", false); err != nil {
		return c, err
	}
	if c.ReadOnly, err = boolean("READ_ONLY", false); err != nil {
		return c, err
	}
	if c.TrustProxy, err = boolean("TRUST_PROXY", false); err != nil {
		return c, err
	}
	if c.NotifyUnraid, err = boolean("NOTIFY_UNRAID", true); err != nil {
		return c, err
	}
	if c.LogRequests, err = boolean("LOG_REQUESTS", true); err != nil {
		return c, err
	}
	if c.LogHealthchecks, err = boolean("LOG_HEALTHCHECKS", false); err != nil {
		return c, err
	}
	if c.UnraidURL == "" {
		return c, fmt.Errorf("UNRAID_URL is required (e.g. http://192.168.1.10)")
	}
	if !strings.HasPrefix(c.UnraidURL, "http://") && !strings.HasPrefix(c.UnraidURL, "https://") {
		return c, fmt.Errorf("UNRAID_URL must start with http:// or https://")
	}
	if (c.TLSCert == "") != (c.TLSKey == "") {
		return c, fmt.Errorf("TLS_CERT and TLS_KEY must be set together")
	}
	switch c.UserAuth {
	case "off", "optional", "required":
	default:
		return c, fmt.Errorf("USER_AUTH must be off, optional or required")
	}
	if c.SMBAddr == "" {
		host := strings.TrimPrefix(strings.TrimPrefix(c.UnraidURL, "https://"), "http://")
		if i := strings.IndexAny(host, ":/"); i >= 0 {
			host = host[:i]
		}
		c.SMBAddr = host + ":445"
	}
	return c, nil
}

func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func dur(key string, def time.Duration) time.Duration {
	v := env(key, "")
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

func num(key string, def int) int {
	v := env(key, "")
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func boolean(key string, def bool) (bool, error) {
	v := env(key, "")
	if v == "" {
		return def, nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def, fmt.Errorf("%s: expected true/false, got %q", key, v)
	}
	return b, nil
}
