# CialloFromTerminal

一个基于终端的简易ASCII动画播放器。

## 快速开始

### 直接运行（推荐）
```bash
# 下载并运行
go run main.go

# 或者编译后运行
go build -o ciallo
./ciallo
```


### 安装到系统 (可选)

```bash
# 编译并安装
go build -o ciallo
sudo cp ciallo /usr/local/bin/

# 现在可以在任何地方运行
ciallo
```


## 前置要求

### 音频支持（可选）

需要安装`ffplay`（来自ffmpeg包）：

• Ubuntu/Debian: `sudo apt install ffmpeg`

• CentOS/Fedora: `sudo dnf install ffmpeg`

• Archlinux: `sudo pacman -S ffmpeg`

• macOS: `brew install ffmpeg`

验证安装：`ffplay -version`

### 使用方法


```text
ciallo [选项]

选项：
  -h, --help           显示帮助信息
  -f, --fps <数值>     设置帧率（默认：20）
  -i, --input <文件>   指定动画JSON文件
  -s, --sound <文件>   指定音频文件（默认：./ciallo.wav）
  -t, --trigger <帧数> 音频触发帧（默认：14）
  -q, --quiet         静音播放

示例：
  ciallo              默认设置播放
  ciallo -f 30        30FPS播放
  ciallo -i anim.json 自定义动画
  ciallo -q           静音模式
  
使用Ctrl+C退出程序
```


### 动画文件格式

动画数据采用JSON格式，包含帧序列：
```json
[
    ["第一帧第一行", "第一帧第二行"],
    ["第二帧第一行", "第二帧第二行"]
]
```


程序支持差异帧渲染，但不支持读取差异帧的json文件，需要完整数据。

---

# 附：GIF转换工具

使用附带的Python脚本将GIF转换为JSON动画：

环境准备

```bash
python3 -m venv venv
source venv/bin/activate  # Linux/macOS
pip install -r requirements.txt
```

## 转换GIF

### 调整GIF分辨率
```bash
ffmpeg -i input.gif -vf scale=80:40 out.gif
```

**注意调整scale规定的长宽比**，终端输出时会因为字体长宽比与行间距导致高度拉伸，需要依据终端效果具体调整。

例子给出的 `80：40`即为 `宽：高`

### 转换为JSON
```bash
python3 tool.py
```

程序默认读取`./out.gif`和输出`./ciallo.json`如果需要更改可修改源码 6、7 两行内容

支持的颜色模式：gray（灰度）、16color、8bit、24bit（真彩色）
切换颜色模式可以在`tool.py`文件末尾取消注释对应模式并注释其他模式

```项目结构
.
├── ciallo          # 编译后的可执行文件
├── main.go         # 主程序源码
├── tool.py         # GIF转换脚本
├── ciallo.json     # 动画数据
├── ciallo.wav      # 音频文件
└── requirements.txt # Python依赖
```


注意事项

• 程序默认使用内嵌动画、音频数据，无需外部文件

• 编译后想要更改程序内默认的动画与音频需要重新编译

• 静音模式使用 -q 参数，无需音频依赖

• 支持运行时通过参数自定义动画和音频文件路径

• 按 Ctrl+C 退出程序

Ciallo～(∠・ω< )⌒★