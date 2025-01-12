package networkmanager

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	ModeDual   = "dual"
	ModeClient = "client"
	ModeAP     = "ap"
)

type NetworkStatus struct {
	State        string
	Connectivity string
	WifiHW       string
	Wifi         string
	WifiSSID     string
	APSSID       string
	SignalStr    int32
	Mode         string
	IPs          NetworkIPs
}

type NetworkManager interface {
	GetNetworkStatus() (NetworkStatus, error)
	SetWifiMode(mode string) error
	SetupAPConnection() error

	// FindAvailableNetworks() ([]string, error)
	// ListWifiConnections() ([]string, error)
}

type networkManager struct {
	status NetworkStatus
}

func New() NetworkManager {
	nm := &networkManager{
		status: NetworkStatus{},
	}
	nm.GetNetworkStatus()
	return nm
}

func (nm *networkManager) GetNetworkStatus() (NetworkStatus, error) {
	cmd := exec.Command("nmcli", "g")
	output, err := cmd.Output()
	networkStatus := NetworkStatus{
		State:        "unknown",
		Connectivity: "unknown",
		WifiHW:       "unknown",
		Wifi:         "unknown",
		SignalStr:    0,
		Mode:         "unknown",
		APSSID:       "PiFi-AP",
		IPs:          NetworkIPs{},
	}
	if err != nil {
		return networkStatus, err
	}
	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return networkStatus, fmt.Errorf("unexpected nmcli output format")
	}
	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		return networkStatus, fmt.Errorf("invalid nmcli output fields")
	}

	networkStatus = NetworkStatus{
		State:        fields[0],
		Connectivity: fields[1],
		WifiHW:       fields[2],
		Wifi:         fields[3],
		WifiSSID:     nm.getWifiSSID(),
		SignalStr:    nm.getWifiSignal(),
		Mode:         nm.getWifiMode(),
		IPs:          nm.getNetworkIps(),
	}
	nm.status = networkStatus
	return networkStatus, nil
}

func (nm *networkManager) SetWifiMode(mode string) error {
	// Safety checks
	if mode == ModeAP {
		return fmt.Errorf("switching to AP-only mode is disabled for safety")
	}

	// Get current active connections
	cmd := exec.Command("nmcli", "-t", "-f", "NAME,TYPE,DEVICE", "con", "show", "--active")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get active connections: %v", err)
	}

	// Parse active connections
	hasAP := strings.Contains(string(output), "PiFi-AP")
	hasClient := strings.Contains(string(output), "wifi") || strings.Contains(string(output), "802-11-wireless")

	switch mode {
	case ModeDual:
		if !hasClient {
			return fmt.Errorf("must have active client connection for dual mode")
		}
		if !hasAP {
			err = nm.verifyAPConnection()
			if err != nil {
				return err
			}
			cmd = exec.Command("nmcli", "con", "up", "PiFi-AP")
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to enable AP mode: %v", err)
			}
		}
	case ModeClient:
		if hasAP {
			cmd = exec.Command("nmcli", "con", "down", "PiFi-AP")
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to disable AP mode: %v", err)
			}
		}
		if !hasClient {
			return fmt.Errorf("no active client connection")
		}
	default:
		return fmt.Errorf("unsupported mode: %s", mode)
	}

	// Verify mode change
	time.Sleep(time.Second)
	newMode := nm.getWifiMode()
	if newMode != mode {
		return fmt.Errorf("mode change verification failed")
	}

	return nil
}

func (nm *networkManager) verifyAPConnection() error {
	cmd := exec.Command("nmcli", "connection", "show", "PiFi-AP")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("AP connection not configured. Run: sudo nmcli connection add type wifi ifname wlan0 con-name PiFi-AP autoconnect no ssid PiFi mode ap 802-11-wireless.band bg")
	}
	return nil
}

func (nm *networkManager) SetupAPConnection() error {
	// Create virtual interface, if it doesn't exist
	if !checkInterfaceExists("wlan0_ap") {
		cmd := exec.Command("iw", "dev", "wlan0", "interface", "add", "wlan0_ap", "type", "__ap")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create virtual interface: %v", err)
		}
	}

	// Check if AP connection already exists
	cmd := exec.Command("nmcli", "connection", "show", "PiFi-AP")
	if err := cmd.Run(); err == nil {
		return nil
	}

	// Create AP connection with required settings
	cmd = exec.Command("nmcli", "connection", "add",
		"type", "wifi",
		"ifname", "wlan0_ap",
		"con-name", "PiFi-AP",
		"autoconnect", "no",
		"ssid", "PiFi",
		"mode", "ap",
		"ipv4.method", "shared",
		"ipv6.method", "disabled",
		"802-11-wireless.band", "bg",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create AP connection: %v\nOutput: %s", err, output)
	}

	cmd = exec.Command("nmcli", "connection", "show", "PiFi-AP")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("AP connection verification failed: %v", err)
	}
	return nil
}

func checkInterfaceExists(name string) bool {
	cmd := exec.Command("iw", "dev")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), name)
}

func (nm *networkManager) getWifiSignal() int32 {
	cmd := exec.Command("nmcli", "-f", "IN-USE,SIGNAL", "dev", "wifi", "list")
	output, err := cmd.Output()
	if err != nil {
		return -1
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "*") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				signal, err := strconv.ParseInt(fields[1], 10, 32)
				if err != nil {
					return -1
				}
				return int32(signal)
			}
		}
	}
	return -1
}

func (nm *networkManager) getWifiMode() string {
	cmd := exec.Command("nmcli", "-t", "-f", "NAME,TYPE,DEVICE", "con", "show", "--active")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	hasAP := strings.Contains(string(output), "PiFi-AP")
	hasClient := strings.Contains(string(output), "wifi") || strings.Contains(string(output), "802-11-wireless")

	if hasAP && hasClient {
		return ModeDual
	} else if hasClient {
		return ModeClient
	} else if hasAP {
		return ModeAP
	}
	return "inactive"
}

func (nm *networkManager) getWifiIP() string {
	cmd := exec.Command("nmcli", "-g", "IP4.ADDRESS", "dev", "show", "wlan0")
	output, err := cmd.Output()
	if err != nil {
		return "not connected"
	}

	ip := strings.TrimSpace(string(output))
	if ip == "" {
		return "not connected"
	}

	// Remove CIDR notation if present
	if strings.Contains(ip, "/") {
		ip = strings.Split(ip, "/")[0]
	}

	return ip
}

func (nm *networkManager) getEthernetIP() string {
	cmd := exec.Command("nmcli", "-g", "IP4.ADDRESS", "dev", "show", "eth0")
	output, err := cmd.Output()
	if err != nil {
		return "not connected"
	}

	ip := strings.TrimSpace(string(output))
	if ip == "" {
		return "not connected"
	}

	// Remove CIDR notation if present
	if strings.Contains(ip, "/") {
		ip = strings.Split(ip, "/")[0]
	}

	return ip
}

func (nm *networkManager) getAPIP() string {
	cmd := exec.Command("nmcli", "-g", "IP4.ADDRESS", "dev", "show", "wlan0_ap")
	output, err := cmd.Output()
	if err != nil {
		return "not connected"
	}

	ip := strings.TrimSpace(string(output))
	if ip == "" {
		return "not connected"
	}

	// Remove CIDR notation if present
	if strings.Contains(ip, "/") {
		ip = strings.Split(ip, "/")[0]
	}

	return ip
}

type NetworkIPs struct {
	WifiIP     string
	WifiState  string
	EthernetIP string
	EthState   string
	APIP       string
	APState    string
}

func (nm *networkManager) getNetworkIps() NetworkIPs {
	status := NetworkIPs{
		WifiState: "offline",
		EthState:  "offline",
		APState:   "offline",
	}

	// Check WiFi
	if output, err := exec.Command("nmcli", "-g", "IP4.ADDRESS", "dev", "show", "wlan0").Output(); err == nil {
		if ip := strings.TrimSpace(string(output)); ip != "" {
			status.WifiIP = strings.Split(ip, "/")[0]
			status.WifiState = "online"
		}
	}

	// Check Ethernet
	if output, err := exec.Command("nmcli", "-g", "IP4.ADDRESS", "dev", "show", "eth0").Output(); err == nil {
		if ip := strings.TrimSpace(string(output)); ip != "" {
			status.EthernetIP = strings.Split(ip, "/")[0]
			status.EthState = "online"
		}
	}

	// Check AP
	if output, err := exec.Command("nmcli", "-g", "IP4.ADDRESS", "dev", "show", "wlan0_ap").Output(); err == nil {
		if ip := strings.TrimSpace(string(output)); ip != "" {
			status.APIP = strings.Split(ip, "/")[0]
			status.APState = "online"
		}
	}

	return status
}

func (nm *networkManager) getWifiSSID() string {
	cmd := exec.Command("nmcli", "-t", "-f", "active,ssid", "dev", "wifi")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) == 2 && fields[0] == "yes" {
			return fields[1]
		}
	}
	return ""
}
