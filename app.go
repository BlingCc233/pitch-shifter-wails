package main

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"math" // 2. 引入 math 包以使用更高效的 Pow 函数
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

//go:embed bin/ffmpeg.exe
var ffmpegData []byte // 3. 嵌入 bin/ffmpeg.exe 文件数据

//go:embed bin/ffprobe.exe
var ffprobeData []byte // 4. 嵌入 bin/ffprobe.exe 文件数据

// App struct
type App struct {
	ctx         context.Context
	ffmpegPath  string // 5. 新增字段，用于存储 ffmpeg.exe 的运行时路径
	ffprobePath string // 6. 新增字段，用于存储 ffprobe.exe 的运行时路径
}

type ProcessResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Output  string `json:"output"`
}

type FileInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Size int64  `json:"size"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts.
// 在这里进行初始化，特别是提取 FFmpeg。
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	var err error

	// 提取 ffmpeg.exe
	a.ffmpegPath, err = extractExecutable(ctx, ffmpegData, "ffmpeg.exe")
	if err != nil {
		// 如果提取失败，记录错误并可能退出应用，因为核心功能依赖它
		runtime.LogErrorf(ctx, "无法提取 ffmpeg.exe: %v", err)
		runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "致命错误",
			Message: fmt.Sprintf("无法初始化 ffmpeg.exe: %v", err),
		})
		runtime.Quit(ctx)
		return
	}

	// 提取 ffprobe.exe
	a.ffprobePath, err = extractExecutable(ctx, ffprobeData, "ffprobe.exe")
	if err != nil {
		runtime.LogErrorf(ctx, "无法提取 ffprobe.exe: %v", err)
		runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "致命错误",
			Message: fmt.Sprintf("无法初始化 ffprobe.exe: %v", err),
		})
		runtime.Quit(ctx)
		return
	}

	runtime.LogInfof(ctx, "FFmpeg 路径设置为: %s", a.ffmpegPath)
	runtime.LogInfof(ctx, "FFprobe 路径设置为: %s", a.ffprobePath)
}

// extractExecutable 是一个辅助函数，用于将嵌入的二进制文件写入磁盘并返回其路径
// 它会将文件保存在用户配置目录下的一个应用专属文件夹中，以避免权限问题和污染系统
func extractExecutable(ctx context.Context, data []byte, name string) (string, error) {
	// 获取用户配置目录，这是一个存放应用配置的安全位置
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("无法获取用户配置目录: %w", err)
	}

	// 在配置目录中为我们的应用创建一个子目录
	// 请将 "YourAppName" 替换为您应用的真实名称，以避免冲突
	appDir := filepath.Join(configDir, "AudioPitchChanger")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", fmt.Errorf("无法创建应用目录 '%s': %w", appDir, err)
	}

	execPath := filepath.Join(appDir, name)

	// 检查文件是否已存在。如果存在，我们假设它是正确的，直接返回路径。
	// 在更复杂的应用中，您可能需要检查文件大小或哈希值来确定是否需要覆盖更新。
	if _, err := os.Stat(execPath); err == nil {
		runtime.LogInfof(ctx, "'%s' 已存在，跳过提取。", name)
		return execPath, nil
	}

	// 将嵌入的数据写入文件。权限 0755 确保文件是可执行的。
	runtime.LogInfof(ctx, "正在提取 '%s' 到 '%s'", name, execPath)
	err = os.WriteFile(execPath, data, 0755)
	if err != nil {
		return "", fmt.Errorf("无法写入可执行文件 '%s': %w", name, err)
	}

	return execPath, nil
}

func (a *App) domReady(ctx context.Context) {
	// 在DOM准备就绪时执行的操作
}

func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	return false
}

func (a *App) shutdown(ctx context.Context) {
	// 应用程序关闭时的清理操作
}

func (a *App) onFileDropped(ctx context.Context, x, y int, paths []string) {
	if len(paths) > 0 {
		runtime.EventsEmit(ctx, "file-dropped", paths[0])
	}
}

func (a *App) onSecondInstanceLaunch(secondInstanceData options.SecondInstanceData) {
	runtime.WindowUnminimise(a.ctx)
	runtime.WindowShow(a.ctx)
}

// SelectAudioFile 选择音频文件
func (a *App) SelectAudioFile() (FileInfo, error) {
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择音频文件",
		Filters: []runtime.FileFilter{
			{DisplayName: "音频文件", Pattern: "*.mp3;*.wav;*.flac;*.aac;*.ogg;*.m4a;*.wma"},
			{DisplayName: "所有文件", Pattern: "*.*"},
		},
	})
	if err != nil {
		return FileInfo{}, err
	}
	if filePath == "" {
		return FileInfo{}, fmt.Errorf("未选择文件")
	}
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return FileInfo{}, err
	}
	return FileInfo{Name: fileInfo.Name(), Path: filePath, Size: fileInfo.Size()}, nil
}

// SelectOutputPath 选择输出路径
func (a *App) SelectOutputPath(defaultName string) (string, error) {
	outputPath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "保存音频文件",
		DefaultFilename: defaultName,
		Filters: []runtime.FileFilter{
			{DisplayName: "音频文件", Pattern: "*.mp3;*.wav;*.flac;*.aac;*.ogg;*.m4a"},
		},
	})
	if err != nil {
		return "", err
	}
	if outputPath == "" {
		return "", fmt.Errorf("未选择输出路径")
	}
	return outputPath, nil
}

// CheckFFmpeg 检查FFmpeg是否可用
func (a *App) CheckFFmpeg() ProcessResult {
	// 7. 修改 CheckFFmpeg，使用 a.ffmpegPath
	if a.ffmpegPath == "" {
		return ProcessResult{
			Success: false,
			Message: "FFmpeg 路径未在启动时成功设置，请重启应用。",
		}
	}

	cmd := exec.Command(a.ffmpegPath, "-version")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return ProcessResult{
			Success: false,
			Message: "内嵌的 FFmpeg 执行失败",
			Output:  string(output),
		}
	}

	return ProcessResult{
		Success: true,
		Message: "内嵌的 FFmpeg 可用",
		Output:  string(output),
	}
}

func (a *App) getAudioSampleRate(inputPath string) (int, error) {
	// 使用 ffprobe 获取采样率
	// ffprobe -v error -select_streams a:0 -show_entries stream=sample_rate -of default=noprint_wrappers=1:nokey=1 [input]
	cmd := exec.Command(a.ffprobePath,
		"-v", "error",
		"-select_streams", "a:0",
		"-show_entries", "stream=sample_rate",
		"-of", "default=noprint_wrappers=1:nokey=1",
		inputPath,
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w, output: %s", err, string(output))
	}

	sampleRateStr := strings.TrimSpace(string(output))
	sampleRate, err := strconv.Atoi(sampleRateStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse sample rate: %w", err)
	}

	return sampleRate, nil
}

// ProcessAudio 处理音频变调
func (a *App) ProcessAudio(inputPath, outputPath string, semitones float64, preserveTempo bool) ProcessResult {
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return ProcessResult{Success: false, Message: "输入文件不存在"}
	}

	var args []string
	args = append(args, "-i", inputPath)
	args = append(args, "-y")

	if preserveTempo {
		pitchRatio := calculatePitchRatio(semitones)
		filterComplex := fmt.Sprintf("rubberband=pitch=%.6f:pitchq=quality", pitchRatio)
		args = append(args, "-af", filterComplex)
	} else {
		if semitones != 0 {
			ratio := calculatePitchRatio(semitones)
			args = append(args, "-filter:a", fmt.Sprintf("asetrate=44100*%f,aresample=44100", ratio))
		}
	}

	args = append(args, "-c:a")
	ext := strings.ToLower(filepath.Ext(outputPath))
	switch ext {
	case ".mp3":
		args = append(args, "libmp3lame", "-b:a", "320k")
	case ".flac":
		args = append(args, "flac")
	case ".wav":
		args = append(args, "pcm_s16le")
	case ".aac", ".m4a":
		args = append(args, "aac", "-b:a", "256k")
	case ".ogg":
		args = append(args, "libvorbis", "-b:a", "256k")
	default:
		args = append(args, "libmp3lame", "-b:a", "320k")
	}

	args = append(args, outputPath)

	// 8. 修改 ProcessAudio，使用 a.ffmpegPath
	cmd := exec.Command(a.ffmpegPath, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	runtime.EventsEmit(a.ctx, "processing-started")
	output, err := cmd.CombinedOutput()
	runtime.EventsEmit(a.ctx, "processing-finished")

	if err != nil {
		return ProcessResult{
			Success: false,
			Message: fmt.Sprintf("音频处理失败: %v", err),
			Output:  string(output),
		}
	}
	exec.Command("explorer", "/select,", outputPath).Start()

	return ProcessResult{
		Success: true,
		Message: "音频处理完成",
		Output:  string(output),
	}
}

// calculatePitchRatio 计算音调比率
func calculatePitchRatio(semitones float64) float64 {
	// 9. 使用标准库 math.Pow 替换自定义的 pow 函数
	return math.Pow(2.0, semitones/12.0)
}

// 10. 自定义的 pow 函数已被移除

// GetAudioInfo 获取音频文件信息
func (a *App) GetAudioInfo(filePath string) ProcessResult {
	// 11. 修改 GetAudioInfo，使用 a.ffprobePath
	cmd := exec.Command(a.ffprobePath, "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", filePath)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return ProcessResult{
			Success: false,
			Message: "无法获取音频信息",
			Output:  string(output),
		}
	}

	return ProcessResult{
		Success: true,
		Message: "音频信息获取成功",
		Output:  string(output),
	}
}

// ShowErrorDialog 显示错误对话框
func (a *App) ShowErrorDialog(title, message string) {
	runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    runtime.ErrorDialog,
		Title:   title,
		Message: message,
	})
}

// ShowInfoDialog 显示信息对话框
func (a *App) ShowInfoDialog(title, message string) {
	runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    runtime.InfoDialog,
		Title:   title,
		Message: message,
	})
}
