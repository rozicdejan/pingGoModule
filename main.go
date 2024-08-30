package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-ping/ping"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

// Device struct with description and IP
type Device struct {
	Description string `yaml:"description"`
	IP          string `yaml:"ip"`
}

// Config struct for reading devices from the YAML file
type Config struct {
	Devices []Device `yaml:"devices"`
}

// TelegramMessage struct to format the message payload
type TelegramMessage struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

// readConfig reads the devices.yaml file and parses the devices with descriptions and IPs
func readConfig(filename string) (*Config, error) {
	config := &Config{}
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}
	err = yaml.Unmarshal(file, config)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal config: %w", err)
	}
	return config, nil
}

// icmpPing pings a single device using ICMP
func icmpPing(ip string) bool {
	pinger, err := ping.NewPinger(ip)
	if err != nil {
		fmt.Printf("Failed to create pinger: %v\n", err)
		return false
	}
	pinger.Count = 3
	pinger.Timeout = 5 * time.Second
	pinger.SetPrivileged(true) // Required for Windows; on Linux, it's needed to run as root or with sudo

	err = pinger.Run()
	if err != nil {
		fmt.Printf("Ping failed: %v\n", err)
		return false
	}
	stats := pinger.Statistics()
	return stats.PacketsRecv > 0
}

// sendTelegramMessage sends a message to the specified Telegram chat
func sendTelegramMessage(botToken, chatID, message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	msg := TelegramMessage{
		ChatID: chatID,
		Text:   message,
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("could not encode message to JSON: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("could not send message to Telegram: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// monitorDevices pings each device every 30 seconds, prints their statuses in a table, and sends a Telegram notification on status change
func monitorDevices(devices []Device, botToken, chatID string) {
	statuses := make(map[string]string)

	for {
		// Create a buffer to store the Telegram message
		var messageBuilder strings.Builder

		// Print table header
		fmt.Printf("\n| %-20s | %-15s | %-10s |\n", "Description", "Device IP", "Status")
		fmt.Println("|----------------------|-----------------|--------------|")

		messageChanged := false

		for _, device := range devices {
			isOnline := icmpPing(device.IP)
			status := "offline"
			statusEmoji := "🔴  " // Red circle for offline

			if isOnline {
				status = "online"
				statusEmoji = "🟢  " // Green circle for online
			}

			fmt.Printf("| %-20s | %-15s | %s%-10s%s |\n", device.Description, device.IP, statusEmoji, status, "")

			// Check if the status has changed
			if previousStatus, exists := statuses[device.IP]; !exists || previousStatus != status {
				// Append status to the message builder
				messageBuilder.WriteString(fmt.Sprintf("%s Description: %s, IP: %s is %s\n", statusEmoji, device.Description, device.IP, status))
				statuses[device.IP] = status
				messageChanged = true
			}
		}

		// Send the message if it's the first run or if there was a status change
		if messageChanged || len(statuses) == 0 {
			message := messageBuilder.String()
			err := sendTelegramMessage(botToken, chatID, message)
			if err != nil {
				fmt.Printf("Error sending Telegram message: %v\n", err)
			}
		}

		// Print a separator and wait 30 seconds
		fmt.Println("===================================")
		time.Sleep(30 * time.Second)
	}
}

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading .env file: %v\n", err)
		return
	}

	// Retrieve bot token and chat ID from environment variables
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")

	if botToken == "" || chatID == "" {
		fmt.Println("Telegram bot token or chat ID is missing in the environment variables")
		return
	}

	// Load the device list from devices.yaml
	config, err := readConfig("devices.yaml")
	if err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		return
	}

	// Monitor all devices in a single loop
	monitorDevices(config.Devices, botToken, chatID)

	// Keep the main function running (not necessary here since monitorDevices blocks)
	select {}
}
