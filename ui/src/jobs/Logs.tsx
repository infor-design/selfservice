import { useEffect, useState } from "react";
import Highlight from "react-highlight";
import { fetchLogs } from "../requests/logs";
import { getErrorMessage } from "../requests/utils";

import "highlight.js/styles/a11y-dark.css";

const waitForOpenConnection = (socket: WebSocket) => {
  return new Promise((resolve, reject) => {
    const maxNumberOfAttempts = 10;
    const intervalTime = 200;
    let currentAttempt = 0;
    const interval = setInterval(() => {
      if (currentAttempt > maxNumberOfAttempts - 1) {
        clearInterval(interval);
        reject(new Error("Maximum number of attempts exceeded"));
      } else if (socket.readyState === socket.OPEN) {
        clearInterval(interval);
        resolve("");
      }
      currentAttempt++;
    }, intervalTime);
  });
};

const sendMessage = async (socket: WebSocket, msg: any) => {
  if (socket.readyState !== socket.OPEN) {
    try {
      await waitForOpenConnection(socket);
      socket.send(msg);
    } catch (err) {
      console.error(err);
    }
  } else {
    socket.send(msg);
  }
};

const Logs = ({ job, ws }: { job: any; ws: WebSocket | null }) => {
  const [logs, setLogs] = useState<string[]>([]);
  const [liveLogs, setLiveLogs] = useState<string[]>([]);
  const [fetchErrors, setFetchErrors] = useState<string>();

  useEffect(() => {
    if (job && job.phase === "Running") {
      const apiCall = {
        event: "logs:subscribe",
        data: {
          jobId: job.id,
        },
        resource_namespace: job.meta.namespace,
        resource_name: job.name,
      };

      if (ws && ws.readyState === 1) {
        sendMessage(ws, JSON.stringify(apiCall));
      }
    }
  }, [ws, job]);

  useEffect(() => {
    if (ws && ws.OPEN) {
      ws.onmessage = function (event) {
        const json = JSON.parse(event.data);
        setLiveLogs((prev) => [...prev, ...[json.data]]);
      };
    }
  }, [ws]);

  useEffect(() => {
    fetchLogs(job.id)
      .then((data) => {
        let fetchedLogs: string[] = [];
        data.forEach((obj: any) => {
          fetchedLogs = [...fetchedLogs, ...obj.file_data.logs];
        });
        setLogs(fetchedLogs);
      })
      .catch((e) => setFetchErrors(getErrorMessage(e)));
  }, []);

  return (
    <>
      {fetchErrors && <>{fetchErrors}</>}

      <Highlight className="plaintext">
        {logs && logs.length > 0 && logs?.map((log, index) => <div key={index}>{log}</div>)}

        {liveLogs &&
          liveLogs.length > 0 &&
          liveLogs?.map((log, index) => <div key={index}>{log}</div>)}
      </Highlight>
    </>
  );
};

export default Logs;
