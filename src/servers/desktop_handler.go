package servers

import (
	"crypto/subtle"
	"net"
	"net/http"
	"os"
)

// shutdownDesktop 仅允许启动此后端的本机桌面程序触发完整关闭流程。
func shutdownDesktop(writer http.ResponseWriter, r *http.Request) {
	token := os.Getenv("BILILIVE_DESKTOP_SHUTDOWN_TOKEN")
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	ip := net.ParseIP(host)
	if token == "" || err != nil || ip == nil || !ip.IsLoopback() ||
		subtle.ConstantTimeCompare([]byte(r.Header.Get("X-Desktop-Shutdown-Token")), []byte(token)) != 1 {
		http.Error(writer, "forbidden", http.StatusForbidden)
		return
	}
	if shutdownFunc == nil {
		http.Error(writer, "shutdown unavailable", http.StatusServiceUnavailable)
		return
	}

	writer.WriteHeader(http.StatusAccepted)
	go shutdownFunc()
}
