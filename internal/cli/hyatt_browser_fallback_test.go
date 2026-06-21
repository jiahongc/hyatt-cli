package cli

import "testing"

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
