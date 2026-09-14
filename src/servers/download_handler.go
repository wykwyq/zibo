package servers

import (
	"fmt"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gorilla/mux"

	"github.com/bililive-go/bililive-go/src/configs"
	"github.com/bililive-go/bililive-go/src/pkg/utils"
)

var mp4ConversionLocks sync.Map

// downloadFileAsMP4 下载录制文件；非 MP4 视频会先使用 FFmpeg 无损封装为 MP4。
func downloadFileAsMP4(writer http.ResponseWriter, request *http.Request) {
	cfg := configs.GetCurrentConfig()
	if cfg == nil {
		http.Error(writer, "配置尚未就绪", http.StatusServiceUnavailable)
		return
	}
	sourcePath, err := getSafePath(cfg.OutPutPath, mux.Vars(request)["path"])
	if err != nil {
		http.Error(writer, "无效或越权路径", http.StatusBadRequest)
		return
	}

	info, err := os.Stat(sourcePath)
	if err != nil || info.IsDir() || !isVideoFile(sourcePath) {
		http.Error(writer, "视频文件不存在", http.StatusNotFound)
		return
	}

	targetPath := sourcePath
	if !strings.EqualFold(filepath.Ext(sourcePath), ".mp4") {
		targetPath = strings.TrimSuffix(sourcePath, filepath.Ext(sourcePath)) + ".mp4"
		lock, _ := mp4ConversionLocks.LoadOrStore(sourcePath, &sync.Mutex{})
		mutex := lock.(*sync.Mutex)
		mutex.Lock()
		defer mutex.Unlock()

		if targetInfo, statErr := os.Stat(targetPath); statErr != nil || targetInfo.ModTime().Before(info.ModTime()) {
			if err := convertVideoToMP4(request, sourcePath, targetPath); err != nil {
				http.Error(writer, "转换 MP4 失败："+err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	serveMP4Download(writer, request, targetPath)
}

func convertVideoToMP4(request *http.Request, sourcePath, targetPath string) error {
	ffmpegPath, err := utils.GetFFmpegPath(request.Context())
	if err != nil {
		return fmt.Errorf("FFmpeg 不可用: %w", err)
	}

	tempFile, err := os.CreateTemp(filepath.Dir(targetPath), ".mp4-download-*.mp4")
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()
	if err := tempFile.Close(); err != nil {
		return err
	}
	defer os.Remove(tempPath)

	command := exec.CommandContext(
		request.Context(),
		ffmpegPath,
		"-i", sourcePath,
		"-c", "copy",
		"-movflags", "+faststart",
		"-y",
		tempPath,
	)
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}

	if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.Rename(tempPath, targetPath)
}

func serveMP4Download(writer http.ResponseWriter, request *http.Request, path string) {
	file, err := os.Open(path)
	if err != nil {
		http.Error(writer, "读取 MP4 文件失败", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		http.Error(writer, "读取 MP4 文件信息失败", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "video/mp4")
	writer.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
		"filename": filepath.Base(path),
	}))
	http.ServeContent(writer, request, filepath.Base(path), info.ModTime(), file)
}
