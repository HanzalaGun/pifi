# PiFi

Modern headless WiFi configuration tool for Raspberry Pi.    
Remotely manage IoT projects without physical access to the device.

Works with Bookworm using NetworkManager.  
Tested on Raspberry Pi: 2B, Zero W, Zero 2 W, and 5. 

## Key Features
- Web Interface for WiFi management
- Systemd service for automatic WiFi configuration
- Access point mode to manage offline devices

## Web Interface

A simple web service that allows you to configure the WiFi settings of your Raspberry Pi.   
Connect to the same network as your device running PiFi then navigate to `http://<device-ip>:8088`

<img width="810" alt="image" src="https://github.com/user-attachments/assets/8a36b61c-3f19-4546-bcef-fb7426817186" />

## Systemd Service

`pifi.service` is a daemon that runs on boot and automatically configures the WiFi settings of your Raspberry Pi.
If your device is not connected to a network, the service will start an access point that you can use to configure the WiFi settings.

Connect a client to the access point and navigate to `http://10.42.0.1:8088` to view the web interface.   
The AP should act as a captive portal and redirect you to the configuration page in most cases.

### Setup

- Create the new systemd service file:   
`sudo vim /etc/systemd/system/pifi.service`

```shell
[Unit]
Description=PiFi Service
After=network.target

[Service]
ExecStart=<path-to-pifi-binary>
Environment="PATH=/usr/bin:/usr/sbin"
WorkingDirectory=<directory-of-pifi-binary>
User=root
Restart=always

[Install]
WantedBy=multi-user.target
```

- Reload systemd to recognize the new service:   
`sudo systemctl daemon-reload`

- Enable the service to start on boot:   
`sudo systemctl enable pifi.service`

- Start the service immediately:   
`sudo systemctl start pifi.service`

- Check the status of the service:   
`sudo systemctl status pifi.service`
