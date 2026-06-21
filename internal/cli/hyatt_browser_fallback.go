package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/jiahongc/hyatt-cli/internal/client"
)

func hyattBrowserCalendarFallback(ctx context.Context, baseURL, path string, params map[string]string, cause error) (json.RawMessage, bool, error) {
	if !isHyattForbidden(cause) {
		return nil, false, nil
	}
	data, err := hyattBrowserCalendar(ctx, baseURL, path, params)
	if err != nil {
		return nil, true, fmt.Errorf("direct Hyatt request was blocked and browser fallback failed: %w", err)
	}
	return data, true, nil
}

func hyattBrowserCalendar(ctx context.Context, baseURL, path string, params map[string]string) (json.RawMessage, error) {
	if err := requireHyattBrowserTransport(); err != nil {
		return nil, err
	}
	targetURL := htmlExtractionRequestURL(baseURL, path, params)
	session := firstNonEmpty(os.Getenv("HYATT_BROWSER_SESSION"), "hyatt-cli")
	profile := strings.TrimSpace(os.Getenv("HYATT_BROWSER_PROFILE"))

	openArgs := []string{"--session", session, "--headed"}
	if profile != "" {
		openArgs = append(openArgs, "--profile", profile)
	}
	openArgs = append(openArgs, "open", targetURL)
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if out, err := runBrowserUseWithSessionRetry(ctx, session, openArgs); err != nil {
			if !isIgnorableBrowserNavigationAbort(out) {
				return nil, fmt.Errorf("browser open failed: %w: %s", err, strings.TrimSpace(string(out)))
			}
		}
		store, err := waitForHyattBrowserStore(ctx, session)
		if err == nil {
			if bytes.Equal(store, []byte("null")) || len(store) == 0 {
				return nil, fmt.Errorf("browser page did not expose window.STORE")
			}
			return store, nil
		}
		lastErr = err
		_ = exec.CommandContext(ctx, "browser-use", "--session", session, "close").Run()
	}
	return nil, lastErr
}

func hyattBrowserJSONFallback(ctx context.Context, baseURL, path string, params map[string]string, cause error) (json.RawMessage, bool, error) {
	if !isHyattForbidden(cause) {
		return nil, false, nil
	}
	data, err := hyattBrowserJSON(ctx, baseURL, path, params)
	if err != nil {
		return nil, true, fmt.Errorf("direct Hyatt request was blocked and browser fallback failed: %w", err)
	}
	return data, true, nil
}

func hyattBrowserJSON(ctx context.Context, baseURL, path string, params map[string]string) (json.RawMessage, error) {
	if err := requireHyattBrowserTransport(); err != nil {
		return nil, err
	}
	targetURL := htmlExtractionRequestURL(baseURL, path, params)
	session := firstNonEmpty(os.Getenv("HYATT_BROWSER_SESSION"), "hyatt-cli")
	profile := strings.TrimSpace(os.Getenv("HYATT_BROWSER_PROFILE"))
	openArgs := []string{"--session", session, "--headed"}
	if profile != "" {
		openArgs = append(openArgs, "--profile", profile)
	}
	openArgs = append(openArgs, "open", targetURL)
	if out, err := runBrowserUseWithSessionRetry(ctx, session, openArgs); err != nil {
		if !isIgnorableBrowserNavigationAbort(out) {
			return nil, fmt.Errorf("browser open failed: %w: %s", err, strings.TrimSpace(string(out)))
		}
	}

	data, err := waitForHyattBrowserJSONBody(ctx, session)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func shouldUseHyattBrowserFirst() bool {
	if os.Getenv("HYATT_BROWSER_FALLBACK") == "0" {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(os.Getenv("HYATT_TRANSPORT"))) {
	case "", "browser", "browser-use":
		return true
	case "http", "direct":
		return false
	default:
		return true
	}
}

func requireHyattBrowserTransport() error {
	if os.Getenv("HYATT_BROWSER_FALLBACK") == "0" {
		return fmt.Errorf("Hyatt browser transport is disabled by HYATT_BROWSER_FALLBACK=0")
	}
	if transport := strings.ToLower(strings.TrimSpace(os.Getenv("HYATT_TRANSPORT"))); transport == "http" || transport == "direct" {
		return fmt.Errorf("Hyatt browser transport is disabled by HYATT_TRANSPORT=%s", transport)
	}
	if _, err := exec.LookPath("browser-use"); err != nil {
		return fmt.Errorf("Hyatt browser transport requires browser-use on PATH")
	}
	return nil
}

func waitForHyattBrowserJSONBody(ctx context.Context, session string) (json.RawMessage, error) {
	deadline := time.Now().Add(30 * time.Second)
	var lastOut []byte
	var lastErr error
	var lastParseErr error
	for {
		evalArgs := []string{"--session", session, "eval", "document.body.innerText"}
		out, err := exec.CommandContext(ctx, "browser-use", evalArgs...).CombinedOutput()
		lastOut, lastErr = out, err
		if err == nil {
			data, parseErr := parseBrowserUseJSONResult(out)
			if parseErr == nil {
				return data, nil
			}
			lastParseErr = parseErr
		}
		if time.Now().After(deadline) {
			break
		}
		timer := time.NewTimer(250 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	if lastErr != nil {
		return nil, fmt.Errorf("direct Hyatt request was blocked and browser JSON fallback failed: %w: %s", lastErr, strings.TrimSpace(string(lastOut)))
	}
	if lastParseErr != nil {
		return nil, lastParseErr
	}
	return nil, fmt.Errorf("direct Hyatt request was blocked and browser JSON fallback returned no data")
}

func waitForHyattBrowserStore(ctx context.Context, session string) (json.RawMessage, error) {
	deadline := time.Now().Add(15 * time.Second)
	var lastOut []byte
	var lastErr error
	for {
		evalArgs := []string{"--session", session, "eval", "window.STORE ? JSON.stringify(window.STORE) : null"}
		out, err := exec.CommandContext(ctx, "browser-use", evalArgs...).CombinedOutput()
		lastOut, lastErr = out, err
		if err == nil {
			store, parseErr := parseBrowserUseJSONResult(out)
			if parseErr == nil && !bytes.Equal(store, []byte("null")) && len(store) > 0 {
				return store, nil
			}
		}
		if time.Now().After(deadline) {
			break
		}
		timer := time.NewTimer(250 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	if lastErr != nil {
		return nil, fmt.Errorf("direct Hyatt request was blocked and browser eval fallback failed: %w: %s", lastErr, strings.TrimSpace(string(lastOut)))
	}
	return nil, fmt.Errorf("direct Hyatt request was blocked and browser fallback did not expose window.STORE")
}

func runBrowserUseWithSessionRetry(ctx context.Context, session string, args []string) ([]byte, error) {
	out, err := exec.CommandContext(ctx, "browser-use", args...).CombinedOutput()
	if err == nil || !bytes.Contains(out, []byte("already running with different config")) {
		return out, err
	}
	_ = exec.CommandContext(ctx, "browser-use", "--session", session, "close").Run()
	return exec.CommandContext(ctx, "browser-use", args...).CombinedOutput()
}

func isIgnorableBrowserNavigationAbort(out []byte) bool {
	return bytes.Contains(out, []byte("Navigation failed: net::ERR_ABORTED"))
}

func isHyattForbidden(err error) bool {
	var apiErr *client.APIError
	return As(err, &apiErr) && apiErr.StatusCode == 403
}

func parseBrowserUseJSONResult(out []byte) (json.RawMessage, error) {
	text := strings.TrimSpace(string(out))
	text = strings.TrimPrefix(text, "result:")
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("browser fallback returned empty output")
	}
	if !json.Valid([]byte(text)) {
		return nil, fmt.Errorf("browser fallback returned invalid JSON: %s", truncateForError(text, 500))
	}
	return json.RawMessage(text), nil
}

func truncateForError(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	return s[:limit] + "..."
}
