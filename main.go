package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var doc = `用法：ciallo [选项]

一个基于终端的简易ASCII动画播放器。

选项：
  -h, --help                   显示此帮助信息
  -f, --fps <数值>             设置帧率（默认：20）
  -i, --input <文件路径>       指定输入JSON文件（默认：内嵌的ciallo.json）
  -s, --sound <文件路径>       指定音频文件路径（默认：/tmp/ciallo.wav）
  -t, --trigger <帧数>         设置音频触发帧（默认：14）
  -a, --audio                  启用音频播放（默认）
  -q, --quiet                  禁用音频播放

示例：
  ./ciallo                     使用默认设置播放
  ./ciallo -f 30 -i anim.json  30FPS播放自定义动画
  ./ciallo -q                  静音播放
  ./ciallo -t 0                从第一帧开始播放音频` + "\n\n\033[1m使用 Ctrl+C 来退出程序\033[0m"

//go:embed ciallo.json
var content []byte

//go:embed ciallo.wav
var CialloSoundFile embed.FS

var metaFrames [][]string
var frames [][][]byte
var delta [][][]byte
var CleanScreen = []byte("\033[2J")
var EnableSound = true

var CursorHidden = []byte("\033[?25l")
var CursorShown = []byte("\033[?25h")

var GoBack = []byte("\033[H")

var FPS int64
var SPF time.Duration
var SoundDelayTime time.Duration
var SoundTriggerFrame int64
var SoundSource string

var InputFile string

var channel = make(chan os.Signal, 1)

func processLine(line string) []byte {
	return []byte(line + "\n")
}

func init() {
	signal.Notify(channel, syscall.SIGINT)
	args := os.Args[1:]
	for i, arg := range args {
		arg = strings.ToLower(args[i])
		if arg == "-a" || arg == "--audio" {
			EnableSound = true
		} else if arg == "-q" || arg == "--quiet" {
			EnableSound = false
		} else if arg == "-h" || arg == "--help" {
			fmt.Println(doc)
			os.Exit(0)
		}
		if i+1 >= len(args) {
			continue
		}
		nextArg := args[i+1]
		switch arg {
		case "-f", "--fps":
			FPS, _ = strconv.ParseInt(nextArg, 10, 64)
		case "-i", "--input":
			InputFile = nextArg
		case "-s", "--sound":
			SoundSource = nextArg
		case "-t", "--trigger":
			SoundTriggerFrame, _ = strconv.ParseInt(nextArg, 10, 64)
		}
	}

	if FPS == 0 {
		FPS = 20
	}

	if SoundTriggerFrame == 0 {
		SoundTriggerFrame = 14
	}

	if SoundSource == "" {
		SoundSource = "/tmp/ciallo.wav"
	}

	SPF = time.Second / time.Duration(FPS)
	SoundDelayTime = SPF * time.Duration(SoundTriggerFrame)

	if InputFile != "" {
		temp, err := os.ReadFile(InputFile)
		if err == nil {
			content = temp
		}
	}

	err := json.Unmarshal(content, &metaFrames)
	if err != nil {
		panic(err)
	}

	frames = make([][][]byte, len(metaFrames))
	for i, frame := range metaFrames {
		temp := make([][]byte, len(frame))
		for j, line := range frame {
			temp[j] = processLine(line)
		}
		frames[i] = temp
	}

	delta = make([][][]byte, len(metaFrames))
	delta[0] = frames[0]
	for i := 1; i < len(metaFrames); i++ {
		currentFrame := metaFrames[i]
		prevFrame := metaFrames[i-1]
		var deltaLines [][]byte

		maxLines := len(currentFrame)
		if len(prevFrame) > maxLines {
			maxLines = len(prevFrame)
		}

		for j := 0; j < maxLines; j++ {
			currentLineExists := j < len(currentFrame)
			prevLineExists := j < len(prevFrame)

			var currentLine, prevLine string
			if currentLineExists {
				currentLine = currentFrame[j]
			}
			if prevLineExists {
				prevLine = prevFrame[j]
			}

			if currentLine != prevLine {
				cursorMove := fmt.Sprintf("\033[%d;1H", j+1)
				var lineContent []byte
				if currentLineExists {
					lineContent = append([]byte(cursorMove))
					lineContent = append(lineContent, processLine(currentLine)...)
				} else {
					lineContent = append([]byte(cursorMove))
					lineContent = append(lineContent, []byte("\n")...)
				}
				deltaLines = append(deltaLines, lineContent)
			}
		}

		delta[i] = deltaLines
	}

	extractEmbeddedSound()
}

func extractEmbeddedSound() {
	embeddedFile, err := CialloSoundFile.Open("ciallo.wav")
	if err != nil {
		return
	}
	defer func(embeddedFile fs.File) {
		err := embeddedFile.Close()
		if err != nil {
			return
		}
	}(embeddedFile)

	if err := os.MkdirAll(filepath.Dir(SoundSource), 0755); err != nil {
		return
	}

	localFile, err := os.Create(SoundSource)
	if err != nil {
		return
	}
	defer func(localFile *os.File) {
		err := localFile.Close()
		if err != nil {
			return
		}
	}(localFile)

	if _, err := io.Copy(localFile, embeddedFile); err != nil {
		return
	}

	if err := os.Chmod(SoundSource, 0644); err != nil {
		return
	}
}

func main() {
	go func() {
		sig := <-channel
		switch sig {
		case syscall.SIGINT:
			_, _ = os.Stdout.Write(CleanScreen)
			_, _ = os.Stdout.Write(GoBack)
			_, _ = os.Stdout.Write(CursorShown)
			_, _ = os.Stdout.Write([]byte(fmt.Sprintf("Ciallo～(∠・ω< )⌒★ @ %d FPS\n", FPS)))
			_ = os.Remove("/tmp/ciallo.wav")
			os.Exit(0)
		}
	}()

	_, _ = os.Stdout.Write(CursorHidden)
	_, _ = os.Stdout.Write(CleanScreen)
	_, _ = os.Stdout.Write(GoBack)

	//Only Update changed line
	for {
		go func() {
			if EnableSound {
				time.Sleep(SoundDelayTime)
				_ = exec.Command("ffplay", "-nodisp", "-autoexit", SoundSource).Run()
			}
		}()
		for i, frame := range delta {
			if i == 0 {
				_, _ = os.Stdout.Write(CleanScreen)
				_, _ = os.Stdout.Write(GoBack)
				for _, line := range frame {
					_, _ = os.Stdout.Write(line)
				}
			} else {
				for _, deltaLine := range frame {
					_, _ = os.Stdout.Write(deltaLine)
				}
			}
			time.Sleep(SPF)
		}
	}
}
