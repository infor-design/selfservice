import { parseOrThrowRequest } from "./utils";
import { SERVER_URL } from "../constants";

export const fetchSettings = async () => {
  const url = `${SERVER_URL}/settings`;
  return (await parseOrThrowRequest(url)) as Promise<Record<string, string>>;
};
