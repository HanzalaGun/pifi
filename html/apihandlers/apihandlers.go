
package apihandlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	// "os"
	"os/exec"
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

// func ModifyNetworkHandler(nm networkmanager.NetworkManager) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		r.ParseForm()
// 		err := nm.ModifyNetworkConnection(r.Form.Get("ssid"), r.Form.Get("password"), false)
// 		if err != nil {
// 			jsonResponse(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
// 			return
// 		}
// 		// cmd := exec.Command("/home/optistok/.nvm/versions/node/v20.9.0/bin/node","/home/optistok/.nvm/versions/node/v20.9.0/bin/pm2", "restart", "optistokscrapping")
// 		// if err := cmd.Run(); err != nil {
// 		// 	jsonResponse(w, map[string]string{"error": "Failed to restart PM2: " + err.Error()}, http.StatusInternalServerError)
// 		// 	return
// 		// }
// 		// cmd := exec.Command("bash", "-c", "source /home/optistok/.nvm/nvm.sh && nvm use 20 && pm2 restart optistokscrapping")
// 		cmd := exec.Command("bash", "-c", "source /home/optistok/.nvm/nvm.sh && nvm use 20 && export PM2_HOME=/home/optistok/.pm2 && pm2 restart optistokscrapping")
// 		cmd.Stdout = os.Stdout
// 		cmd.Stderr = os.Stderr		
// 		output, err := cmd.CombinedOutput()
// 		if err != nil {
// 			fmt.Println("PM2 Hatası:", err)
// 			fmt.Println("PM2 Çıktısı:", string(output)) // Çıktıyı yazdırarak kullanılmış hale getiriyoruz
// 			jsonResponse(w, map[string]string{"error": "Failed to restart PM2: " + err.Error()}, http.StatusInternalServerError)
// 			return
// 		}
// 		fmt.Println("PM2 Çıktısı:", string(output))

// 		jsonResponse(w, map[string]string{"message": "Network modified successfully"}, http.StatusOK)
// 	}
// }
func ModifyNetworkHandler(nm networkmanager.NetworkManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// r.ParseForm()
		// err := nm.ModifyNetworkConnection(r.Form.Get("ssid"), r.Form.Get("password"), true)
		// if err != nil {
		// 	jsonResponse(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		// 	return
		// }

		// Komutun çalıştırılması
		
		cmd := exec.Command("bash", "-c", "source /home/optistok/.nvm/nvm.sh && nvm use 20 && export PM2_HOME=/home/optistok/.pm2 && pm2 restart optistokscrapping")
		log.Println("Komut Çalıştırılıyor:", cmd.String())
		// cmd.Stdout = os.Stdout  // Standart çıktıyı yazdır
		// cmd.Stderr = os.Stderr  // Hata çıktısını yazdır

		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Println("PM2 Hatası:", err)
			log.Println("PM2 Çıktısı:", string(output)) // Çıktıyı yazdır
			jsonResponse(w, map[string]string{"error": "Failed to restart PM2: " + err.Error()}, http.StatusInternalServerError)
			return
		}
		log.Println("PM2 Çıktısı:", string(output))

		// Başarı durumu
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
func RemoveAllNetworkConnectionHandler(nm networkmanager.NetworkManager) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        err := nm.SetupAPConnection()
        if err != nil {
            log.Fatalf("Error setting up AP connection: %v", err)
            jsonResponse(w, map[string]string{"error": "Failed to remove all networks"}, http.StatusInternalServerError)
            return
        }
        jsonResponse(w, map[string]string{"message": "All networks removed successfully"}, http.StatusOK)
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
