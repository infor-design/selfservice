import { deleteRequest, parseOrThrowRequest, post, put } from "./utils";
import { Repo, RepoCreate } from "../types";
import { SERVER_URL } from "../constants";

export const fetchRepos = async () => {
  const url = `${SERVER_URL}/repos`;
  return (await parseOrThrowRequest(url)) as Promise<Repo[]>;
};

export const createRepo = async (values: RepoCreate) => {
  const url = `${SERVER_URL}/repos`;
  return (await post(url, values)) as Promise<Repo>;
};

export const updateRepo = async (id: string, values: Partial<any>) => {
  const url = `${SERVER_URL}/repos/${id}`;
  return (await put(url, values)) as Promise<any>;
};

export const fetchRepo = async (id: string) => {
  const url = `${SERVER_URL}/repos/${id}`;
  return (await parseOrThrowRequest(url)) as Promise<Repo>;
};

export const syncRepo = async (id: string) => {
  const url = `${SERVER_URL}/repos/${id}/sync`;
  return (await post(url, {})) as Promise<Repo>;
};

export const deleteRepo = async (id: string) => {
  const url = `${SERVER_URL}/repos/${id}`;
  return (await deleteRequest(url, {})) as Promise<any>;
};
