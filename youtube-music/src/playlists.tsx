import {
  Action,
  ActionPanel,
  List,
  Icon,
  Toast,
  showToast,
} from "@raycast/api";
import { useQuery, useMutation } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { initClient } from "./lib/client";
import { QueryProvider } from "./components/QueryProvider";
import {
  getPlaylistsOptions,
  getPlaylistsByIdOptions,
  postPlaylistsByIdPlayMutation,
  postPlaylistsByIdCacheMutation,
  postPlayerPlayMutation,
  postQueueAddMutation,
  postQueuePlayNextMutation,
} from "./generated/@tanstack/react-query.gen";

export default function Command() {
  const [initialized, setInitialized] = useState(false);

  useEffect(() => {
    initClient().then(() => setInitialized(true));
  }, []);

  if (!initialized) return <List isLoading />;

  return (
    <QueryProvider>
      <PlaylistsView />
    </QueryProvider>
  );
}

function PlaylistsView() {
  const { data, isLoading, error } = useQuery(getPlaylistsOptions());

  const playPlaylistMutation = useMutation({
    ...postPlaylistsByIdPlayMutation(),
    onSuccess: () => showToast(Toast.Style.Success, "Playing playlist"),
    onError: (err) =>
      showToast(Toast.Style.Failure, "Failed to play playlist", String(err)),
  });

  const cachePlaylistMutation = useMutation({
    ...postPlaylistsByIdCacheMutation(),
    onSuccess: () =>
      showToast(Toast.Style.Success, "Caching playlist in background"),
    onError: (err) =>
      showToast(
        Toast.Style.Failure,
        "Failed to queue playlist caching",
        String(err),
      ),
  });

  if (error) {
    showToast(Toast.Style.Failure, "Failed to load playlists", String(error));
  }

  return (
    <List isLoading={isLoading}>
      {data?.playlists?.map((pl) => (
        <List.Item
          key={pl.id}
          title={pl.title ?? "Unknown"}
          subtitle={`${pl.track_count} tracks`}
          icon={pl.thumbnail_url ?? Icon.List}
          actions={
            <ActionPanel>
              <Action.Push
                title="View Tracks"
                target={
                  <QueryProvider>
                    <PlaylistDetail id={pl.id!} />
                  </QueryProvider>
                }
              />
              <Action
                title="Play Playlist"
                icon={Icon.Play}
                onAction={() =>
                  playPlaylistMutation.mutate({ path: { id: pl.id! } })
                }
              />
              <Action
                title="Cache Playlist"
                icon={Icon.Download}
                onAction={() =>
                  cachePlaylistMutation.mutate({ path: { id: pl.id! } })
                }
              />
            </ActionPanel>
          }
        />
      ))}
    </List>
  );
}

function PlaylistDetail({ id }: { id: string }) {
  const { data, isLoading } = useQuery(
    getPlaylistsByIdOptions({ path: { id } }),
  );

  const playMutation = useMutation({
    ...postPlayerPlayMutation(),
    onSuccess: () => showToast(Toast.Style.Success, "Playing track"),
    onError: (err) =>
      showToast(Toast.Style.Failure, "Failed to play track", String(err)),
  });

  const queueMutation = useMutation({
    ...postQueueAddMutation(),
    onSuccess: () => showToast(Toast.Style.Success, "Added to queue"),
    onError: (err) =>
      showToast(Toast.Style.Failure, "Failed to add to queue", String(err)),
  });

  const playNextMutation = useMutation({
    ...postQueuePlayNextMutation(),
    onSuccess: () => showToast(Toast.Style.Success, "Playing next"),
    onError: (err) =>
      showToast(Toast.Style.Failure, "Failed to play next", String(err)),
  });

  return (
    <List
      isLoading={isLoading}
      navigationTitle={data?.playlist?.title || "Playlist Tracks"}
    >
      {data?.tracks?.map((track, i) => (
        <List.Item
          key={`${track.video_id}-${i}`}
          title={track.title ?? "Unknown"}
          subtitle={`${track.artist} • ${track.album}`}
          icon={track.thumbnail_url ?? Icon.Music}
          accessories={[{ text: track.duration }]}
          actions={
            <ActionPanel>
              <Action
                title="Play Now"
                icon={Icon.Play}
                onAction={() =>
                  playMutation.mutate({ body: { video_id: track.video_id! } })
                }
              />
              <Action
                title="Play Next"
                icon={Icon.SkipForward}
                shortcut={{ modifiers: ["ctrl", "shift"], key: "return" }}
                onAction={() =>
                  playNextMutation.mutate({ body: { video_id: track.video_id! } })
                }
              />
              <Action
                title="Add to Queue"
                icon={Icon.Plus}
                onAction={() =>
                  queueMutation.mutate({ body: { video_id: track.video_id! } })
                }
              />
            </ActionPanel>
          }
        />
      ))}
    </List>
  );
}
