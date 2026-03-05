import { postPlayerRepeat } from "./generated/sdk.gen";
import { initClient } from "./lib/client";

export default async function Command() {
  await initClient();
  const { error } = await postPlayerRepeat({ body: { repeat: "cycle" } });
  if (error) {
    console.error("Failed to cycle repeat:", error);
  }
}
