package servers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestShutdownDesktop(t *testing.T) {
	t.Setenv("BILILIVE_DESKTOP_SHUTDOWN_TOKEN", "test-token")
	previous := shutdownFunc
	defer func() { shutdownFunc = previous }()
	called := make(chan struct{}, 1)
	shutdownFunc = func() { called <- struct{}{} }

	for _, tc := range []struct {
		name       string
		remoteAddr string
		token      string
		want       int
	}{
		{"missing token", "127.0.0.1:1234", "", http.StatusForbidden},
		{"wrong token", "127.0.0.1:1234", "wrong", http.StatusForbidden},
		{"remote host", "192.0.2.1:1234", "test-token", http.StatusForbidden},
		{"desktop", "127.0.0.1:1234", "test-token", http.StatusAccepted},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/desktop/shutdown", nil)
			req.RemoteAddr = tc.remoteAddr
			req.Header.Set("X-Desktop-Shutdown-Token", tc.token)
			writer := httptest.NewRecorder()
			shutdownDesktop(writer, req)
			if writer.Code != tc.want {
				t.Fatalf("status = %d, want %d", writer.Code, tc.want)
			}
		})
	}
	select {
	case <-called:
	case <-time.After(time.Second):
		t.Fatal("shutdown callback was not called")
	}
	t.Setenv("BILILIVE_DESKTOP_SHUTDOWN_TOKEN", "")
	req := httptest.NewRequest(http.MethodPost, "/api/desktop/shutdown", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("X-Desktop-Shutdown-Token", "test-token")
	writer := httptest.NewRecorder()
	shutdownDesktop(writer, req)
	if writer.Code != http.StatusForbidden {
		t.Fatalf("status without desktop token = %d, want %d", writer.Code, http.StatusForbidden)
	}
}
