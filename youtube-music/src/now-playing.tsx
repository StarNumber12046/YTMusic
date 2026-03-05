import {
  Action,
  ActionPanel,
  Detail,
  Icon,
  Toast,
  showToast,
} from "@raycast/api";
import { useQuery, useMutation } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { initClient } from "./lib/client";
import { QueryProvider } from "./components/QueryProvider";
import {
  getPlayerStateOptions,
  postPlayerPauseMutation,
  postPlayerNextMutation,
  postPlayerPreviousMutation,
  postPlayerStopMutation,
  postPlayerShuffleMutation,
  postPlayerRepeatMutation,
} from "./generated/@tanstack/react-query.gen";

export default function Command() {
  const [initialized, setInitialized] = useState(false);

  useEffect(() => {
    initClient().then(() => setInitialized(true));
  }, []);

  if (!initialized) return <Detail isLoading />;

  return (
    <QueryProvider>
      <NowPlaying />
    </QueryProvider>
  );
}

function NowPlaying() {
  const { data, isLoading, error } = useQuery({
    ...getPlayerStateOptions(),
    refetchInterval: 2000,
  });

  const pauseMutation = useMutation({
    ...postPlayerPauseMutation(),
    onError: (err) =>
      showToast(
        Toast.Style.Failure,
        "Failed to toggle play/pause",
        String(err),
      ),
  });

  const nextMutation = useMutation({
    ...postPlayerNextMutation(),
    onError: (err) =>
      showToast(Toast.Style.Failure, "Failed to skip track", String(err)),
  });

  const prevMutation = useMutation({
    ...postPlayerPreviousMutation(),
    onError: (err) =>
      showToast(Toast.Style.Failure, "Failed to go back", String(err)),
  });

  const stopMutation = useMutation({
    ...postPlayerStopMutation(),
    onError: (err) =>
      showToast(Toast.Style.Failure, "Failed to stop playback", String(err)),
  });

  const shuffleMutation = useMutation({
    ...postPlayerShuffleMutation(),
    onError: (err) =>
      showToast(Toast.Style.Failure, "Failed to toggle shuffle", String(err)),
  });

  const repeatMutation = useMutation({
    ...postPlayerRepeatMutation(),
    onError: (err) =>
      showToast(Toast.Style.Failure, "Failed to set repeat", String(err)),
  });

  const cycleRepeat = () => {
    const current = data?.repeat || "off";
    const next = current === "off" ? "all" : current === "all" ? "one" : "off";
    repeatMutation.mutate({ body: { repeat: next } });
  };

  if (error) {
    return (
      <Detail
        markdown={`**Error loading player state**\n\n\`\`\`\n${error}\n\`\`\`\n\nMake sure the API server is running and you are logged in.`}
      />
    );
  }

  const track = data?.current_track;
  const isPlaying = data?.is_playing && !data?.is_paused;

  const markdown = track
    ? `![](${track.thumbnail_url})\n\n# ${track.title}\n\n**${track.artist}** — ${track.album}\n\n${track.duration}`
    : "No track playing.";

  return (
    <Detail
      isLoading={isLoading}
      markdown={markdown}
      metadata={
        <Detail.Metadata>
          <Detail.Metadata.Label
            title="Status"
            text={
              isPlaying ? "Playing" : data?.is_paused ? "Paused" : "Stopped"
            }
            icon={
              isPlaying ? Icon.Play : data?.is_paused ? Icon.Pause : Icon.Stop
            }
          />
          <Detail.Metadata.Label
            title="Volume"
            text={`${data?.volume ?? 0}%`}
            icon={Icon.SpeakerUp}
          />
          {data?.queue_length ? (
            <>
              <Detail.Metadata.Label
                title="Queue"
                text={`${(data.queue_position ?? 0) + 1} / ${data.queue_length}`}
                icon={Icon.List}
              />
              <Detail.Metadata.Label
                title="Shuffle"
                text={data?.shuffle ? "On" : "Off"}
                icon={data?.shuffle ? Icon.ArrowUp : Icon.ArrowDown}
              />
              <Detail.Metadata.Label
                title="Repeat"
                text={
                  data?.repeat === "one"
                    ? "One"
                    : data?.repeat === "all"
                      ? "All"
                      : "Off"
                }
                icon={
                  data?.repeat === "one"
                    ? Icon.Repeat
                    : data?.repeat === "all"
                      ? Icon.Repeat
                      : Icon.List
                }
              />
            </>
          ) : null}
        </Detail.Metadata>
      }
      actions={
        <ActionPanel>
          <Action
            title={isPlaying ? "Pause" : "Play"}
            icon={isPlaying ? Icon.Pause : Icon.Play}
            onAction={() => pauseMutation.mutate()}
          />
          <Action
            title="Next Track"
            icon={Icon.Forward}
            onAction={() => nextMutation.mutate()}
          />
          <Action
            title="Previous Track"
            icon={Icon.Rewind}
            onAction={() => prevMutation.mutate()}
          />
          <Action
            title="Stop"
            icon={Icon.Stop}
            onAction={() => stopMutation.mutate()}
          />
          <Action
            title={data?.shuffle ? "Disable Shuffle" : "Enable Shuffle"}
            icon={data?.shuffle ? Icon.ArrowUp : Icon.ArrowDown}
            onAction={() => shuffleMutation.mutate()}
          />
          <Action
            title={`Repeat: ${data?.repeat === "one" ? "One" : data?.repeat === "all" ? "All" : "Off"}`}
            icon={Icon.Repeat}
            onAction={cycleRepeat}
          />
        </ActionPanel>
      }
    />
  );
}
