package networkmanager

import (
	"fmt"
	"log"
	"math/rand"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
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

type NetworkIPs struct {
	WifiIP     string
	WifiState  string
	EthernetIP string
	EthState   string
	APIP       string
	APState    string
}

type ConnectionInfo struct {
	SSID     string
	Password string
}

type NetworkManager interface {
	SetupAPConnection() error
	ManageOfflineAP(connectionLossTimeout time.Duration) error

	// Network Status
	GetNetworkStatus() (NetworkStatus, error)
	SetWifiMode(mode string) error

	// Network Configuration
	FindAvailableNetworks() ([]string, error)
	GetConfiguredConnections() ([]ConnectionInfo, error)
	ModifyNetworkConnection(ssid, password string, autoConnect bool) error
	RemoveNetworkConnection(ssid string) error
	SetAutoConnectConnection(ssid string, autoConnect bool) error
	ConnectNetwork(ssid string) error
}

type networkManager struct {
	status NetworkStatus
}

func New() NetworkManager {
	nm := &networkManager{
		status: NetworkStatus{
			APSSID: "PiFi-AP-" + randSeq(4),
		},
	}
	nm.GetNetworkStatus()
	return nm
}

func randSeq(n int) string {
	var letters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (nm *networkManager) GetNetworkStatus() (NetworkStatus, error) {
	cmd := exec.Command("nmcli", "g")
	output, err := cmd.Output()
	if err != nil {
		return nm.status, err
	}
	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return nm.status, fmt.Errorf("unexpected nmcli output format")
	}
	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		return nm.status, fmt.Errorf("invalid nmcli output fields")
	}

	// Parse network status
	state := fields[0]
	connectivity := fields[1]
	wifiHW := fields[2]
	wifi := fields[3]
	if fields[1] == "(site" || fields[1] == "(local" && len(fields) >= 6 {
		state += " " + fields[1] + " " + fields[2]
		connectivity = fields[3]
		wifiHW = fields[4]
		wifi = fields[5]
	}

	setCase := cases.Title(language.English)
	networkStatus := NetworkStatus{
		APSSID:       nm.status.APSSID,
		State:        setCase.String(state),
		Connectivity: setCase.String(connectivity),
		WifiHW:       setCase.String(wifiHW),
		Wifi:         setCase.String(wifi),
		WifiSSID:     getWifiSSID(),
		SignalStr:    getWifiSignal(),
		Mode:         getWifiMode(nm.status.APSSID),
		IPs:          getNetworkIps(),
	}
	nm.status = networkStatus
	return networkStatus, nil
}

// Switches between client and AP modes
func (nm *networkManager) SetWifiMode(mode string) error {
	// Get current active connections
	cmd := exec.Command("nmcli", "-t", "-f", "NAME,TYPE,DEVICE", "con", "show", "--active")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get active connections: %v", err)
	}

	hasAP := strings.Contains(string(output), nm.status.APSSID)
	hasClient := strings.Contains(string(output), "wifi") || strings.Contains(string(output), "802-11-wireless")
	switch mode {
	case ModeAP:
		if !hasClient {
			return fmt.Errorf("must have active client connection for ap mode")
		}
		if !hasAP {
			err = verifyAPConnection(nm.status.APSSID)
			if err != nil {
				return err
			}
			cmd = exec.Command("nmcli", "con", "up", nm.status.APSSID)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("failed to create AP connection: %v\nOutput: %s", err, output)
			}
			time.Sleep(time.Second)
			newMode := getWifiMode(nm.status.APSSID)
			if newMode != "ap" {
				return fmt.Errorf("mode change verification failed")
			}
		}
	case ModeClient:
		if hasAP {
			cmd = exec.Command("nmcli", "con", "down", nm.status.APSSID)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to disable AP mode: %v", err)
			}
		}
		if !hasClient {
			return fmt.Errorf("no active client connection")
		}
		time.Sleep(time.Second)
		newMode := getWifiMode(nm.status.APSSID)
		if newMode != "inactive" && newMode != "client" {
			return fmt.Errorf("mode change verification failed")
		}
	default:
		return fmt.Errorf("unsupported mode: %s", mode)
	}

	return nil
}

// Creates a new AP connection for wlan0 if it doesn't exist
func (nm *networkManager) SetupAPConnection() error {
	// Check if AP connection already exists
	cmd := exec.Command("nmcli", "connection", "show", nm.status.APSSID)
	if err := cmd.Run(); err == nil {
		return nil
	}

	// Remove all existing AP interfaces, PiFi-AP-*
	removeExistingAPs()

	// Create AP connection with required settings
	cmd = exec.Command("nmcli", "connection", "add",
		"type", "wifi",
		"ifname", "wlan0",
		"con-name", nm.status.APSSID,
		"autoconnect", "no",
		"ssid", nm.status.APSSID,
		"mode", "ap",
		"ipv4.method", "shared",
		"ipv6.method", "disabled",
		"802-11-wireless.band", "bg",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create AP connection: %v\nOutput: %s", err, output)
	}

	cmd = exec.Command("nmcli", "connection", "show", nm.status.APSSID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("AP connection verification failed: %v", err)
	}
	return nil
}

// Scan for available networks and returns a list of SSIDs
func (nm *networkManager) FindAvailableNetworks() ([]string, error) {
	// Perform a network rescan
	scanCmd := exec.Command("nmcli", "device", "wifi", "rescan")
	if err := scanCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to initiate network scan: %v", err)
	}
	time.Sleep(2 * time.Second)

	// List available networks
	cmd := exec.Command("nmcli", "--fields", "SSID", "device", "wifi", "list", "--rescan", "yes")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list available networks: %v", err)
	}

	seenNetworks := make(map[string]bool)
	networks := make([]string, 0)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		ssid := strings.TrimSpace(line)
		if ssid != "" && ssid != "SSID" && !seenNetworks[ssid] {
			seenNetworks[ssid] = true
			networks = append(networks, ssid)
		}
	}

	return networks, nil
}

// Get a list of configured connections
func (nm *networkManager) GetConfiguredConnections() ([]ConnectionInfo, error) {
	cmd := exec.Command("nmcli", "-t", "-f", "NAME,TYPE,DEVICE", "connection", "show")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list configured connections: %v", err)
	}

	connections := make([]ConnectionInfo, 0)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Split(line, ":")
		if len(fields) >= 2 && fields[1] == "802-11-wireless" {
			connName := fields[0]
			pskCmd := exec.Command("nmcli", "-t", "-f", "802-11-wireless-security.psk", "connection", "show", connName)
			pskOutput, _ := pskCmd.Output()
			password := strings.TrimSpace(string(pskOutput))
			connections = append(connections, ConnectionInfo{
				SSID:     connName,
				Password: password,
			})
		}
	}

	return connections, nil
}

// Modify a connection if it exists, otherwise create a new one
func (nm *networkManager) ModifyNetworkConnection(ssid, password string, autoConnect bool) error {
	checkCmd := exec.Command("nmcli", "connection", "show", ssid)
	if err := checkCmd.Run(); err == nil {
		// Connection exists - modify it
		args := []string{"connection", "modify", ssid}
		if password != "" {
			args = append(args,
				"802-11-wireless-security.key-mgmt", "wpa-psk",
				"802-11-wireless-security.psk", password)
		}
		args = append(args, "connection.autoconnect",
			map[bool]string{true: "yes", false: "no"}[autoConnect])

		cmd := exec.Command("nmcli", args...)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to modify connection: %v\nOutput: %s", err, output)
		}
		return nil
	}

	// Connection doesn't exist - create new
	args := []string{
		"connection", "add",
		"type", "wifi",
		"ifname", "wlan0",
		"con-name", ssid,
		"autoconnect", map[bool]string{true: "yes", false: "no"}[autoConnect],
		"ssid", ssid,
	}

	if password != "" {
		args = append(args,
			"802-11-wireless-security.key-mgmt", "wpa-psk",
			"802-11-wireless-security.psk", password)
	}

	cmd := exec.Command("nmcli", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create connection: %v\nOutput: %s", err, output)
	}

	return nil
}

// Remove a saved connection by name
func (nm *networkManager) RemoveNetworkConnection(ssid string) error {
	cmd := exec.Command("nmcli", "connection", "delete", ssid)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete connection: %v", err)
	}
	return nil
}

// Set autoconnect for a saved connection by name
func (nm *networkManager) SetAutoConnectConnection(ssid string, autoConnect bool) error {
	autoConnectStr := "no"
	if autoConnect {
		autoConnectStr = "yes"
	}

	cmd := exec.Command("nmcli", "connection", "modify", ssid,
		"connection.autoconnect", autoConnectStr)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set autoconnect for %s: %v\nOutput: %s",
			ssid, err, output)
	}

	return nil
}

// Connect to a saved network by name
func (nm *networkManager) ConnectNetwork(ssid string) error {
	cmd := exec.Command("nmcli", "connection", "up", ssid)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v\nOutput: %s", ssid, err, output)
	}
	return nil
}

// Enable the AP if there's no internet connection for a certain amount of time. This will run in the background.
func (nm *networkManager) ManageOfflineAP(connectionLossTimeout time.Duration) error {
	for {
		apMode := getWifiMode(nm.status.APSSID)
		if !nm.checkWlanConnection() && apMode != "ap" {
			log.Println("Device offline, waiting for recovery...")
			time.Sleep(connectionLossTimeout)
			if !nm.checkWlanConnection() {
				log.Println("No connection after timeout, enabling AP mode")
				if err := nm.ConnectNetwork(nm.status.APSSID); err != nil {
					log.Printf("Failed to enable AP mode: %v", err)
				}
			} else {
				log.Println("Device connection recovered")
			}
		}
		time.Sleep(60 * time.Second)
	}
}
