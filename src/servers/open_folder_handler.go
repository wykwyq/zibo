package servers

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bililive-go/bililive-go/src/configs"
)

// openVideoFolder 在本机资源管理器中打开录像所在目录。
func openVideoFolder(writer http.ResponseWriter, r *http.Request) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	ip := net.ParseIP(host)
	if runtime.GOOS != "windows" || err != nil || ip == nil || !ip.IsLoopback() {
		http.Error(writer, "仅支持在本机 Windows 桌面版打开文件夹", http.StatusForbidden)
		return
	}
	var body struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(writer, "无效请求", http.StatusBadRequest)
		return
	}
	cfg := configs.GetCurrentConfig()
	if cfg == nil {
		http.Error(writer, "配置尚未就绪", http.StatusServiceUnavailable)
		return
	}
	folder, err := recordedVideoFolder(cfg.OutPutPath, body.Path)
	if err != nil {
		http.Error(writer, "录像文件不存在或路径无效", http.StatusBadRequest)
		return
	}
	command := exec.Command("explorer.exe", folder)
	if err := command.Start(); err != nil {
		http.Error(writer, "打开文件夹失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_ = command.Process.Release()
	writeJSON(writer, commonResp{Data: "OK"})
}

func recordedVideoFolder(root, relativePath string) (string, error) {
	path, err := getSafePath(root, relativePath)
	if err != nil {
		return "", err
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	realRoot, err := filepath.EvalSymlinks(absRoot)
	if err != nil {
		return "", err
	}
	realPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(realRoot, realPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", errors.New("路径超出录像目录")
	}
	info, err := os.Stat(realPath)
	if err != nil || !info.Mode().IsRegular() || !isVideoFile(realPath) {
		return "", errors.New("录像文件不存在")
	}
	return filepath.Dir(realPath), nil
}
