package cli

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"os"
	"os/exec"
	"runtime"
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

	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if page, err := fetchHyattBrowserHTML(ctx, session, profile, targetURL, "window.STORE"); err == nil {
			store, ok := extractHyattStoreFromHTML(page)
			if ok {
				data, marshalErr := json.Marshal(store)
				if marshalErr == nil {
					return data, nil
				}
			}
			lastErr = fmt.Errorf("browser page did not expose parseable window.STORE")
		} else {
			lastErr = err
		}
		if err := navigateHyattBrowser(ctx, session, profile, targetURL); err != nil {
			return nil, err
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
	if page, err := fetchHyattBrowserHTML(ctx, session, profile, targetURL, "spiritCode"); err == nil {
		if data, ok := extractJSONFromBrowserHTML(page); ok {
			return data, nil
		}
	}
	if err := navigateHyattBrowser(ctx, session, profile, targetURL); err != nil {
		return nil, err
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

func fetchHyattBrowserHTML(ctx context.Context, session, profile, targetURL, waitFor string) ([]byte, error) {
	quotedURL, err := json.Marshal(targetURL)
	if err != nil {
		return nil, err
	}
	quotedWait, err := json.Marshal(waitFor)
	if err != nil {
		return nil, err
	}
	args := []string{"--session", session}
	if !hyattBrowserHeadless() {
		args = append(args, "--headed")
	}
	if profile != "" {
		args = append(args, "--profile", profile)
	}
	code := fmt.Sprintf(`import base64, time
browser.goto(%s)
deadline = time.time() + 15
needle = %s
page = browser.html
while needle not in page and time.time() < deadline:
    time.sleep(0.1)
    page = browser.html
print(base64.b64encode(page.encode("utf-8")).decode("ascii"))`, string(quotedURL), string(quotedWait))
	args = append(args, "python", code)
	out, err := exec.CommandContext(ctx, "browser-use", args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("browser python extraction failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	data, err := parseBrowserUseBase64Result(out)
	if err != nil {
		return nil, err
	}
	backgroundHyattBrowser(ctx)
	return data, nil
}

func navigateHyattBrowser(ctx context.Context, session, profile, targetURL string) error {
	if isBrowserUseSessionRunning(ctx, session) {
		if err := navigateExistingBrowserUseSession(ctx, session, targetURL); err == nil {
			backgroundHyattBrowser(ctx)
			return nil
		}
	}
	openArgs := browserUseOpenArgs(session, profile, targetURL)
	out, err := exec.CommandContext(ctx, "browser-use", openArgs...).CombinedOutput()
	if err == nil || isIgnorableBrowserNavigationAbort(out) {
		backgroundHyattBrowser(ctx)
		return nil
	}
	if bytes.Contains(out, []byte("already running with different config")) {
		if navErr := navigateExistingBrowserUseSession(ctx, session, targetURL); navErr == nil {
			backgroundHyattBrowser(ctx)
			return nil
		}
	}
	return fmt.Errorf("browser open failed: %w: %s", err, strings.TrimSpace(string(out)))
}

func browserUseOpenArgs(session, profile, targetURL string) []string {
	args := []string{"--session", session}
	if !hyattBrowserHeadless() {
		args = append(args, "--headed")
	}
	if profile != "" {
		args = append(args, "--profile", profile)
	}
	return append(args, "open", targetURL)
}

func hyattBrowserHeadless() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("HYATT_BROWSER_HEADLESS"))) {
	case "1", "true", "yes":
		return true
	default:
		return false
	}
}

func shouldBackgroundHyattBrowser() bool {
	if hyattBrowserHeadless() {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(os.Getenv("HYATT_BROWSER_BACKGROUND"))) {
	case "0", "false", "no":
		return false
	default:
		return true
	}
}

func backgroundHyattBrowser(ctx context.Context) {
	if runtime.GOOS != "darwin" || !shouldBackgroundHyattBrowser() {
		return
	}
	script := `tell application "Google Chrome"
  repeat with w in windows
    try
      if (URL of active tab of w) contains "hyatt.com/explore-hotels" then
        set minimized of w to true
      end if
    end try
  end repeat
end tell`
	bgCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
	cmd := exec.CommandContext(bgCtx, "osascript")
	cmd.Stdin = strings.NewReader(script)
	if err := cmd.Start(); err != nil {
		cancel()
		return
	}
	go func() {
		defer cancel()
		_ = cmd.Wait()
	}()
}

func isBrowserUseSessionRunning(ctx context.Context, session string) bool {
	out, err := exec.CommandContext(ctx, "browser-use", "sessions").CombinedOutput()
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == session && fields[1] == "running" {
			return true
		}
	}
	return false
}

func navigateExistingBrowserUseSession(ctx context.Context, session, targetURL string) error {
	quoted, err := json.Marshal(targetURL)
	if err != nil {
		return err
	}
	expr := "window.location.href = " + string(quoted)
	out, err := exec.CommandContext(ctx, "browser-use", "--session", session, "eval", expr).CombinedOutput()
	if err != nil {
		return fmt.Errorf("browser navigation failed: %w: %s", err, strings.TrimSpace(string(out)))
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

func parseBrowserUseBase64Result(out []byte) ([]byte, error) {
	text := strings.TrimSpace(string(out))
	lines := strings.Split(text, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(strings.TrimPrefix(lines[i], "result:"))
		if line == "" {
			continue
		}
		data, err := base64.StdEncoding.DecodeString(line)
		if err == nil {
			return data, nil
		}
	}
	return nil, fmt.Errorf("browser python extraction returned no base64 HTML")
}

func extractJSONFromBrowserHTML(raw []byte) (json.RawMessage, bool) {
	text := strings.TrimSpace(string(raw))
	if json.Valid([]byte(text)) {
		return json.RawMessage(text), true
	}
	lower := strings.ToLower(text)
	start := strings.Index(lower, "<pre>")
	end := strings.LastIndex(lower, "</pre>")
	if start < 0 || end <= start {
		return nil, false
	}
	body := html.UnescapeString(text[start+len("<pre>") : end])
	body = strings.TrimSpace(body)
	if !json.Valid([]byte(body)) {
		return nil, false
	}
	return json.RawMessage(body), true
}

func truncateForError(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	return s[:limit] + "..."
}
