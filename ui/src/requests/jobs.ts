import { deleteRequest, parseOrThrowRequest } from "./utils";
import { SERVER_URL } from "../constants";

export const fetchJob = async (id: number) => {
  const url = `${SERVER_URL}/jobs/${id}`;
  return (await parseOrThrowRequest(url)) as Promise<any>;
};

export const deleteJob = async (id: string) => {
  const url = `${SERVER_URL}/jobs/${id}`;
  return (await deleteRequest(url, {})) as Promise<any>;
};
