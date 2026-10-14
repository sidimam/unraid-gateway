// Package health computes the one-glance state of the gateway and of the Unraid server behind
// it: green, yellow for warnings, red for problems — the same dot the Unraid Drive apps show.
package health

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Level is the traffic light.
type Level string

const (
	OK      Level = "ok"
	Warning Level = "warning"
	Error   Level = "error"
)

// Check is one thing that was looked at.
type Check struct {
	Name   string `json:"name"`
	Level  Level  `json:"level"`
	Detail string `json:"detail,omitempty"`
}

// Summary is what GET /api/v1/health returns.
type Summary struct {
	Status  string   `json:"status"`
	Version string   `json:"version"`
	Level   Level    `json:"level"`
	Reasons []string `json:"reasons"`
	Checks  []Check  `json:"checks"`
}

// Querier runs a GraphQL query against Unraid (the proxy) and returns its "data" object.
type Querier interface {
	Query(ctx context.Context, apiKey, query string, variables map[string]any) (json.RawMessage, error)
}

// Gateway describes the gateway-side facts the server already knows.
type Gateway struct {
	Version              string
	MountedShares        []string
	UnwritableShares     []string
	ConfigWritable       bool
	UserAuth             string
	SharesConfigReadable bool
	NotifyChannels       []string
	IndexEnabled         bool
	Devices              int
}

// unraidQuery mirrors the apps' dashboard query (validated on Unraid 7.3 / unraid-api 4.37).
const unraidQuery = `{ docker { containers { names image state } } }`

type unraidData struct {
	Docker *struct {
		Containers []struct {
			Names []string `json:"names"`
			Image string   `json:"image"`
			State string   `json:"state"`
		} `json:"containers"`
	} `json:"docker"`
	Notifications *struct {
		Overview *struct {
			Unread *struct {
				Total   int `json:"total"`
				Warning int `json:"warning"`
				Alert   int `json:"alert"`
			} `json:"unread"`
		} `json:"overview"`
	} `json:"notifications"`
}

// Assess combines the gateway facts with a live look at Unraid (skipped when apiKey is empty).
func Assess(ctx context.Context, gw Gateway, q Querier, apiKey string) Summary {
	var checks []Check
	var errs, warns []string
	add := func(name string, lvl Level, detail string) {
		checks = append(checks, Check{Name: name, Level: lvl, Detail: detail})
		switch lvl {
		case Error:
			errs = append(errs, detail)
		case Warning:
			warns = append(warns, detail)
		}
	}
	// Gateway side.
	if len(gw.MountedShares) == 0 {
		add("shares", Error, "no share mounted under /data")
	} else if len(gw.UnwritableShares) > 0 {
		add("shares", Warning, fmt.Sprintf("not writable: %s (mounted read/write but the container user cannot write)", strings.Join(gw.UnwritableShares, ", ")))
	} else {
		add("shares", OK, fmt.Sprintf("%d mounted", len(gw.MountedShares)))
	}
	if gw.ConfigWritable {
		add("config", OK, "/config writable")
	} else {
		add("config", Warning, "/config is not writable: devices, remembered key and index are lost at restart")
	}
	if gw.UserAuth != "off" && !gw.SharesConfigReadable {
		add("users", Warning, "USER_AUTH is on but the Unraid share configuration is not readable (mount /etc/samba/smb-shares.conf)")
	} else {
		add("users", OK, "user auth "+gw.UserAuth)
	}
	if len(gw.NotifyChannels) == 0 {
		add("notifications", OK, "no channel configured")
	} else {
		add("notifications", OK, strings.Join(gw.NotifyChannels, ", "))
	}
	add("devices", OK, fmt.Sprintf("%d registered", gw.Devices))
	// Unraid side.
	if q != nil && apiKey != "" {
		raw, err := q.Query(ctx, apiKey, unraidQuery, nil)
		if err != nil {
			add("unraid", Error, "Unraid API: "+err.Error())
		} else {
			var d unraidData
			if jerr := json.Unmarshal(raw, &d); jerr != nil {
				add("unraid", Warning, "Unraid API answered something unexpected")
			} else {
				assessUnraid(d, add)
			}
		}
	}
	level := OK
	reasons := []string{}
	if len(errs) > 0 {
		level = Error
		reasons = append(errs, warns...)
	} else if len(warns) > 0 {
		level = Warning
		reasons = warns
	}
	return Summary{Status: "ok", Version: gw.Version, Level: level, Reasons: reasons, Checks: checks}
}

func assessUnraid(d unraidData, add func(string, Level, string)) {
	// Only what concerns the gateway (rule of 13/10/2026): Unraid's own state — array, disks,
	// parity, load, notifications — stays in the dashboard as information and never colours the dot.
	if dk := d.Docker; dk != nil {
		for _, c := range dk.Containers {
			hay := strings.ToLower(c.Image + " " + strings.Join(c.Names, " "))
			if strings.Contains(hay, "unraid-gateway") && c.State != "RUNNING" {
				add("gateway-container", Warning, "unraid-gateway container "+strings.ToLower(c.State))
			}
		}
	}
}
