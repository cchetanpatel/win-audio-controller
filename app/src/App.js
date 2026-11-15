import logo from "./logo.svg";
import "./App.css";
import { useApp } from "./context/AppContext";

function App() {
  const { devices, currentDevice, setCurrentDevice, loading, error } = useApp();

  if (loading) return <p> Loading Devices... </p>;
  if (error) return <p>error</p>;

  const handleDeviceClick = async (device) => {
    const params = new URLSearchParams({
      Device: device.FriendlyName,
      InstanceId: device.InstanceId,
    });

    try {
      const response = await fetch(`/api/audio?${params.toString()}`);

      /* const response = await fetch(
        `http://10.1.1.11:8080/audio?${params.toString()}`,
      ); */

      if (!response.ok) {
        throw new Error(`HTTP error: ${response.status}`);
      }

      setCurrentDevice(device);
    } catch (error) {
      console.error("Failed to change audio device:", error);
    }
  };

  return (
    <div className="App">
      <header className="App-header">
        <h1>Audio Switcher</h1>
        <ul className="audio-selector">
          {devices.map((i) => (
            <li
              key={i.InstanceID}
              className={
                currentDevice.InstanceId === i.InstanceId
                  ? "selected-device"
                  : ""
              }
              onClick={() => handleDeviceClick(i)}
            >
              {i.FriendlyName}
            </li>
          ))}
        </ul>
      </header>
    </div>
  );
}

export default App;
