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

Clone the repository and install the binary:

```bash
git clone https://github.com/IFAKA/yt-transcription.git
cd yt-transcription
go install ytt.go
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
