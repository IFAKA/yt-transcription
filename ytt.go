package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

const innertubeAPIKey = "AIzaSyAO_FJ2SlqU8Q4STEHLGCilw_Y9_11qcW8"

type innertubeRequest struct {
	Context struct {
		Client struct {
			ClientName    string `json:"clientName"`
			ClientVersion string `json:"clientVersion"`
		} `json:"client"`
	} `json:"context"`
	VideoID string `json:"videoId"`
}

type playerResponse struct {
	Captions *struct {
		PlayerCaptionsTracklistRenderer struct {
			CaptionTracks []struct {
				BaseURL      string `json:"baseUrl"`
				LanguageCode string `json:"languageCode"`
				Name         struct {
					SimpleText string `json:"simpleText"`
				} `json:"name"`
			} `json:"captionTracks"`
		} `json:"playerCaptionsTracklistRenderer"`
	} `json:"captions"`
}

type transcriptJSON struct {
	Events []struct {
		Segs []struct {
			UTF8 string `json:"utf8"`
		} `json:"segs"`
	} `json:"events"`
}

var (
	videoIDPatterns = []*regexp.Regexp{
		regexp.MustCompile(`[?&]v=([A-Za-z0-9_-]{11})`),
		regexp.MustCompile(`youtu\.be/([A-Za-z0-9_-]{11})`),
		regexp.MustCompile(`/shorts/([A-Za-z0-9_-]{11})`),
		regexp.MustCompile(`/embed/([A-Za-z0-9_-]{11})`),
	}
	rawIDRe    = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)
	noiseRe    = regexp.MustCompile(`\[[^\]]*\]`)
	fmtParamRe = regexp.MustCompile(`([?&])fmt=[^&]*`)
	httpClient = &http.Client{Timeout: 15 * time.Second}
)

func extractVideoID(input string) (string, error) {
	for _, re := range videoIDPatterns {
		if m := re.FindStringSubmatch(input); m != nil {
			return m[1], nil
		}
	}
	if rawIDRe.MatchString(input) {
		return input, nil
	}
	return "", fmt.Errorf("could not extract video ID from: %s", input)
}

func formatNumber(n int) string {
	s := strconv.Itoa(n)
	out := make([]byte, 0, len(s)+3)
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, byte(c))
	}
	return string(out)
}

func cleanTranscript(raw string) string {
	cleaned := noiseRe.ReplaceAllString(raw, "")
	var sb strings.Builder
	prevSpace := true
	for _, r := range cleaned {
		if unicode.IsSpace(r) {
			if !prevSpace {
				sb.WriteRune(' ')
				prevSpace = true
			}
		} else {
			sb.WriteRune(r)
			prevSpace = false
		}
	}
	return strings.TrimSpace(sb.String())
}

func fetchPlayerAPI(videoID string) (*playerResponse, time.Duration, error) {
	var body innertubeRequest
	body.Context.Client.ClientName = "ANDROID"
	body.Context.Client.ClientVersion = "20.10.38"
	body.VideoID = videoID

	bodyBytes, _ := json.Marshal(body)

	start := time.Now()
	req, err := http.NewRequest("POST",
		"https://www.youtube.com/youtubei/v1/player?key="+innertubeAPIKey,
		bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	elapsed := time.Since(start)
	if err != nil {
		return nil, elapsed, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, elapsed, fmt.Errorf("player API returned HTTP %d", resp.StatusCode)
	}
	var pr playerResponse
	return &pr, elapsed, json.NewDecoder(resp.Body).Decode(&pr)
}

func fetchTranscript(baseURL string) (string, time.Duration, error) {
	url := fmtParamRe.ReplaceAllString(baseURL, "${1}fmt=json3")
	if url == baseURL {
		url = baseURL + "&fmt=json3"
	}

	start := time.Now()
	resp, err := httpClient.Get(url)
	elapsed := time.Since(start)
	if err != nil {
		return "", elapsed, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", elapsed, fmt.Errorf("transcript fetch returned HTTP %d", resp.StatusCode)
	}

	var tj transcriptJSON
	if err := json.NewDecoder(resp.Body).Decode(&tj); err != nil {
		return "", elapsed, err
	}

	var sb strings.Builder
	for _, ev := range tj.Events {
		for _, seg := range ev.Segs {
			sb.WriteString(seg.UTF8)
		}
	}
	return sb.String(), elapsed, nil
}

type statusReporter struct {
	mu      sync.Mutex
	writer  io.Writer
	message string
	stop    chan struct{}
	done    chan struct{}
}

func newStatusReporter(writer io.Writer) *statusReporter {
	return &statusReporter{writer: writer}
}

func (s *statusReporter) Start(message string) {
	s.message = message
	s.stop = make(chan struct{})
	s.done = make(chan struct{})
	go s.animate()
}

func (s *statusReporter) Update(message string) {
	s.mu.Lock()
	s.message = message
	s.mu.Unlock()
}

func (s *statusReporter) animate() {
	defer close(s.done)
	spinner := []string{"|", "/", "-", "\\"}
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	position := 0
	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			message := s.message
			s.mu.Unlock()
			fmt.Fprintf(s.writer, "\r\033[2K%s %s", spinner[position], message)
			position = (position + 1) % len(spinner)
		case <-s.stop:
			return
		}
	}
}

func (s *statusReporter) finish(symbol, message string) {
	close(s.stop)
	<-s.done
	fmt.Fprintf(s.writer, "\r\033[2K%s %s\n", symbol, message)
}

func (s *statusReporter) Success(message string) {
	s.finish("✓", message)
}

func (s *statusReporter) Failure(err error) {
	s.finish("✗", err.Error())
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: ytt <youtube_url_or_id> [--lang LANG] [--profile]")
		os.Exit(1)
	}

	lang := "en"
	langExplicit := false
	profile := false
	var input string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--lang":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "--lang requires a value")
				os.Exit(1)
			}
			i++
			lang = args[i]
			langExplicit = true
		case "--profile":
			profile = true
		default:
			if input != "" {
				fmt.Fprintln(os.Stderr, "unexpected argument:", args[i])
				os.Exit(1)
			}
			input = args[i]
		}
	}

	if input == "" {
		fmt.Fprintln(os.Stderr, "Usage: ytt <youtube_url_or_id> [--lang LANG] [--profile]")
		os.Exit(1)
	}

	totalStart := time.Now()
	status := newStatusReporter(os.Stderr)
	status.Start("Validating URL")
	validationStart := time.Now()
	videoID, err := extractVideoID(input)
	validationElapsed := time.Since(validationStart)
	if err != nil {
		status.Failure(err)
		os.Exit(1)
	}

	status.Update("Fetching transcript")
	playerStart := time.Now()
	pr, playerElapsed, err := fetchPlayerAPI(videoID)
	playerStepElapsed := time.Since(playerStart)
	if err != nil {
		status.Failure(fmt.Errorf("player API: %w", err))
		os.Exit(1)
	}
	if pr.Captions == nil {
		status.Failure(fmt.Errorf("no captions for this video"))
		os.Exit(1)
	}

	tracks := pr.Captions.PlayerCaptionsTracklistRenderer.CaptionTracks
	if len(tracks) == 0 {
		status.Failure(fmt.Errorf("transcripts disabled for this video"))
		os.Exit(1)
	}

	selectionStart := time.Now()
	var trackURL string
	selectedLang := lang
	for _, t := range tracks {
		if t.LanguageCode == lang {
			trackURL = t.BaseURL
			break
		}
	}
	if trackURL == "" {
		if !langExplicit {
			trackURL = tracks[0].BaseURL
			selectedLang = tracks[0].LanguageCode
		}
	}
	if trackURL == "" {
		available := make([]string, 0, len(tracks))
		for _, t := range tracks {
			available = append(available, fmt.Sprintf("%s (%s)", t.LanguageCode, t.Name.SimpleText))
		}
		status.Failure(fmt.Errorf("no transcript for language %q — available: %s", lang, strings.Join(available, ", ")))
		os.Exit(1)
	}
	if !langExplicit && selectedLang != lang {
		status.Update(fmt.Sprintf("Fetching transcript (using %q)", selectedLang))
	}
	selectionElapsed := time.Since(selectionStart)

	transcriptStart := time.Now()
	rawText, transcriptElapsed, err := fetchTranscript(trackURL)
	transcriptStepElapsed := time.Since(transcriptStart)
	if err != nil {
		status.Failure(fmt.Errorf("transcript fetch: %w", err))
		os.Exit(1)
	}

	processingStart := time.Now()
	text := cleanTranscript(rawText)
	words := len(strings.Fields(text))
	tokens := int(float64(words) * 1.333)
	processingElapsed := time.Since(processingStart)

	status.Update("Copying transcript")
	copyStart := time.Now()
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	if err := cmd.Run(); err != nil {
		status.Failure(fmt.Errorf("clipboard: %w", err))
		os.Exit(1)
	}
	copyElapsed := time.Since(copyStart)
	totalElapsed := time.Since(totalStart)
	status.Success(fmt.Sprintf("Copied ~%s words (~%s tokens) to clipboard", formatNumber(words), formatNumber(tokens)))

	if profile {
		fmt.Printf("  validation:        %.2fs\n", validationElapsed.Seconds())
		fmt.Printf("  player step:       %.2fs\n", playerStepElapsed.Seconds())
		fmt.Printf("  player API:        %.2fs\n", playerElapsed.Seconds())
		fmt.Printf("  transcript select: %.2fs\n", selectionElapsed.Seconds())
		fmt.Printf("  transcript step:   %.2fs\n", transcriptStepElapsed.Seconds())
		fmt.Printf("  transcript fetch:  %.2fs\n", transcriptElapsed.Seconds())
		fmt.Printf("  clean/count:       %.2fs\n", processingElapsed.Seconds())
		fmt.Printf("  clipboard copy:    %.2fs\n", copyElapsed.Seconds())
		fmt.Printf("  total wall time:  %.2fs\n", totalElapsed.Seconds())
	}
}
