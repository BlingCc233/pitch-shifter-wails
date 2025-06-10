# 🎵 音频变调器 (Pitch Shifter)

> 一款基于 Wails 框架、仅适用于 Windows 平台的无损音频升降调处理工具

---

## 目录

- [功能特点](#功能特点)
- [系统需求](#系统需求)
- [下载Release](#Release)
- [安装与启动](#安装与启动)
- [License](#license)

---

## 功能特点

- 🎼 支持单文件无损变调（-12 ~ +12 半音）
- ⏱ 可选“保持节拍不变”或 “联动改变播放速度”
- 🚀 支持快捷按钮一键 +/-1 半音、±6 半音、±12 半音
- 📂 拖拽 & 文件选择 & 自定义输出路径
- 📊 实时进度条 & 处理日志
- 🟢 原生 Windows 桌面应用（Wails + Go + 前端）


---

## 系统需求

- Windows 10/11 (64-bit)
- Go 1.24+
- Node.js 22+ / npm 或 yarn
- Wails v2

---

## Release
 <img src="https://img.shields.io/github/v/release/BlingCc233/pitch-shifter-wails?color=pink&include_prereleases&style=for-the-badge" alt="release">


---

## 安装与启动

1. 克隆仓库
   ```bash
   git clone https://github.com/BlingCc233/pitch-shifter-wails.git
   cd pitch-shifter-wails
    ```
2. 安装依赖
    ```bash
    go mod tidy
    npm install
    ```
   
3. FFmpeg要求
   - 项目根目录下新建`bin`目录
   - 下载 FFmpeg 可执行文件并放入 `bin` 目录
   
3. 安装后端依赖 & 编译
    ```bash
    wails build -platform windows/amd64
    ```
### 使用说明

1. 选择音频文件
2. 设置变调参数
   - 半音数（-12 ~ +12）
   - 是否保持节拍不变
3. 选择输出路径
4. 点击“开始处理”

### License

本项目采用 [MIT License](LICENSE) 许可协议，您可以自由使用、修改和分发代码，但请保留原作者信息。