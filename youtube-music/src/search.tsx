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
  getSearchOptions,
  getArtistsByIdOptions,
  getAlbumsByIdOptions,
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
      <SearchView />
    </QueryProvider>
  );
}

function SearchView() {
  const [query, setQuery] = useState("");
  const [filter, setFilter] = useState("songs");

  const { data, isLoading, error } = useQuery({
    ...getSearchOptions({ query: { q: query, filter, limit: 20 } }),
    enabled: query.trim().length > 0,
  });

  const playMutation = useMutation({
    ...postPlayerPlayMutation(),
    onSuccess: () => showToast(Toast.Style.Success, "Playing track"),
    onError: (err) =>
      showToast(Toast.Style.Failure, "Failed to play", String(err)),
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
      searchBarPlaceholder="Search YouTube Music..."
      onSearchTextChange={setQuery}
      throttle
      searchBarAccessory={
        <List.Dropdown
          tooltip="Filter Type"
          storeValue={true}
          onChange={setFilter}
        >
          <List.Dropdown.Item title="Songs" value="songs" />
          <List.Dropdown.Item title="Albums" value="albums" />
          <List.Dropdown.Item title="Artists" value="artists" />
          <List.Dropdown.Item title="Playlists" value="playlists" />
          <List.Dropdown.Item title="Videos" value="videos" />
        </List.Dropdown>
      }
    >
      {data?.results?.map((result, i) => {
        const type = result.result_type;

        if (type === "song" && result.track) {
          return (
            <List.Item
              key={`song-${result.track.video_id}-${i}`}
              title={result.track.title ?? "Unknown"}
              subtitle={`${result.track.artist} • ${result.track.album}`}
              icon={result.track.thumbnail_url ?? Icon.Music}
              actions={
                <ActionPanel>
                  <Action
                    title="Play Now"
                    icon={Icon.Play}
                    onAction={() =>
                      playMutation.mutate({
                        body: { video_id: result.track!.video_id! },
                      })
                    }
                  />
                  <Action
                    title="Play Next"
                    icon={Icon.SkipForward}
                    shortcut={{ modifiers: ["ctrl", "shift"], key: "return" }}
                    onAction={() =>
                      playNextMutation.mutate({
                        body: { video_id: result.track!.video_id! },
                      })
                    }
                  />
                  <Action
                    title="Add to Queue"
                    icon={Icon.Plus}
                    onAction={() =>
                      queueMutation.mutate({
                        body: { video_id: result.track!.video_id! },
                      })
                    }
                  />
                </ActionPanel>
              }
            />
          );
        }

        if (type === "artist" && result.artist) {
          return (
            <List.Item
              key={`artist-${result.artist.browse_id}-${i}`}
              title={result.artist.name ?? "Unknown"}
              subtitle="Artist"
              icon={Icon.Person}
              actions={
                <ActionPanel>
                  <Action.Push
                    title="View Artist"
                    icon={Icon.Person}
                    target={
                      <QueryProvider>
                        <ArtistDetail browseId={result.artist!.browse_id!} />
                      </QueryProvider>
                    }
                  />
                </ActionPanel>
              }
            />
          );
        }

        if (type === "album" && result.album) {
          return (
            <List.Item
              key={`album-${result.album.browse_id}-${i}`}
              title={result.album.title ?? "Unknown"}
              subtitle={`${result.album.artist} • ${result.album.year ?? ""}`}
              icon={result.album.thumbnail_url ?? Icon.Music}
              actions={
                <ActionPanel>
                  <Action.Push
                    title="View Album"
                    icon={Icon.Music}
                    target={
                      <QueryProvider>
                        <AlbumDetail browseId={result.album!.browse_id!} />
                      </QueryProvider>
                    }
                  />
                </ActionPanel>
              }
            />
          );
        }

        if (type === "playlist" && result.playlist) {
          return (
            <List.Item
              key={`playlist-${result.playlist.id}-${i}`}
              title={result.playlist.title ?? "Unknown"}
              subtitle={`${result.playlist.author} • ${result.playlist.track_count} tracks`}
              icon={result.playlist.thumbnail_url ?? Icon.List}
            />
          );
        }

        return null;
      })}
    </List>
  );
}

function ArtistDetail({ browseId }: { browseId: string }) {
  const { data, isLoading } = useQuery(
    getArtistsByIdOptions({ path: { id: browseId } }),
  );

  const playMutation = useMutation({
    ...postPlayerPlayMutation(),
    onSuccess: () => showToast(Toast.Style.Success, "Playing track"),
    onError: (err) =>
      showToast(Toast.Style.Failure, "Failed to play", String(err)),
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
      navigationTitle={data?.name || "Artist"}
    >
      <List.Section title="Top Tracks">
        {data?.top_tracks?.map((track, i) => (
          <List.Item
            key={`track-${track.video_id}-${i}`}
            title={track.title ?? "Unknown"}
            subtitle={track.album}
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
      </List.Section>
      <List.Section title="Albums">
        {data?.albums?.map((album, i) => (
          <List.Item
            key={`album-${album.browse_id}-${i}`}
            title={album.title ?? "Unknown"}
            subtitle={album.year}
            icon={album.thumbnail_url ?? Icon.Music}
            actions={
              <ActionPanel>
                <Action.Push
                  title="View Album"
                  icon={Icon.Music}
                  target={
                    <QueryProvider>
                      <AlbumDetail browseId={album.browse_id!} />
                    </QueryProvider>
                  }
                />
              </ActionPanel>
            }
          />
        ))}
      </List.Section>
    </List>
  );
}

function AlbumDetail({ browseId }: { browseId: string }) {
  const { data, isLoading } = useQuery(
    getAlbumsByIdOptions({ path: { id: browseId } }),
  );

  const playMutation = useMutation({
    ...postPlayerPlayMutation(),
    onSuccess: () => showToast(Toast.Style.Success, "Playing track"),
    onError: (err) =>
      showToast(Toast.Style.Failure, "Failed to play", String(err)),
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
      navigationTitle={data?.title || "Album"}
    >
      {data?.tracks?.map((track, i) => (
        <List.Item
          key={`track-${track.video_id}-${i}`}
          title={track.title ?? "Unknown"}
          subtitle={track.artist ?? data?.artist}
          icon={track.thumbnail_url ?? data?.thumbnail_url ?? Icon.Music}
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
