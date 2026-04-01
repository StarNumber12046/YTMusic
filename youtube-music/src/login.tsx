import {
  Action,
  ActionPanel,
  Detail,
  Form,
  Toast,
  showToast,
} from "@raycast/api";
import { useQuery, useMutation } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { initClient } from "./lib/client";
import { saveToken, clearToken, AuthError, needsCookieRefresh, isAuthError } from "./lib/auth";
import { QueryProvider } from "./components/QueryProvider";
import { queryClient } from "./lib/query";
import {
  getAuthStatusOptions,
  postAuthLoginMutation,
  deleteAuthLogoutMutation,
} from "./generated/@tanstack/react-query.gen";
import { client } from "./generated/client.gen";

export default function Command() {
  const [initialized, setInitialized] = useState(false);

  useEffect(() => {
    initClient().then(() => setInitialized(true));
  }, []);

  if (!initialized) return <Detail isLoading />;

  return (
    <QueryProvider>
      <LoginView />
    </QueryProvider>
  );
}

function LoginView() {
  const [showRefreshForm, setShowRefreshForm] = useState(false);
  const { data: status, isLoading, refetch } = useQuery(getAuthStatusOptions());

  const loginMutation = useMutation({
    ...postAuthLoginMutation(),
    onSuccess: async (data) => {
      if (data?.token) {
        await saveToken(data.token, data.expires_at!);
        showToast(Toast.Style.Success, "Logged in successfully");
        await initClient();
        refetch();
      }
    },
    onError: (err) =>
      showToast(Toast.Style.Failure, "Login failed", String(err)),
  });

  const refreshMutation = useMutation({
    mutationFn: async (cookies?: string) => {
      const response = await client.post({
        url: "/auth/refresh",
        body: cookies ? { cookies } : {},
      });
      return response.data as { token: string; expires_at: string; used_preseeded: boolean };
    },
    onSuccess: async (data) => {
      if (data?.token) {
        await saveToken(data.token, data.expires_at, data.used_preseeded);
        showToast(
          Toast.Style.Success,
          data.used_preseeded ? "Session refreshed from config" : "Session refreshed"
        );
        await initClient();
        setShowRefreshForm(false);
        queryClient.clear();
        refetch();
      }
    },
    onError: async (err) => {
      if (isAuthError(err)) {
        const authErr = err as AuthError;
        if (needsCookieRefresh(authErr)) {
          setShowRefreshForm(true);
          showToast(Toast.Style.Failure, "Cookies expired", "Please paste new cookies");
          return;
        }
      }
      showToast(Toast.Style.Failure, "Refresh failed", String(err));
    },
  });

  const logoutMutation = useMutation({
    ...deleteAuthLogoutMutation(),
    onSuccess: async () => {
      await clearToken();
      queryClient.clear();
      showToast(Toast.Style.Success, "Logged out");
      await initClient();
      refetch();
    },
    onError: (err) =>
      showToast(Toast.Style.Failure, "Logout failed", String(err)),
  });

  if (isLoading) return <Detail isLoading />;

  if (showRefreshForm) {
    return (
      <Form
        actions={
          <ActionPanel>
            <Action.SubmitForm
              title="Refresh Session"
              onSubmit={(values: { cookies: string }) =>
                refreshMutation.mutate(values.cookies)
              }
            />
            <Action
              title="Cancel"
              onAction={() => setShowRefreshForm(false)}
            />
          </ActionPanel>
        }
      >
        <Form.Description
          title="Session Expired"
          text="Your YouTube Music cookies have expired. Please paste fresh cookies from your browser's DevTools (Network tab → any request to music.youtube.com → Cookie header)."
        />
        <Form.TextArea
          id="cookies"
          title="Browser Cookies"
          placeholder="Paste your 'Cookie' header string..."
        />
      </Form>
    );
  }

  if (status?.authenticated) {
    return (
      <Detail
        markdown={`# You are authenticated\n\n**Account**: ${status.account_name || "Unknown"}\n\n**Session Expires**: ${new Date(status.expires_at!).toLocaleString()}`}
        actions={
          <ActionPanel>
            <Action title="Logout" onAction={() => logoutMutation.mutate({} as any)} />
            <Action
              title="Refresh Session"
              onAction={() => refreshMutation.mutate(undefined)}
              shortcut={{ modifiers: ["cmd"], key: "r" }}
            />
            <Action
              title="Paste New Cookies"
              onAction={() => setShowRefreshForm(true)}
              shortcut={{ modifiers: ["cmd", "shift"], key: "r" }}
            />
          </ActionPanel>
        }
      />
    );
  }

  return (
    <Form
      actions={
        <ActionPanel>
          <Action.SubmitForm
            title="Login"
            onSubmit={(values: { cookies: string }) =>
              loginMutation.mutate({ body: { cookies: values.cookies } })
            }
          />
        </ActionPanel>
      }
    >
      <Form.TextArea
        id="cookies"
        title="Browser Cookies"
        placeholder="Paste your 'Cookie' header string from the browser network tab..."
      />
    </Form>
  );
}
