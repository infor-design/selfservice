import { parseOrThrowRequest } from "./utils";
import { SERVER_URL } from "../constants";

export const fetchLogs = async (id: number) => {
  const url = `${SERVER_URL}/jobs/${id}/logs`;
  return (await parseOrThrowRequest(url)) as Promise<any>;
};
