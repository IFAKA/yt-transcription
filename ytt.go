package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
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
	rawIDRe   = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)
	noiseRe   = regexp.MustCompile(`\[[^\]]*\]`)
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

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: ytt <youtube_url_or_id> [--lang LANG] [--profile]")
		os.Exit(1)
	}

	lang := "en"
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

	videoID, err := extractVideoID(input)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	totalStart := time.Now()

	pr, playerElapsed, err := fetchPlayerAPI(videoID)
	if err != nil {
		fmt.Fprintln(os.Stderr, "player API:", err)
		os.Exit(1)
	}
	if pr.Captions == nil {
		fmt.Fprintln(os.Stderr, "no captions for this video")
		os.Exit(1)
	}

	tracks := pr.Captions.PlayerCaptionsTracklistRenderer.CaptionTracks
	if len(tracks) == 0 {
		fmt.Fprintln(os.Stderr, "transcripts disabled for this video")
		os.Exit(1)
	}

	var trackURL string
	for _, t := range tracks {
		if t.LanguageCode == lang {
			trackURL = t.BaseURL
			break
		}
	}
	if trackURL == "" {
		fmt.Fprintf(os.Stderr, "no transcript for language %q — available:\n", lang)
		for _, t := range tracks {
			fmt.Fprintf(os.Stderr, "  %s  (%s)\n", t.LanguageCode, t.Name.SimpleText)
		}
		os.Exit(1)
	}

	rawText, transcriptElapsed, err := fetchTranscript(trackURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "transcript fetch:", err)
		os.Exit(1)
	}

	totalElapsed := time.Since(totalStart)
	text := cleanTranscript(rawText)
	words := len(strings.Fields(text))
	tokens := int(float64(words) * 1.333)

	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "clipboard:", err)
	}

	if profile {
		fmt.Printf("  player API:       %.2fs\n", playerElapsed.Seconds())
		fmt.Printf("  transcript fetch: %.2fs\n", transcriptElapsed.Seconds())
		fmt.Printf("  total wall time:  %.2fs\n", totalElapsed.Seconds())
	}
	fmt.Printf("Copied ~%s words (~%s tokens) to clipboard.\n",
		formatNumber(words), formatNumber(tokens))
}
