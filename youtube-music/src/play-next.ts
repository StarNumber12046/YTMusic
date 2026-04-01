import { showHUD, showToast, Toast } from "@raycast/api";
import { initClient } from "./lib/client";
import { postQueuePlayNext } from "./generated/sdk.gen";

export default async function Command({ videoId }: { videoId?: string }) {
  await initClient();

  if (!videoId) {
    await showToast(Toast.Style.Failure, "Error", "No video ID provided");
    return;
  }

  try {
    const { data, error } = await postQueuePlayNext({
      body: { video_id: videoId },
    });

    if (error) {
      throw new Error(
        (error as unknown as Record<string, string>).error ||
          "Failed to play next",
      );
    }

    await showHUD(`Play Next: ${data?.items?.[(data?.current_position ?? 0) + 1]?.track?.title || "track added"}`);
  } catch (e) {
    await showToast(Toast.Style.Failure, "Error", String(e));
  }
}
