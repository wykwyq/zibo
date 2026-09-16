using System.Diagnostics;
using System.Security.Cryptography;
using System.Text.Json;
using Microsoft.Web.WebView2.Core;
using Microsoft.Web.WebView2.WinForms;

namespace BililiveGoDesktop;

internal static class Program
{
    [STAThread]
    private static void Main()
    {
        ApplicationConfiguration.Initialize();
        Application.Run(new MainForm());
    }
}

internal sealed class MainForm : Form
{
    private const string ServerUrl = "http://127.0.0.1:28080/";
    private readonly string _appDirectory = AppContext.BaseDirectory;
    private readonly WebView2 _webView = new() { Dock = DockStyle.Fill };
    private readonly NotifyIcon _trayIcon;
    private Process? _backend;
    private string? _shutdownToken;
    private bool _allowExit;

    public MainForm()
    {
        Text = "Bililive-go 桌面版";
        Width = 1280;
        Height = 820;
        MinimumSize = new Size(900, 600);
        StartPosition = FormStartPosition.CenterScreen;

        var iconPath = Path.Combine(_appDirectory, "favicon.ico");
        if (File.Exists(iconPath))
        {
            Icon = new Icon(iconPath);
        }

        Controls.Add(_webView);

        var menu = new ContextMenuStrip();
        menu.Items.Add("打开界面", null, (_, _) => ShowWindow());
        menu.Items.Add("刷新", null, (_, _) => _webView.Reload());
        menu.Items.Add(new ToolStripSeparator());
        menu.Items.Add("退出并停止服务", null, async (_, _) => await ExitApplicationAsync());

        _trayIcon = new NotifyIcon
        {
            Icon = Icon ?? SystemIcons.Application,
            Text = "Bililive-go 桌面版",
            ContextMenuStrip = menu,
            Visible = true,
        };
        _trayIcon.DoubleClick += (_, _) => ShowWindow();

        Shown += async (_, _) => await InitializeAsync();
        FormClosing += OnFormClosing;
        Resize += (_, _) =>
        {
            if (WindowState == FormWindowState.Minimized)
            {
                Hide();
            }
        };
    }

    private async Task InitializeAsync()
    {
        try
        {
            if (!await IsServerReadyAsync())
            {
                StartBackend();
            }

            await WaitForServerAsync();
            var dataDirectory = Path.Combine(_appDirectory, ".webview2");
            var environment = await CoreWebView2Environment.CreateAsync(userDataFolder: dataDirectory);
            await _webView.EnsureCoreWebView2Async(environment);
            _webView.Source = new Uri(ServerUrl);
        }
        catch (Exception exception)
        {
            MessageBox.Show(
                $"桌面版启动失败：{exception.Message}",
                "Bililive-go",
                MessageBoxButtons.OK,
                MessageBoxIcon.Error);
            await ExitApplicationAsync();
        }
    }

    private void StartBackend()
    {
        var backendPath = Path.Combine(_appDirectory, "backend", "bililive-go.exe");
        var configPath = Path.Combine(_appDirectory, "config.yml");
        if (!File.Exists(backendPath))
        {
            throw new FileNotFoundException("未找到后端程序。", backendPath);
        }

        var startInfo = new ProcessStartInfo
        {
            FileName = backendPath,
            WorkingDirectory = _appDirectory,
            UseShellExecute = false,
            CreateNoWindow = true,
            WindowStyle = ProcessWindowStyle.Hidden,
        };
        _shutdownToken = Convert.ToHexString(RandomNumberGenerator.GetBytes(32));
        startInfo.Environment["BILILIVE_DESKTOP_SHUTDOWN_TOKEN"] = _shutdownToken;
        startInfo.ArgumentList.Add("--config");
        startInfo.ArgumentList.Add(configPath);
        startInfo.ArgumentList.Add("--no-launcher");

        _backend = Process.Start(startInfo) ?? throw new InvalidOperationException("后端进程启动失败。");
    }

    private async Task<bool> IsServerReadyAsync()
    {
        try
        {
            using var client = new HttpClient { Timeout = TimeSpan.FromSeconds(1) };
            using var response = await client.GetAsync(ServerUrl + "api/info");
            if (!response.IsSuccessStatusCode)
            {
                return false;
            }

            using var info = JsonDocument.Parse(await response.Content.ReadAsStringAsync());
            var backendPath = Path.Combine(_appDirectory, "backend", "bililive-go.exe");
            return info.RootElement.GetProperty("app_name").GetString() == "BiliLive-go"
                && string.Equals(
                    Path.GetFullPath(info.RootElement.GetProperty("bgo_exe_path").GetString() ?? ""),
                    backendPath,
                    StringComparison.OrdinalIgnoreCase);
        }
        catch
        {
            return false;
        }
    }

    private async Task WaitForServerAsync()
    {
        for (var attempt = 0; attempt < 60; attempt++)
        {
            if (await IsServerReadyAsync())
            {
                return;
            }

            if (_backend?.HasExited == true)
            {
                throw new InvalidOperationException($"后端进程已退出，退出代码：{_backend.ExitCode}。");
            }

            await Task.Delay(500);
        }

        throw new TimeoutException("等待本地服务启动超时。");
    }

    private void ShowWindow()
    {
        Show();
        WindowState = FormWindowState.Normal;
        Activate();
    }

    private void OnFormClosing(object? sender, FormClosingEventArgs e)
    {
        if (_allowExit)
        {
            return;
        }

        e.Cancel = true;
        Hide();
        _trayIcon.ShowBalloonTip(2000, "Bililive-go", "程序正在后台运行，可从托盘菜单退出。", ToolTipIcon.Info);
    }

    private async Task ExitApplicationAsync()
    {
        if (_allowExit)
        {
            return;
        }
        _allowExit = true;

        if (_backend is { HasExited: false } && _shutdownToken is not null)
        {
            try
            {
                using var client = new HttpClient { Timeout = TimeSpan.FromSeconds(5) };
                using var request = new HttpRequestMessage(HttpMethod.Post, ServerUrl + "api/desktop/shutdown");
                request.Headers.Add("X-Desktop-Shutdown-Token", _shutdownToken);
                using var response = await client.SendAsync(request);
                response.EnsureSuccessStatusCode();
                using var timeout = new CancellationTokenSource(TimeSpan.FromSeconds(30));
                await _backend.WaitForExitAsync(timeout.Token);
            }
            catch (Exception exception)
            {
                MessageBox.Show(
                    $"录制服务尚未确认停止，已保留后台进程和录制文件。\n{exception.Message}",
                    "Bililive-go",
                    MessageBoxButtons.OK,
                    MessageBoxIcon.Warning);
            }
        }

        _trayIcon.Visible = false;
        _trayIcon.Dispose();
        Close();
    }
}
