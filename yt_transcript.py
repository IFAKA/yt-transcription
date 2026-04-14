#!/usr/bin/env python3
"""
yt_transcript.py — Fetch a YouTube transcript and copy it to clipboard.

Usage:
    ytt <youtube_url>
    ytt <youtube_url> --lang es
"""

import sys
import re
import subprocess
import argparse
import time
from concurrent.futures import ThreadPoolExecutor

try:
    from youtube_transcript_api import YouTubeTranscriptApi
    from youtube_transcript_api._errors import NoTranscriptFound, TranscriptsDisabled
except ImportError:
    print("Missing dependency. Run: pip3 install youtube-transcript-api --break-system-packages")
    sys.exit(1)


def extract_video_id(url: str) -> str:
    patterns = [
        r"(?:v=)([a-zA-Z0-9_-]{11})",
        r"(?:youtu\.be/)([a-zA-Z0-9_-]{11})",
        r"(?:shorts/)([a-zA-Z0-9_-]{11})",
        r"(?:embed/)([a-zA-Z0-9_-]{11})",
    ]
    for pattern in patterns:
        match = re.search(pattern, url)
        if match:
            return match.group(1)
    if re.fullmatch(r"[a-zA-Z0-9_-]{11}", url):
        return url
    raise ValueError(f"Could not extract video ID from: {url}")


def get_video_title(url: str, video_id: str) -> str:
    try:
        import urllib.request
        import json
        oembed_url = f"https://www.youtube.com/oembed?url=https://www.youtube.com/watch?v={video_id}&format=json"
        with urllib.request.urlopen(oembed_url, timeout=5) as resp:
            data = json.loads(resp.read())
            return data.get('title', video_id)
    except Exception:
        return video_id


def clean_transcript(snippets) -> str:
    parts = []
    for snippet in snippets:
        text = snippet.text
        # Strip bracketed sound tags like [Music], [Applause], [Laughter]
        text = re.sub(r'\[[^\]]+\]', '', text)
        text = text.strip()
        if text:
            parts.append(text)
    return ' '.join(parts)


def copy_to_clipboard(text: str) -> None:
    subprocess.run(["pbcopy"], input=text.encode("utf-8"), check=True)


def main():
    parser = argparse.ArgumentParser(
        description="Fetch a YouTube transcript and copy it to clipboard."
    )
    parser.add_argument("url", help="YouTube video URL or video ID")
    parser.add_argument(
        "--lang", default="en", metavar="LANG",
        help="Language code for transcript (default: en)"
    )
    parser.add_argument(
        "--profile", action="store_true",
        help="Print timing breakdown"
    )
    args = parser.parse_args()

    url = args.url
    lang = args.lang
    profile = args.profile

    t0 = time.perf_counter()

    try:
        video_id = extract_video_id(url)
    except ValueError as e:
        print(f"Error: {e}")
        sys.exit(1)

    title_times = {}
    transcript_times = {}

    def fetch_title():
        t = time.perf_counter()
        result = get_video_title(url, video_id)
        title_times['elapsed'] = time.perf_counter() - t
        return result

    def fetch_transcript():
        t = time.perf_counter()
        api = YouTubeTranscriptApi()
        transcript_list = api.list(video_id)
        try:
            transcript = transcript_list.find_transcript([lang])
        except NoTranscriptFound:
            available = [t.language_code for t in transcript_list]
            raise NoTranscriptFound(video_id, [lang], available, [])
        snippets = transcript.fetch()
        transcript_times['elapsed'] = time.perf_counter() - t
        return snippets

    with ThreadPoolExecutor(max_workers=2) as executor:
        title_future = executor.submit(fetch_title)
        transcript_future = executor.submit(fetch_transcript)

        try:
            snippets = transcript_future.result()
        except TranscriptsDisabled:
            print("Transcripts are disabled for this video.")
            sys.exit(1)
        except NoTranscriptFound:
            print(f"No transcript found for language '{lang}'.")
            sys.exit(1)
        except Exception as e:
            print(f"Failed to fetch transcript: {e}")
            sys.exit(1)

        title = title_future.result()

    t_total = time.perf_counter() - t0

    if profile:
        print(f"  title fetch:      {title_times.get('elapsed', 0):.2f}s")
        print(f"  transcript fetch: {transcript_times.get('elapsed', 0):.2f}s")
        print(f"  total wall time:  {t_total:.2f}s")

    print(f'Fetched: "{title}"')

    body = clean_transcript(snippets)

    output = f"# {title}\nSource: {url}\n\n{body}"

    copy_to_clipboard(output)

    word_count = len(body.split())
    token_estimate = int(word_count * 1.33)
    print(f"Copied ~{word_count:,} words (~{token_estimate:,} tokens) to clipboard.")


if __name__ == "__main__":
    main()
