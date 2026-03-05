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
  getUserOptions,
  deleteAuthLogoutMutation,
} from "./generated/@tanstack/react-query.gen";
import { clearToken } from "./lib/auth";

export default function Command() {
  const [initialized, setInitialized] = useState(false);

  useEffect(() => {
    initClient().then(() => setInitialized(true));
  }, []);

  if (!initialized) return <Detail isLoading />;

  return (
    <QueryProvider>
      <UserInfoView />
    </QueryProvider>
  );
}

function UserInfoView() {
  const { data, isLoading, error } = useQuery(getUserOptions());

  const logoutMutation = useMutation({
    ...deleteAuthLogoutMutation(),
    onSuccess: async () => {
      await clearToken();
      showToast(Toast.Style.Success, "Logged out");
    },
    onError: (err) =>
      showToast(Toast.Style.Failure, "Logout failed", String(err)),
  });

  if (isLoading) return <Detail isLoading />;

  if (error) {
    return (
      <Detail
        markdown={`**Failed to load user info**\n\n\`\`\`\n${error}\n\`\`\`\n\nMake sure you are logged in.`}
        actions={
          <ActionPanel>
            <Action title="Logout" onAction={() => logoutMutation.mutate()} />
          </ActionPanel>
        }
      />
    );
  }

  return (
    <Detail
      markdown={`# User Info\n\n**Account**: ${data?.account_name || "Unknown"}\n\n**Channel ID**: ${data?.channel_id || "Unknown"}\n\n**Channel Title**: ${data?.channel_title || "Unknown"}`}
      actions={
        <ActionPanel>
          <Action
            title="Logout"
            icon={Icon.Logout}
            onAction={() => logoutMutation.mutate()}
          />
        </ActionPanel>
      }
    />
  );
}
