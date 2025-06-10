// 导入Wails运行时
import { SelectAudioFile, SelectOutputPath, ProcessAudio, CheckFFmpeg, GetAudioInfo, ShowErrorDialog, ShowInfoDialog } from '../wailsjs/go/main/App.js';
import { EventsOn, Quit, WindowMinimise, WindowToggleMaximise } from '../wailsjs/runtime/runtime.js';


document.addEventListener('DOMContentLoaded', () => {
    const closeBtn = document.getElementById('closeBtn');
    const minBtn = document.getElementById('minBtn');
    const maxBtn = document.getElementById('maxBtn');

    if (closeBtn && minBtn && maxBtn) {
        // 关闭按钮事件
        closeBtn.addEventListener('click', () => {
            //调用 Wails API 退出应用
            Quit();
        });

        // 最小化按钮事件
        minBtn.addEventListener('click', () => {
            // 调用 Wails API 最小化窗口
            WindowMinimise();
        });

        // 最大化/还原按钮事件
        maxBtn.addEventListener('click', () => {
            // 调用 Wails API 切换窗口最大化状态
            WindowToggleMaximise();
        });
    } else {
        console.error("Control buttons not found!");
    }
});

class AudioPitchShifter {
    constructor() {
        this.selectedFile = null;
        this.outputPath = '';
        this.isProcessing = false;
        this.fileInfo = null;

        this.initializeElements();
        this.bindEvents();
        this.checkFFmpegAvailability();
    }

    initializeElements() {
        // 获取DOM元素
        this.dropZone = document.getElementById('dropZone');
        this.fileInfo = document.getElementById('fileInfo');
        this.fileName = document.getElementById('fileName');
        this.fileSize = document.getElementById('fileSize');
        this.selectFileBtn = document.getElementById('selectFileBtn');
        this.changeFileBtn = document.getElementById('changeFileBtn');

        this.semitonesSlider = document.getElementById('semitonesSlider');
        this.semitonesValue = document.getElementById('semitonesValue');
        this.preserveTempo = document.getElementById('preserveTempo');
        this.quickButtons = document.querySelectorAll('.quick-btn');

        this.outputPathInput = document.getElementById('outputPath');
        this.selectOutputBtn = document.getElementById('selectOutputBtn');

        this.processBtn = document.getElementById('processBtn');
        this.processText = document.getElementById('processText');
        this.loadingSpinner = document.getElementById('loadingSpinner');

        this.progressSection = document.getElementById('progressSection');
        this.progressFill = document.getElementById('progressFill');
        this.progressText = document.getElementById('progressText');

        this.logOutput = document.getElementById('logOutput');
    }

    bindEvents() {
        // 文件选择事件
        this.selectFileBtn.addEventListener('click', () => this.selectFile());
        this.changeFileBtn.addEventListener('click', () => this.selectFile());

        // 拖拽事件
        this.dropZone.addEventListener('dragover', (e) => this.handleDragOver(e));
        this.dropZone.addEventListener('dragleave', (e) => this.handleDragLeave(e));
        this.dropZone.addEventListener('drop', (e) => this.handleDrop(e));

        // 滑块事件
        this.semitonesSlider.addEventListener('input', (e) => this.updateSemitonesValue(e));

        // 快捷按钮事件
        this.quickButtons.forEach(btn => {
            btn.addEventListener('click', (e) => this.setQuickValue(e));
        });

        // 输出路径选择
        this.selectOutputBtn.addEventListener('click', () => this.selectOutputPath());

        // 处理按钮
        this.processBtn.addEventListener('click', () => this.processAudio());

        // 监听Wails事件
        EventsOn('file-dropped', (filePath) => this.handleFileDropped(filePath));
        EventsOn('processing-started', () => this.onProcessingStarted());
        EventsOn('processing-finished', () => this.onProcessingFinished());
    }

    async checkFFmpegAvailability() {
        try {
            const result = await CheckFFmpeg();
            if (!result.success) {
                this.logMessage('错误: ' + result.message, 'error');
                ShowErrorDialog('FFmpeg未找到', result.message + '\n\n请下载FFmpeg并将其添加到系统PATH中。');
            } else {
                this.logMessage('FFmpeg检查通过', 'success');
            }
        } catch (error) {
            this.logMessage('FFmpeg检查失败: ' + error.message, 'error');
        }
    }

    async selectFile() {
        try {
            const fileInfo = await SelectAudioFile();
            this.setSelectedFile(fileInfo);
        } catch (error) {
            console.error('文件选择失败:', error);
        }
    }

    setSelectedFile(fileInfo) {
        this.selectedFile = fileInfo;
        this.fileName.textContent = `文件名: ${fileInfo.name}`;
        this.fileSize.textContent = `大小: ${this.formatFileSize(fileInfo.size)}`;

        this.dropZone.style.display = 'none';
        this.fileInfo.style.display = 'flex';

        this.generateDefaultOutputPath();
        this.updateProcessButton();

        this.logMessage(`已选择文件: ${fileInfo.name}`, 'info');

        // 获取音频信息
        this.getAudioInfo(fileInfo.path);
    }

    async getAudioInfo(filePath) {
        try {
            const result = await GetAudioInfo(filePath);
            if (result.success) {
                this.logMessage('音频信息获取成功', 'success');
                // 这里可以解析JSON并显示更多音频信息
            }
        } catch (error) {
            console.error('获取音频信息失败:', error);
        }
    }

    generateDefaultOutputPath() {
        if (!this.selectedFile) return;

        const fileName = this.selectedFile.name;
        const lastDotIndex = fileName.lastIndexOf('.');
        const baseName = lastDotIndex > 0 ? fileName.substring(0, lastDotIndex) : fileName;
        const extension = lastDotIndex > 0 ? fileName.substring(lastDotIndex) : '.mp3';

        const semitones = parseFloat(this.semitonesSlider.value);
        const semitonesText = semitones > 0 ? `+${semitones}` : semitones.toString();

        const defaultName = `${baseName}_${semitonesText}semitones${extension}`;
        this.outputPathInput.value = this.selectedFile.path + defaultName;
    }

    async selectOutputPath() {
        if (!this.selectedFile) {
            ShowErrorDialog('错误', '请先选择输入文件');
            return;
        }

        try {
            const defaultName = this.outputPathInput.value || 'output.mp3';
            const outputPath = await SelectOutputPath(defaultName);
            this.outputPath = outputPath;
            this.outputPathInput.value = outputPath;
            this.updateProcessButton();
        } catch (error) {
            console.error('输出路径选择失败:', error);
        }
    }

    handleDragOver(e) {
        e.preventDefault();
        this.dropZone.classList.add('drag-over');
    }

    handleDragLeave(e) {
        e.preventDefault();
        this.dropZone.classList.remove('drag-over');
    }

    handleDrop(e) {
        e.preventDefault();
        this.dropZone.classList.remove('drag-over');

        const files = e.dataTransfer.files;
        if (files.length > 0) {
            const file = files[0];
            this.handleFileDropped(file.path);
        }
    }

    handleFileDropped(filePath) {
        // 模拟文件信息
        const fileName = filePath.split(/[\\/]/).pop();
        const fileInfo = {
            name: fileName,
            path: filePath,
            size: 0 // 实际大小需要从Go端获取
        };

        this.setSelectedFile(fileInfo);
    }

    updateSemitonesValue(e) {
        const value = parseInt(e.target.value);
        this.semitonesValue.textContent = value > 0 ? `+${value}` : value.toString();

        // 更新快捷按钮状态
        this.quickButtons.forEach(btn => {
            const btnValue = parseFloat(btn.dataset.value);
            btn.classList.toggle('active', btnValue === value);
        });

        // 更新默认输出路径
        this.generateDefaultOutputPath();
    }

    setQuickValue(e) {
        const value = parseFloat(e.target.dataset.value);
        this.semitonesSlider.value = value;
        this.updateSemitonesValue({ target: this.semitonesSlider });
    }

    updateProcessButton() {
        const canProcess = this.selectedFile &&
            (this.outputPath || this.outputPathInput.value) &&
            !this.isProcessing;
        this.processBtn.disabled = !canProcess;
    }

    async processAudio() {
        if (!this.selectedFile) {
            ShowErrorDialog('错误', '请选择输入文件');
            return;
        }

        const outputPath = this.outputPath || this.outputPathInput.value;
        if (!outputPath) {
            ShowErrorDialog('错误', '请选择输出路径');
            return;
        }

        const semitones = parseFloat(this.semitonesSlider.value);
        const preserveTempo = this.preserveTempo.checked;

        this.isProcessing = true;
        this.updateProcessButton();
        this.showProgress();

        this.logMessage(`开始处理音频...`, 'info');
        this.logMessage(`输入文件: ${this.selectedFile.path}`, 'info');
        this.logMessage(`输出文件: ${outputPath}`, 'info');
        this.logMessage(`变调: ${semitones > 0 ? '+' : ''}${semitones} 半音`, 'info');
        this.logMessage(`保持节拍: ${preserveTempo ? '是' : '否'}`, 'info');

        try {
            const result = await ProcessAudio(
                this.selectedFile.path,
                outputPath,
                semitones,
                preserveTempo
            );

            if (result.success) {
                this.logMessage('音频处理完成!', 'success');
                ShowInfoDialog('成功', '音频处理完成!\n\n输出文件: ' + outputPath);
            } else {
                this.logMessage('处理失败: ' + result.message, 'error');
                ShowErrorDialog('处理失败', result.message);
            }

            if (result.output) {
                this.logMessage('FFmpeg输出:', 'info');
                this.logMessage(result.output, 'output');
            }

        } catch (error) {
            this.logMessage('处理错误: ' + error.message, 'error');
            ShowErrorDialog('错误', '处理过程中发生错误: ' + error.message);
        } finally {
            this.isProcessing = false;
            this.updateProcessButton();
            this.hideProgress();
        }
    }

    onProcessingStarted() {
        this.processText.textContent = '处理中...';
        this.loadingSpinner.style.display = 'inline-block';
        this.simulateProgress();
    }

    onProcessingFinished() {
        this.processText.textContent = '开始处理';
        this.loadingSpinner.style.display = 'none';
        this.progressFill.style.width = '100%';
        setTimeout(() => this.hideProgress(), 1000);
    }

    showProgress() {
        this.progressSection.style.display = 'block';
        this.progressFill.style.width = '0%';
        this.progressText.textContent = '处理中...';
    }

    hideProgress() {
        this.progressSection.style.display = 'none';
        this.progressFill.style.width = '0%';
    }

    simulateProgress() {
        let progress = 0;
        const interval = setInterval(() => {
            if (!this.isProcessing) {
                clearInterval(interval);
                return;
            }

            progress += Math.random() * 10;
            if (progress > 90) progress = 90;

            this.progressFill.style.width = progress + '%';
        }, 200);
    }

    logMessage(message, type = 'info') {
        const timestamp = new Date().toLocaleTimeString();
        const prefix = type === 'error' ? '[错误]' :
            type === 'success' ? '[成功]' :
                type === 'output' ? '[输出]' : '[信息]';

        const logEntry = `${timestamp} ${prefix} ${message}\n`;
        this.logOutput.textContent += logEntry;
        this.logOutput.scrollTop = this.logOutput.scrollHeight;
    }

    formatFileSize(bytes) {
        if (bytes === 0) return '未知';

        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));

        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }
}

// 页面加载完成后初始化应用
document.addEventListener('DOMContentLoaded', () => {
    new AudioPitchShifter();
});