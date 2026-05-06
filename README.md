# ytt: Lightning-Fast YouTube Transcript CLI Tool

[![GitHub repo size](https://img.shields.io/github/repo-size/IFAKA/yt-transcription?style=flat-square)](https://github.com/IFAKA/yt-transcription)

**ytt** is a high-performance CLI utility for fetching and copying YouTube transcripts directly to your clipboard. Designed for speed and efficiency, it streamlines the process of extracting text from YouTube videos for researchers, developers, and content creators.

## 🚀 Features

- **Fast Transcript Fetching:** Optimized for speed, retrieving transcripts in seconds.
- **Direct-to-Clipboard:** Automatically copies fetched transcripts to your system clipboard.
- **Smart AI Integration:** After copying, easily open your favorite AI interface (Gemini, Claude, or ChatGPT) via an interactive CLI menu.
- **Interactive Navigation:** Use `j`/`k` or arrow keys to navigate the menu, `Enter` to select, and `Esc` to skip.
- **CLI-First Workflow:** Simple commands for terminal-based automation.
- **Lightweight & Fast:** Written in Go for maximum performance.

## 🛠 Installation

Install `ytt` with curl:

```bash
curl -fsSL https://raw.githubusercontent.com/IFAKA/yt-transcription/main/install.sh | sh
```

The installer builds the Go CLI and installs it as `/usr/local/bin/ytt`.
Make sure `/usr/local/bin` is in your `PATH`.

To install somewhere else:

```bash
curl -fsSL https://raw.githubusercontent.com/IFAKA/yt-transcription/main/install.sh | BIN_DIR="$HOME/.local/bin" sh
```

## 🧹 Uninstall

Uninstall `ytt` with curl:

```bash
curl -fsSL https://raw.githubusercontent.com/IFAKA/yt-transcription/main/uninstall.sh | sh
```

If you installed to a custom directory, pass the same `BIN_DIR`:

```bash
curl -fsSL https://raw.githubusercontent.com/IFAKA/yt-transcription/main/uninstall.sh | BIN_DIR="$HOME/.local/bin" sh
```

## 📖 Usage

Pass a YouTube URL to `ytt` to fetch and copy the transcript:

```bash
ytt https://www.youtube.com/watch?v=VIDEO_ID
```

After the transcript is copied, an interactive menu will appear:
```
Select an AI site to open (j/k to navigate, Enter to select, Esc to close):
> Gemini
  Claude
  ChatGPT
```

## 🎯 Keywords

YouTube transcript downloader, CLI tool, Go, Python, transcript automation, developer tools, YouTube API, text extraction, Gemini, Claude, ChatGPT.

## 📄 License

[MIT](LICENSE)
