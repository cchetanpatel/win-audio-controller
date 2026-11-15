import React, { createContext, useContext, useState, useEffect } from "react";

const AppContext = createContext();

export function AppProvider({ children }) {
  const [devices, setDevices] = useState([]);
  const [currentDevice, setCurrentDevice] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    async function fetchDevices() {
      try {
        //const response = await fetch("http://10.1.1.11:8080/audio");
        const response = await fetch("/api/audio");

        if (!response.ok) {
          throw new Error("Failed to fetch devices");
        }

        const data = await response.json();

        setDevices(data["devices"]);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    }

    async function fetchCurrentDevice() {
      try {
        //const response = await fetch("http://10.1.1.11:8080/getaudio");
        const response = await fetch("/api/getaudio");

        if (!response.ok) {
          throw new Error("Failed to fetch current device");
        }

        const data = await response.json();

        setCurrentDevice(data);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    }

    fetchDevices();
    fetchCurrentDevice();
  }, []);

  return (
    <AppContext.Provider
      value={{
        devices,
        setDevices,
        currentDevice,
        setCurrentDevice,
        loading,
        error,
      }}
    >
      {children}
    </AppContext.Provider>
  );
}

export function useApp() {
  return useContext(AppContext);
}
