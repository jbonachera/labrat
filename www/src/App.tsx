import "./App.css";
import React, { useEffect } from "react";
import crc32 from "crc-32";
import colors from "./colors.json";

interface appState {
  instance_id?: string;
  timestamp?: string;
  started_at?: string;
  version?: string;
}

function App() {
  const [state, setState] = React.useState<appState>({});

  const load = async () => {
    const resp = await fetch("/health");
    const data = (await resp.json()) as appState;
    setState(data);
  };

  const fail = async () => {
    const resp = await fetch("/closeListener");
    const data = (await resp.json()) as appState;
    setState(data);
  };

  const formatVersion = (version?: string): string => {
    if (version === undefined) {
      return "unknown";
    }
    return version.slice(0, 7);
  };
  const formatID = (id?: string): string => {
    if (id === undefined) {
      return "Unknown";
    }
    const num = Math.abs(crc32.str(id));
    return colors[num % colors.length].name;
  };
  const dateToDuration = (timestamp?: string) => {
    if (timestamp === undefined) {
      return "never";
    }
    return new Date(timestamp).toLocaleTimeString();
  };

  useEffect(() => {
    load();
  }, []);

  return (
    <div className="App">
      <header className="App-header">
        <p>
          We are running on {formatID(state.instance_id)} (version{" "}
          {formatVersion(state.version)}).
          <br />
          Last refresh: {dateToDuration(state.timestamp)}
        </p>
        <div>
          <button onClick={(e) => load()}>Refresh</button>
          <button onClick={(e) => fail()}>Crash</button>
        </div>
      </header>
    </div>
  );
}

export default App;
