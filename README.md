# ytt: Lightning-Fast YouTube Transcript CLI Tool

[![GitHub repo size](https://img.shields.io/github/repo-size/IFAKA/yt-transcription?style=flat-square)](https://github.com/IFAKA/yt-transcription)

**ytt** is a high-performance CLI utility for fetching and copying YouTube transcripts directly to your clipboard. Designed for speed and efficiency, it streamlines the process of extracting text from YouTube videos for researchers, developers, and content creators.

## 🚀 Features

- **Fast Transcript Fetching:** Optimized for speed, retrieving transcripts in seconds.
- **Direct-to-Clipboard:** Automatically copies fetched transcripts to your system clipboard.
- **Automatic Workflow:** Validates the input, fetches the transcript, copies it, and exits after reporting success or failure.
- **Live Status Feedback:** Shows compact progress states and final word/token counts in the terminal.
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

`ytt` accepts full YouTube URLs, short URLs, Shorts URLs, embed URLs, and raw 11-character video IDs. It prefers English transcripts by default and automatically uses the first transcript language YouTube provides when English is unavailable. Use `--lang LANG` to require a specific language:

```bash
ytt https://www.youtube.com/watch?v=VIDEO_ID --lang es
```

The command prints the elapsed time for each pipeline step by default, including URL validation, transcript selection, cleanup, and clipboard copy. The legacy `--profile` flag is still accepted:

```bash
ytt https://www.youtube.com/watch?v=VIDEO_ID --profile
```

The terminal reports the state-based flow as it runs:

```text
| Validating URL
/ Fetching transcript
- Copying transcript
✓ Copied ~1,234 words (~1,644 tokens) to clipboard
```

## 🎯 Keywords

YouTube transcript downloader, CLI tool, Go, Python, transcript automation, developer tools, YouTube API, text extraction.

## 📄 License

[MIT](LICENSE)
