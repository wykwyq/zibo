# Bililive-go 桌面版

Windows 桌面壳使用 WebView2 显示 Bililive-go WebUI，并负责自动启动和停止后端服务。

## 构建

需要 Go 1.25+、Node.js 18+、Yarn、GNU Make、.NET 10 SDK 和 WebView2 Runtime。

```powershell
powershell -ExecutionPolicy Bypass -File .\desktop\publish.ps1
```

默认输出目录为 `desktop-dist\BililiveGo-Desktop`。运行其中的 `BililiveGoDesktop.exe` 即可启动桌面版。

## 行为

- 后端未运行时自动启动本地服务。
- 后端已运行时复用 `http://127.0.0.1:28080`，并核对是否为当前桌面版的后端。
- 关闭窗口后最小化到系统托盘。
- 托盘菜单中的“退出并停止服务”会先让后端完成录制收尾，再关闭桌面窗口；若后端未确认停止，保留后台进程，不强制截断录像。
- 默认录像目录为桌面版目录下的 `Videos`。
