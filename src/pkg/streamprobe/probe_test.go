package streamprobe

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestUpstreamReadIdleTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("A"))
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	defer server.Close()

	upstreamURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	p := New(Config{UpstreamURL: upstreamURL, ReadIdleTimeout: 100 * time.Millisecond})
	p.ctx, p.cancel = context.WithCancel(context.Background())
	defer p.Stop()
	if err := p.connectUpstream(); err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 1)
	if _, err := p.upstreamBody.Read(buf); err != nil || buf[0] != 'A' {
		t.Fatalf("首次读取失败: data=%q, err=%v", buf, err)
	}
	_, err = p.upstreamBody.Read(buf)
	var netErr net.Error
	if !errors.As(err, &netErr) || !netErr.Timeout() {
		t.Fatalf("上游停止发送数据后应超时，实际错误: %v", err)
	}
}
