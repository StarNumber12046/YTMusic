import { postPlayerShuffle } from "./generated/sdk.gen";
import { initClient } from "./lib/client";

export default async function Command() {
  await initClient();
  const { error } = await postPlayerShuffle();
  if (error) {
    console.error("Failed to toggle shuffle:", error);
  }
}
