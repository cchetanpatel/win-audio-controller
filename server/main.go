// Install-Module -Name AudioDeviceCmdlets -Scope CurrentUser

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"fyne.io/systray"
	_ "embed"
	"strings"
	//"sync"
)
//go:embed icon.ico
var iconBytes []byte

// Item represents our backend data model
type Item struct {
	Device    string `json:"Device"`
	InstanceId    string `json:"InstanceId"`
}


type AudioDevice struct {
    FriendlyName string
    InstanceId   string
}

// Install-Module -Name AudioDeviceCmdlets -Scope CurrentUser

// healthHandler handles simple monitoring checks
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// audioHandler routes basic GET and POST methods
func audioHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
        query := r.URL.Query()

        device := query.Get("Device")
        instanceId := query.Get("InstanceId")

        // If parameters are supplied, change audio device
        if device != "" && instanceId != "" {
            changeAudioDevice(w, device, instanceId)
            return
        }

        // Otherwise, return available devices
        getItems(w, r)
	// case http.MethodPost:
	// 	createItem(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
	}
}

// audioHandler routes basic GET
func getAudioHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
    
		cmd := exec.Command(
			"powershell",
			"-NoProfile",
			"-NonInteractive",
			"-Command",
			"Get-AudioDevice -Playback | Select-Object @{Name='FriendlyName';Expression={$_.Name}},@{Name='InstanceID';Expression={$_.ID}} | ConvertTo-Json -Compress",
		)

		out, err := cmd.Output()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var device AudioDevice

		// ConvertFrom-Json can return a single object instead of an array
		if err := json.Unmarshal(out, &device); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(device)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
	}
}


// getItems retrieves all unique audio devices
func getItems(w http.ResponseWriter, r *http.Request) {

    cmd := exec.Command(
        "powershell",
        "-NoProfile",
        "-NonInteractive",
        "-Command",
        "Get-PnpDevice -Class AudioEndpoint  -PresentOnly -Status OK | Select-Object FriendlyName, InstanceId | ConvertTo-Json -Compress",
    )

    out, err := cmd.Output()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    var devices []AudioDevice

    // ConvertFrom-Json can return a single object instead of an array
    if err := json.Unmarshal(out, &devices); err != nil {
        var device AudioDevice
        if err := json.Unmarshal(out, &device); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        devices = []AudioDevice{device}
    }

    // Remove duplicate friendly names
    unique := make(map[string]AudioDevice)

    for _, device := range devices {
		if strings.Contains(device.InstanceId, "{0.0.0.") {
			device.InstanceId =	 strings.ToLower(string(device.InstanceId)[13:])
			unique[device.FriendlyName] = device

		}
        
    }

    result := make([]AudioDevice, 0, len(unique))
    for _, device := range unique {
        result = append(result, device)
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)

    json.NewEncoder(w).Encode(map[string][]AudioDevice{
        "devices": result,
    })
}

func changeAudioDevice(w http.ResponseWriter, device string, instanceId string) {
    cmd := exec.Command(
        "powershell",
        "-NoProfile",
        "-NonInteractive",
        "-Command",
        fmt.Sprintf(`Set-AudioDevice -ID "%s"`, instanceId),
    )

    output, err := cmd.CombinedOutput()
    if err != nil {
        http.Error(w, string(output), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "success",
        "device": device,
    })
}

// createItem adds a record to the memory store (Thread-Safe write)
func createItem(w http.ResponseWriter, r *http.Request) {
	var newItem Item

	// Decode incoming JSON request payload
	if err := json.NewDecoder(r.Body).Decode(&newItem); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if newItem.Device == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Item Device is required"})
		return
	}


	cmd := exec.Command(
        "powershell",
        "-NoProfile",
        "-NonInteractive",
        "-Command",
        fmt.Sprintf(`Set-AudioDevice -ID "%s"`, string(newItem.InstanceId)),
    )

    _, err := cmd.Output()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }



	// Write lock to prevent concurrent write panic errors
	// db.Lock()
	// db.items[newItem.ID] = newItem
	// db.Unlock()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newItem)
}

func enableCORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

        w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        // Handle browser preflight requests
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusNoContent)
            return
        }

        next.ServeHTTP(w, r)
    })
}

func onReady() {
	// Set icon (provide raw bytes of an .ico for Windows or .ico/.png for macOS/Linux)
	// Replace iconBytes with your own embedded or read byte slice

	systray.SetIcon(iconBytes)
	systray.SetTitle("My Go App")
	systray.SetTooltip("My Go Application Running in Background")

	// Add menu items
	mShow := systray.AddMenuItem("Show App", "Bring window to front")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit the application")


	// Register routes
	mux := http.NewServeMux()

	// Health Check
	// Set Device
	// Get Device List Audio
	// Get Current Device 

	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/getaudio", getAudioHandler)	
	mux.HandleFunc("/audio", audioHandler)

	handler := enableCORS(mux)
	port := ":8080"

	// Start server
	
	
	fmt.Printf("🚀 Server starting on http://0.0.0.0%s\n", port)
	
	go func() {
		if err := 	http.ListenAndServe(port, handler); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Handle menu item events in goroutines
	go func() {
		for {
			select {
			case <-mShow.ClickedCh:
				fmt.Println("Show App clicked")
				// Logic to show/restore your main window here
			case <-mQuit.ClickedCh:
				fmt.Println("Quit clicked")
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	// Clean up resources when the tray is closed
}


func main() {
	// Seed data

	systray.Run(onReady, onExit)
	
}