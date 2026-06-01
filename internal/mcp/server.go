package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

const serviceInfoURI = "svarog://service/info"

// Build constructs an MCP server with tools, one static resource and one prompt.
func Build(deps *Deps) *mcpserver.MCPServer {
	s := mcpserver.NewMCPServer(
		deps.Config.AppName+" MCP",
		deps.Config.AppVersion,
		mcpserver.WithToolCapabilities(true),
		mcpserver.WithResourceCapabilities(true, true),
		mcpserver.WithPromptCapabilities(true),
	)

	registerTools(s, deps)
	registerResources(s, deps)
	registerPrompts(s, deps)

	return s
}

func registerTools(s *mcpserver.MCPServer, deps *Deps) {
	weatherTool := mcp.NewTool("weather_forecast",
		mcp.WithDescription(`Fetches a short weather forecast from the public Open-Meteo API (no API key).

Use this tool when the user asks about current weather at a geographic location. Provide WGS-84 coordinates (latitude/longitude). Example: Moscow ≈ lat 55.75, lon 37.62.

Returns current air temperature (°C) and wind speed (km/h). Does not write to svarog-core or Postgres.`),
		mcp.WithNumber("latitude",
			mcp.Required(),
			mcp.Description("Latitude in decimal degrees, range -90..90 (e.g. 55.75 for Moscow)."),
		),
		mcp.WithNumber("longitude",
			mcp.Required(),
			mcp.Description("Longitude in decimal degrees, range -180..180 (e.g. 37.62 for Moscow)."),
		),
	)
	s.AddTool(weatherTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		lat, err := req.RequireFloat("latitude")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		lon, err := req.RequireFloat("longitude")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		text, err := fetchWeather(ctx, deps.HTTPClient, lat, lon)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(text), nil
	})

	listUsersTool := mcp.NewTool("list_users",
		mcp.WithDescription(`Lists registered users from the svarog-core Postgres database (read-only).

Use when investigating accounts, auditing sign-ups, or correlating sessions with emails. Never returns password hashes — only id, email, created_at.

Optional limit caps rows (default 20, max 100). Requires MCP server to reach the same Postgres as svarog-core.`),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of rows to return (1–100). Defaults to 20 when omitted or zero."),
		),
	)
	s.AddTool(listUsersTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		limit := int(req.GetFloat("limit", 20))
		if limit <= 0 {
			limit = 20
		}
		if limit > 100 {
			limit = 100
		}
		text, err := listUsers(ctx, deps, limit)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(text), nil
	})

	listSessionsTool := mcp.NewTool("list_active_sessions",
		mcp.WithDescription(`Lists non-revoked, unexpired sessions from Postgres together with the user email (read-only).

Use to debug “who is logged in”, session sprawl, or stale devices. Fields: session id, user id, email, expires_at, last_seen_at, user_agent, ip.

Optional limit (default 20, max 100). Active means revoked_at IS NULL AND expires_at > now().`),
		mcp.WithNumber("limit",
			mcp.Description("Maximum rows (1–100). Default 20."),
		),
	)
	s.AddTool(listSessionsTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		limit := int(req.GetFloat("limit", 20))
		if limit <= 0 {
			limit = 20
		}
		if limit > 100 {
			limit = 100
		}
		text, err := listActiveSessions(ctx, deps, limit)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(text), nil
	})

	registerTool := mcp.NewTool("register_user",
		mcp.WithDescription(`Creates a new user via the svarog-core public HTTP API (POST /v1/auth/register).

Use to provision demo accounts or test auth flows without raw SQL. Requires SVAROG_HTTP_BASE (e.g. http://app:8080 in Docker). On success returns user_id JSON from the API. Duplicate email surfaces as an error.`),
		mcp.WithString("email",
			mcp.Required(),
			mcp.Description("Unique email address (case-insensitive in DB)."),
		),
		mcp.WithString("password",
			mcp.Required(),
			mcp.Description("Plain-text password; sent only over the register HTTP call and stored hashed by svarog-core."),
		),
	)
	s.AddTool(registerTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		email, err := req.RequireString("email")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		password, err := req.RequireString("password")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		text, err := registerViaHTTP(ctx, deps, email, password)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(text), nil
	})
}

func registerResources(s *mcpserver.MCPServer, deps *Deps) {
	resource := mcp.NewResource(
		serviceInfoURI,
		"Service info",
		mcp.WithResourceDescription("Static JSON document describing svarog-core endpoints, MCP tools, and environment hints for agents."),
		mcp.WithMIMEType("application/json"),
	)
	s.AddResource(resource, func(ctx context.Context, _ mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		payload := map[string]any{
			"service":  deps.Config.AppName,
			"version":  deps.Config.AppVersion,
			"http_api": strings.TrimRight(deps.Config.SvarogHTTPBase, "/"),
			"endpoints": map[string]string{
				"register": "POST /v1/auth/register",
				"login":    "POST /v1/auth/login",
				"logout":   "POST /v1/auth/logout",
				"me":       "GET /v1/auth/me",
			},
			"mcp_tools": []string{
				"weather_forecast",
				"list_users",
				"list_active_sessions",
				"register_user",
			},
			"resource_uri": serviceInfoURI,
			"prompts":      []string{"investigate_user"},
		}
		b, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return nil, err
		}
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      serviceInfoURI,
				MIMEType: "application/json",
				Text:     string(b),
			},
		}, nil
	})
}

func registerPrompts(s *mcpserver.MCPServer, deps *Deps) {
	s.AddPrompt(mcp.NewPrompt("investigate_user",
		mcp.WithPromptDescription(`Structured checklist for investigating a svarog-core account by email.

The agent should: (1) read resource svarog://service/info, (2) list_users filtered mentally by email, (3) list_active_sessions for the same email, (4) optionally register_user only if the account is missing and the user asked to create one.`),
		mcp.WithArgument("email",
			mcp.ArgumentDescription("Email address to investigate (required)."),
			mcp.RequiredArgument(),
		),
	), func(_ context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		email := strings.TrimSpace(request.Params.Arguments["email"])
		if email == "" {
			return nil, fmt.Errorf("email is required")
		}
		body := fmt.Sprintf(`Investigate the svarog-core user with email %q.

Steps:
1. Read resource %s for API and tool names.
2. Call list_users (limit 100) and locate this email.
3. Call list_active_sessions (limit 100) and note sessions for that email.
4. Summarize: user id, created_at, active session count, nearest expiry, suspicious IPs/user-agents.
5. Do NOT expose password hashes or session tokens. Do not call register_user unless explicitly asked to create the account.`,
			email, serviceInfoURI)
		return mcp.NewGetPromptResult(
			"Investigate svarog user",
			[]mcp.PromptMessage{
				mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(body)),
			},
		), nil
	})
}

// --- tool implementations ----------------------------------------------------

type openMeteoResponse struct {
	Current struct {
		Temperature2m float64 `json:"temperature_2m"`
		WindSpeed10m  float64 `json:"wind_speed_10m"`
	} `json:"current"`
}

func fetchWeather(ctx context.Context, client *http.Client, lat, lon float64) (string, error) {
	u, _ := url.Parse("https://api.open-meteo.com/v1/forecast")
	q := u.Query()
	q.Set("latitude", fmt.Sprintf("%f", lat))
	q.Set("longitude", fmt.Sprintf("%f", lon))
	q.Set("current", "temperature_2m,wind_speed_10m")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("open-meteo request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("open-meteo status %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}
	var payload openMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode open-meteo: %w", err)
	}
	return fmt.Sprintf(
		"Location (%.4f, %.4f): temperature %.1f °C, wind %.1f km/h (Open-Meteo current).",
		lat, lon, payload.Current.Temperature2m, payload.Current.WindSpeed10m,
	), nil
}

func listUsers(ctx context.Context, deps *Deps, limit int) (string, error) {
	const q = `
		SELECT id::text, email::text, created_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1`
	rows, err := deps.Pool.Query(ctx, q, limit)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type row struct {
		ID        string    `json:"id"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"created_at"`
	}
	var out []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.ID, &r.Email, &r.CreatedAt); err != nil {
			return "", err
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func listActiveSessions(ctx context.Context, deps *Deps, limit int) (string, error) {
	const q = `
		SELECT s.id::text, s.user_id::text, u.email::text, s.expires_at, s.last_seen_at,
		       s.user_agent, COALESCE(host(s.ip), '')
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.revoked_at IS NULL AND s.expires_at > now()
		ORDER BY s.last_seen_at DESC
		LIMIT $1`
	rows, err := deps.Pool.Query(ctx, q, limit)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type row struct {
		SessionID  string    `json:"session_id"`
		UserID     string    `json:"user_id"`
		Email      string    `json:"email"`
		ExpiresAt  time.Time `json:"expires_at"`
		LastSeenAt time.Time `json:"last_seen_at"`
		UserAgent  string    `json:"user_agent"`
		IP         string    `json:"ip"`
	}
	var out []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.SessionID, &r.UserID, &r.Email, &r.ExpiresAt, &r.LastSeenAt, &r.UserAgent, &r.IP); err != nil {
			return "", err
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func registerViaHTTP(ctx context.Context, deps *Deps, email, password string) (string, error) {
	base := strings.TrimRight(deps.Config.SvarogHTTPBase, "/")
	body := fmt.Sprintf(`{"email":%q,"password":%q}`, email, password)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/v1/auth/register", strings.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := deps.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("register http: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("register http status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	return string(raw), nil
}
