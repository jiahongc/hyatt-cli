package cli

import (
	"encoding/base64"
	"testing"
)

func TestShouldUseHyattBrowserFirstDefault(t *testing.T) {
	t.Setenv("HYATT_BROWSER_FALLBACK", "")
	t.Setenv("HYATT_TRANSPORT", "")

	if !shouldUseHyattBrowserFirst() {
		t.Fatal("expected browser-first transport by default")
	}
}

func TestShouldUseHyattBrowserFirstDisabledByFallbackEnv(t *testing.T) {
	t.Setenv("HYATT_BROWSER_FALLBACK", "0")
	t.Setenv("HYATT_TRANSPORT", "")

	if shouldUseHyattBrowserFirst() {
		t.Fatal("expected HYATT_BROWSER_FALLBACK=0 to disable browser-first transport")
	}
}

func TestShouldUseHyattBrowserFirstDisabledByHTTPTransport(t *testing.T) {
	t.Setenv("HYATT_BROWSER_FALLBACK", "")
	t.Setenv("HYATT_TRANSPORT", "http")

	if shouldUseHyattBrowserFirst() {
		t.Fatal("expected HYATT_TRANSPORT=http to disable browser-first transport")
	}
}

func TestHyattBrowserHeadlessDefault(t *testing.T) {
	t.Setenv("HYATT_BROWSER_HEADLESS", "")

	if hyattBrowserHeadless() {
		t.Fatal("expected headed browser by default because Hyatt blocks basic headless pages")
	}
}

func TestHyattBrowserHeadlessOptIn(t *testing.T) {
	t.Setenv("HYATT_BROWSER_HEADLESS", "true")

	if !hyattBrowserHeadless() {
		t.Fatal("expected HYATT_BROWSER_HEADLESS=true to opt into headless mode")
	}
}

func TestShouldBackgroundHyattBrowserDefault(t *testing.T) {
	t.Setenv("HYATT_BROWSER_HEADLESS", "")
	t.Setenv("HYATT_BROWSER_BACKGROUND", "")

	if !shouldBackgroundHyattBrowser() {
		t.Fatal("expected browser backgrounding by default")
	}
}

func TestShouldBackgroundHyattBrowserDisabled(t *testing.T) {
	t.Setenv("HYATT_BROWSER_HEADLESS", "")
	t.Setenv("HYATT_BROWSER_BACKGROUND", "0")

	if shouldBackgroundHyattBrowser() {
		t.Fatal("expected HYATT_BROWSER_BACKGROUND=0 to disable browser backgrounding")
	}
}

func TestShouldBackgroundHyattBrowserDisabledWhenHeadless(t *testing.T) {
	t.Setenv("HYATT_BROWSER_HEADLESS", "true")
	t.Setenv("HYATT_BROWSER_BACKGROUND", "")

	if shouldBackgroundHyattBrowser() {
		t.Fatal("expected headless mode to skip browser backgrounding")
	}
}

func TestParseBrowserUseBase64Result(t *testing.T) {
	want := []byte("<html>ok</html>")
	out := []byte("noise\n" + base64.StdEncoding.EncodeToString(want) + "\n")

	got, err := parseBrowserUseBase64Result(out)
	if err != nil {
		t.Fatalf("parseBrowserUseBase64Result returned error: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestExtractJSONFromBrowserHTMLPre(t *testing.T) {
	got, ok := extractJSONFromBrowserHTML([]byte(`<html><body><pre>{&quot;ok&quot;:true}</pre></body></html>`))
	if !ok {
		t.Fatal("expected JSON extraction to succeed")
	}
	if string(got) != `{"ok":true}` {
		t.Fatalf("got %s", got)
	}
}
