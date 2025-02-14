// package handlers

// import (
// 	"fmt"
// 	"html/template"
// 	"net/http"
// 	"time"

// 	"github.com/HanzalaGun/pifi/html"
// 	"github.com/HanzalaGun/pifi/networkmanager"
// )

// type StatusResponse struct {
// 	Status      string    `json:"status"`
// 	Timestamp   time.Time `json:"timestamp"`
// 	Version     string    `json:"version"`
// 	NetworkInfo networkmanager.NetworkStatus
// }

// type NetworkResponse struct {
// 	AvailableNetworks  []string                        `json:"availableNetworks"`
// 	ConfiguredNetworks []networkmanager.ConnectionInfo `json:"configuredNetworks"`
// 	Timestamp          time.Time                       `json:"timestamp"`
// }

// func SetMode(nm networkmanager.NetworkManager) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		r.ParseForm()
// 		err := nm.SetWifiMode(r.Form.Get("mode"))
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 	}
// }

// func PiFiHandler(nm networkmanager.NetworkManager) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		tmpl, err := template.ParseFS(html.Templates, "templates/index.gohtml")
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 		err = tmpl.Execute(w, nil)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 	}
// }

// func StatusHandler(nm networkmanager.NetworkManager) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		status := StatusResponse{
// 			Status:    "operational",
// 			Timestamp: time.Now(),
// 			Version:   "1.0.0",
// 		}
// 		netStatus, err := nm.GetNetworkStatus()
// 		if err != nil {
// 			status.Status = fmt.Sprintf("error: %v", err)
// 		}
// 		status.NetworkInfo = netStatus

// 		tmpl, err := template.ParseFS(html.Templates, "templates/status.gohtml")
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 		err = tmpl.Execute(w, status)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 	}
// }

// func NetworksHandler(nm networkmanager.NetworkManager) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		availableNetworks, err := nm.FindAvailableNetworks()
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 		configuredNetworks, err := nm.GetConfiguredConnections()
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}

// 		tmpl, err := template.ParseFS(html.Templates, "templates/network.gohtml")
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 		NetworkResponse := NetworkResponse{
// 			AvailableNetworks:  availableNetworks,
// 			ConfiguredNetworks: configuredNetworks,
// 			Timestamp:          time.Now(),
// 		}
// 		err = tmpl.Execute(w, NetworkResponse)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 	}
// }

// func ModifyNetworkHandler(nm networkmanager.NetworkManager) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		r.ParseForm()
// 		err := nm.ModifyNetworkConnection(r.Form.Get("ssid"), r.Form.Get("password"), false)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 	}
// }

// func RemoveNetworkConnectionHandler(nm networkmanager.NetworkManager) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		r.ParseForm()
// 		err := nm.RemoveNetworkConnection(r.Form.Get("network"))
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 	}
// }

// func AutoConnectNetworkHandler(nm networkmanager.NetworkManager) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		r.ParseForm()
// 		err := nm.SetAutoConnectConnection(r.Form.Get("network"), true)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 	}
// }

// func ConnectNetworkHandler(nm networkmanager.NetworkManager) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		r.ParseForm()
// 		err := nm.ConnectNetwork(r.Form.Get("network"))
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 	}
// }
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/HanzalaGun/pifi/networkmanager"
)

type StatusResponse struct {
	Status      string                        `json:"status"`
	Timestamp   time.Time                     `json:"timestamp"`
	Version     string                        `json:"version"`
	NetworkInfo networkmanager.NetworkStatus `json:"networkInfo"`
}

type NetworkResponse struct {
	AvailableNetworks  []string                        `json:"availableNetworks"`
	ConfiguredNetworks []networkmanager.ConnectionInfo `json:"configuredNetworks"`
	Timestamp          time.Time                       `json:"timestamp"`
}

func jsonResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func SetMode(nm networkmanager.NetworkManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		err := nm.SetWifiMode(r.Form.Get("mode"))
		if err != nil {
			jsonResponse(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
			return
		}
		jsonResponse(w, map[string]string{"message": "Mode set successfully"}, http.StatusOK)
	}
}

func StatusHandler(nm networkmanager.NetworkManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := StatusResponse{
			Status:    "operational",
			Timestamp: time.Now(),
			Version:   "1.0.0",
		}
		netStatus, err := nm.GetNetworkStatus()
		if err != nil {
			status.Status = fmt.Sprintf("error: %v", err)
		}
		status.NetworkInfo = netStatus
		jsonResponse(w, status, http.StatusOK)
	}
}

func NetworksHandler(nm networkmanager.NetworkManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		availableNetworks, err := nm.FindAvailableNetworks()
		if err != nil {
			jsonResponse(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
			return
		}
		configuredNetworks, err := nm.GetConfiguredConnections()
		if err != nil {
			jsonResponse(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
			return
		}
		response := NetworkResponse{
			AvailableNetworks:  availableNetworks,
			ConfiguredNetworks: configuredNetworks,
			Timestamp:          time.Now(),
		}
		jsonResponse(w, response, http.StatusOK)
	}
}

func ModifyNetworkHandler(nm networkmanager.NetworkManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		err := nm.ModifyNetworkConnection(r.Form.Get("ssid"), r.Form.Get("password"), false)
		if err != nil {
			jsonResponse(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
			return
		}
		jsonResponse(w, map[string]string{"message": "Network modified successfully"}, http.StatusOK)
	}
}

func RemoveNetworkConnectionHandler(nm networkmanager.NetworkManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		err := nm.RemoveNetworkConnection(r.Form.Get("network"))
		if err != nil {
			jsonResponse(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
			return
		}
		jsonResponse(w, map[string]string{"message": "Network removed successfully"}, http.StatusOK)
	}
}

func AutoConnectNetworkHandler(nm networkmanager.NetworkManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		err := nm.SetAutoConnectConnection(r.Form.Get("network"), true)
		if err != nil {
			jsonResponse(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
			return
		}
		jsonResponse(w, map[string]string{"message": "Auto-connect enabled"}, http.StatusOK)
	}
}

func ConnectNetworkHandler(nm networkmanager.NetworkManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		err := nm.ConnectNetwork(r.Form.Get("network"))
		if err != nil {
			jsonResponse(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
			return
		}
		jsonResponse(w, map[string]string{"message": "Connected successfully"}, http.StatusOK)
	}
}
