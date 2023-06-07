import { deleteRequest, parseOrThrowRequest, post, put } from "./utils";
import { Application, RunStatus, FormData, ApplicationFull } from "../types";
import { SERVER_URL } from "../constants";

export const createApplication = async (values: Partial<any>) => {
  const url = `${SERVER_URL}/applications`;
  return (await post(url, values)) as Promise<any>;
};

export const updateApplication = async (id: string, values: Partial<any>) => {
  const url = `${SERVER_URL}/applications/${id}`;
  return (await put(url, values)) as Promise<any>;
};

export const deleteApplication = async (id: string) => {
  const url = `${SERVER_URL}/applications/${id}`;
  return (await deleteRequest(url, {})) as Promise<any>;
};

export const fetchApplications = async () => {
  const url = `${SERVER_URL}/applications`;
  return (await parseOrThrowRequest(url)) as Promise<Application[]>;
};

export const fetchApplication = async (id: number) => {
  const url = `${SERVER_URL}/applications/${id}`;
  return (await parseOrThrowRequest(url)) as Promise<ApplicationFull>;
};

export const fetchApplicationJobs = async (id: number) => {
  const url = `${SERVER_URL}/applications/${id}/jobs`;
  return (await parseOrThrowRequest(url)) as Promise<Application[]>;
};

export const startJob = async (id: number, data: FormData) => {
  const url = `${SERVER_URL}/applications/${id}/jobs`;
  return (await post(url, data)) as Promise<RunStatus>;
};
