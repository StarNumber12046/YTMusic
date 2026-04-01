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
  getQueueOptions,
  deleteQueueMutation,
  deleteQueueByPositionMutation,
} from "./generated/@tanstack/react-query.gen";

export default function Command() {
  const [initialized, setInitialized] = useState(false);

  useEffect(() => {
    initClient().then(() => setInitialized(true));
  }, []);

  if (!initialized) return <List isLoading />;

  return (
    <QueryProvider>
      <QueueView />
    </QueryProvider>
  );
}

function QueueView() {
  const { data, isLoading, refetch } = useQuery({
    ...getQueueOptions(),
    refetchInterval: 5000,
  });

  const clearMutation = useMutation({
    ...deleteQueueMutation(),
    onSuccess: () => {
      showToast(Toast.Style.Success, "Queue cleared");
      refetch();
    },
    onError: (err) =>
      showToast(Toast.Style.Failure, "Failed to clear queue", String(err)),
  });

  const removeMutation = useMutation({
    ...deleteQueueByPositionMutation(),
    onSuccess: () => {
      showToast(Toast.Style.Success, "Track removed");
      refetch();
    },
    onError: (err) =>
      showToast(Toast.Style.Failure, "Failed to remove track", String(err)),
  });

  return (
    <List isLoading={isLoading}>
      {data?.items?.map((item) => (
        <List.Item
          key={item.position}
          title={`${item.position! + 1}. ${item.track?.title ?? "Unknown"}`}
          subtitle={item.track?.artist}
          icon={
            item.position === data.current_position ? Icon.Play : Icon.Music
          }
          accessories={[{ text: item.track?.duration }]}
          actions={
            <ActionPanel>
              <Action
                title="Remove from Queue"
                icon={Icon.Trash}
                onAction={() =>
                  removeMutation.mutate({ path: { position: item.position! } })
                }
              />
              <Action
                title="Clear Entire Queue"
                icon={Icon.Trash}
                onAction={() => clearMutation.mutate({} as any)}
              />
            </ActionPanel>
          }
        />
      ))}
    </List>
  );
}
