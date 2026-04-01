import {
  Action,
  ActionPanel,
  Detail,
  Icon,
} from "@raycast/api";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useState, useRef } from "react";
import { initClient } from "./lib/client";
import { QueryProvider } from "./components/QueryProvider";
import {
  getPlayerStateOptions,
  getLyricsOptions,
} from "./generated/@tanstack/react-query.gen";
import { findCurrentLineIndex } from "./lib/lyrics";

function LyricsView() {
  const { data: playerState, isLoading: playerLoading } = useQuery({
    ...getPlayerStateOptions(),
    refetchInterval: 500,
  });

  const track = playerState?.current_track;
  const isPlaying = playerState?.is_playing && !playerState?.is_paused;
  const serverPositionMs = playerState?.current_position_ms || 0;

  const { data: lyrics, isLoading: lyricsLoading, error: lyricsError } = useQuery({
    ...getLyricsOptions({
      query: {
        track_name: track?.title || "",
        artist_name: track?.artist || "",
        album_name: track?.album,
      },
    }),
    enabled: !!track?.title && !!track?.artist,
    staleTime: 5 * 60 * 1000,
    retry: false,
  });

  const [currentLineIndex, setCurrentLineIndex] = useState(-1);
  const lastPositionTimeRef = useRef<number>(Date.now());
  const lastKnownPositionRef = useRef<number>(0);
  const currentTrackIdRef = useRef<string | undefined>(undefined);

  useEffect(() => {
    if (track?.video_id !== currentTrackIdRef.current) {
      currentTrackIdRef.current = track?.video_id;
      setCurrentLineIndex(-1);
      lastKnownPositionRef.current = 0;
      lastPositionTimeRef.current = Date.now();
    }
  }, [track?.video_id]);

  useEffect(() => {
    lastKnownPositionRef.current = serverPositionMs;
    lastPositionTimeRef.current = Date.now();
  }, [serverPositionMs]);

  useEffect(() => {
    if (!lyrics?.parsed_lyrics?.length) return;

    const updateCurrentLine = () => {
      const now = Date.now();
      const elapsedSinceLastUpdate = isPlaying ? now - lastPositionTimeRef.current : 0;
      const effectivePosition = lastKnownPositionRef.current + elapsedSinceLastUpdate;

      const lines = lyrics.parsed_lyrics || [];
      const newIndex = findCurrentLineIndex(lines, effectivePosition);

      setCurrentLineIndex((prev) => (prev !== newIndex ? newIndex : prev));
    };

    updateCurrentLine();

    const interval = setInterval(updateCurrentLine, 100);
    return () => clearInterval(interval);
  }, [isPlaying, lyrics?.parsed_lyrics, serverPositionMs]);

  if (playerLoading) {
    return <Detail isLoading />;
  }

  if (!track) {
    return (
      <Detail markdown="**No track playing**\n\nStart playing a song to see lyrics." />
    );
  }

  if (lyricsLoading) {
    return (
      <Detail
        isLoading
        markdown={`Loading lyrics for **${track.title}** by **${track.artist}**...`}
      />
    );
  }

  if (lyricsError || !lyrics) {
    return (
      <Detail
        markdown={`**No lyrics found**\n\nCouldn't find lyrics for "${track.title}" by ${track.artist}.\n\nTry searching for the song manually.`}
        actions={
          <ActionPanel>
            <Action.OpenInBrowser
              title="Search on Google"
              url={`https://www.google.com/search?q=${encodeURIComponent(`${track.title} ${track.artist} lyrics`)}`}
              icon={Icon.Globe}
            />
          </ActionPanel>
        }
      />
    );
  }

  const lines = lyrics.parsed_lyrics || [];
  const hasSyncedLyrics = !!lyrics.synced_lyrics && lines.length > 0;

  let markdown = `# ${track.title}\n\n**${track.artist}**${track.album ? ` — ${track.album}` : ""}\n\n---\n\n`;

  if (lyrics.instrumental) {
    markdown += "\n🎵 **Instrumental** 🎵\n";
  } else if (hasSyncedLyrics) {
    const contextLines = 2;
    const startIdx = Math.max(0, currentLineIndex - contextLines);
    const endIdx = Math.min(lines.length, currentLineIndex + contextLines + 5);

    if (startIdx > 0) {
      markdown += "*...*\n\n";
    }

    for (let i = startIdx; i < endIdx; i++) {
      const line = lines[i];
      const isCurrent = i === currentLineIndex;

      if (isCurrent) {
        markdown += `**${line.text}**\n\n`;
      } else {
        markdown += `${line.text}\n\n`;
      }
    }

    if (endIdx < lines.length) {
      markdown += "*...*\n\n";
    }
  } else if (lyrics.plain_lyrics) {
    markdown += lyrics.plain_lyrics
      .split("\n")
      .map((line) => line.trim())
      .filter((line) => line)
      .join("\n\n");
  }

  return (
    <Detail
      isLoading={lyricsLoading}
      markdown={markdown}
      metadata={
        <Detail.Metadata>
          <Detail.Metadata.Label
            title="Track"
            text={track.title}
            icon={Icon.Music}
          />
          <Detail.Metadata.Label
            title="Artist"
            text={track.artist}
            icon={Icon.Person}
          />
          {track.album && (
            <Detail.Metadata.Label
              title="Album"
              text={track.album}
              icon={Icon.Folder}
            />
          )}
          <Detail.Metadata.Separator />
          <Detail.Metadata.Label
            title="Synced"
            text={hasSyncedLyrics ? "Yes" : "No"}
            icon={hasSyncedLyrics ? Icon.Clock : Icon.Minus}
          />
          <Detail.Metadata.Label
            title="Source"
            text={lyrics.source || "LrcLib"}
            icon={Icon.Globe}
          />
        </Detail.Metadata>
      }
      actions={
        <ActionPanel>
          <Action.OpenInBrowser
            title="Search on LrcLib"
            url={`https://lrclib.net/search?q=${encodeURIComponent(`${track.title} ${track.artist}`)}`}
            icon={Icon.Globe}
          />
          <Action.OpenInBrowser
            title="Search on Google"
            url={`https://www.google.com/search?q=${encodeURIComponent(`${track.title} ${track.artist} lyrics`)}`}
            icon={Icon.Globe}
          />
          {lyrics.synced_lyrics && (
            <Action.CopyToClipboard
              title="Copy Synced Lyrics"
              content={lyrics.synced_lyrics}
              shortcut={{ modifiers: ["cmd"], key: "c" }}
            />
          )}
          {lyrics.plain_lyrics && (
            <Action.CopyToClipboard
              title="Copy Plain Lyrics"
              content={lyrics.plain_lyrics}
              shortcut={{ modifiers: ["cmd", "shift"], key: "c" }}
            />
          )}
        </ActionPanel>
      }
    />
  );
}

export default function Command() {
  const [initialized, setInitialized] = useState(false);

  useEffect(() => {
    initClient().then(() => setInitialized(true));
  }, []);

  if (!initialized) return <Detail isLoading />;

  return (
    <QueryProvider>
      <LyricsView />
    </QueryProvider>
  );
}
