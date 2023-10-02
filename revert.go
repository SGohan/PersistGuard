package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// Constant task name
const taskName = "support"

func main() {
	// Ask the user for the script path
	scriptPath := getUserInput("Enter the full path of the program (including the name of the executable): ")

	// Revert the process
	err := revertProcess(scriptPath)
	if err != nil {
		fmt.Printf("Error reverting the process: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Process reverted successfully.")
}

// Function to get user input
func getUserInput(prompt string) string {
	fmt.Print(prompt)
	var input string
	fmt.Scanln(&input)
	return input
}

// Function to kill the scheduled task
func killScheduledTask() error {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("schtasks", "/Delete", "/TN", taskName, "/F")
		return cmd.Run()
	default:
		return fmt.Errorf("only Windows is supported for killing scheduled tasks in this example")
	}
}

// Function to close the port in the firewall
func closePortInFirewall(port string) error {
	cmd := exec.Command("netsh", "advfirewall", "firewall", "delete", "rule", "name=RDPPort", "protocol=TCP", "localport="+port)
	return cmd.Run()
}

// Function to disable RDP
func disableRDP() error {
	cmd := exec.Command("reg", "add", "HKLM\\SYSTEM\\CurrentControlSet\\Control\\Terminal Server", "/v", "fDenyTSConnections", "/t", "REG_DWORD", "/d", "1", "/f")
	return cmd.Run()
}

// Function to delete the "support" user
func deleteSupportUser() error {
	cmd := exec.Command("net", "user", "support", "/delete")
	return cmd.Run()
}

// Function to get the RDP port number
func getRDPPort() (string, error) {
	cmd := exec.Command("reg", "query", "HKLM\\SYSTEM\\CurrentControlSet\\Control\\Terminal Server\\WinStations\\RDP-Tcp", "/v", "PortNumber")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running reg query command: %v", err)
	}

	return strings.TrimSpace(extractPort(string(output))), nil
}

// Function to extract the port number from the reg query command output
func extractPort(regOutput string) string {
	lines := strings.Split(regOutput, "\n")
	for _, line := range lines {
		if strings.Contains(line, "PortNumber") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				// Convert hexadecimal value to decimal
				decimalValue, err := strconv.ParseInt(fields[2], 0, 0)
				if err == nil {
					return strconv.FormatInt(decimalValue, 10)
				}
			}
		}
	}
	return "unknown"
}

// Function to revert the process
func revertProcess(scriptPath string) error {
	// Kill the scheduled task
	err := killScheduledTask()
	if err != nil {
		return fmt.Errorf("error killing the scheduled task: %v", err)
	}

	// Close the port in the firewall
	port, err := getRDPPort()
	if err != nil {
		return fmt.Errorf("error getting the RDP port: %v", err)
	}
	err = closePortInFirewall(port)
	if err != nil {
		return fmt.Errorf("error closing the port in the firewall: %v", err)
	}

	// Disable RDP
	err = disableRDP()
	if err != nil {
		return fmt.Errorf("error disabling RDP: %v", err)
	}

	// del the "support" user
	err = deleteSupportUser()
	if err != nil {
		return fmt.Errorf("error deleting the 'support' user: %v", err)
	}

	return nil
}
