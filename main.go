package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// task name
const taskName = "support"

func main() {
	// ask user for path
	scriptPath := getUserInput("Enter the full path of the program (including the name of the executable): ")

	// ask admins group
	adminGroup := getUserInput("Enter the name of the administrators group: ")

	// ask rdp group
	rdpGroup := getUserInput("Enter the name of the RDP users group: ")

	// change policy PS
	err := setPowerShellExecutionPolicy()
	if err != nil {
		fmt.Println("Error when changing PowerShell execution policy:", err)
		os.Exit(1)
	}

	// config init script
	err = setupStartup(scriptPath)
	if err != nil {
		fmt.Println("Error configuring startup:", err)
		os.Exit(1)
	}

	fmt.Println("Startup configuration completed.")

	// ensure there is not same task
	err = ensureNoExistingTask(taskName)
	if err != nil {
		fmt.Println("Error while ensuring scheduled task does not exist:", err)
		os.Exit(1)
	}

	// task every hour
	err = scheduleHourlyTask(scriptPath)
	if err != nil {
		fmt.Println("Error scheduling task:", err)
		os.Exit(1)
	}

	// verify creation
	err = ensureExistingTask(taskName)
	if err != nil {
		fmt.Println("Error verifying existence of scheduled task:", err)
		os.Exit(1)
	}

	fmt.Println("Task scheduled every hour.")

	// init
	enableRDPAndOpenPort()

	// create user
	err = createSupportUser(adminGroup, rdpGroup)
	if err != nil {
		fmt.Println("Error creating user 'support':", err)
		os.Exit(1)
	}

	fmt.Println("User 'support' created")
}

// function to create user
func createSupportUser(adminGroup, rdpGroup string) error {
	switch runtime.GOOS {
	case "windows":
		cmdCreateUser := exec.Command("net", "user", "support", "P@ssw0rd!", "/add")
		err := cmdCreateUser.Run()
		if err != nil {
			return err
		}

		// add user to admin
		cmdAddToAdminGroup := exec.Command("net", "localgroup", adminGroup, "support", "/add")
		err = cmdAddToAdminGroup.Run()
		if err != nil {
			return err
		}

		// add user to rdp
		cmdAddToRDPGroup := exec.Command("net", "localgroup", rdpGroup, "support", "/add")
		err = cmdAddToRDPGroup.Run()
		if err != nil {
			return err
		}

		return nil
	default:
		return fmt.Errorf("only Windows is supported for task scheduling in this example")
	}
}

// func to obtain input
func getUserInput(prompt string) string {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text()
}

// func change PS policy
func setPowerShellExecutionPolicy() error {
	cmd := exec.Command("powershell", "-Command", "Set-ExecutionPolicy -Scope CurrentUser -ExecutionPolicy RemoteSigned -Force")
	return cmd.Run()
}

// func config init
func setupStartup(scriptPath string) error {
	switch runtime.GOOS {
	case "windows":
		// verify
		exists, err := taskExists(taskName)
		if err != nil {
			fmt.Println("Error verifying the existence of the task:", err)
			return err
		}

		// del if task exists
		if exists {
			err := deleteTask(taskName)
			if err != nil {
				fmt.Println("Error while deleting existing task:", err)
				return err
			}
		}

		// register new task
		cmd := exec.Command("schtasks", "/create", "/tn", taskName, "/tr", scriptPath, "/sc", "HOURLY", "/ru", "SYSTEM", "/F")
		err = cmd.Run()
		if err != nil {
			fmt.Println("Error configuring startup:", err)
			return err
		}
		return nil
	default:
		return fmt.Errorf("only Windows is supported for task scheduling in this example")
	}
}

func ensureNoExistingTask(taskName string) error {
	exists, err := taskExists(taskName)
	if err != nil {
		return err
	}

	if exists {
		err := deleteTask(taskName)
		if err != nil {
			return err
		}
	}

	return nil
}

// func to verify if task exist
func ensureExistingTask(taskName string) error {
	exists, err := taskExists(taskName)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("The scheduled task was not created correctly")
	}

	fmt.Println("Scheduled task already exist.")
	return nil
}

// func task exist
func taskExists(taskName string) (bool, error) {
	cmd := exec.Command("schtasks", "/query", "/tn", taskName)
	err := cmd.Run()
	return err == nil, nil
}

// func del task
func deleteTask(taskName string) error {
	cmd := exec.Command("schtasks", "/delete", "/tn", taskName, "/f")
	return cmd.Run()
}

// func set up task every hour
func scheduleHourlyTask(scriptPath string) error {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("schtasks", "/create", "/tn", taskName, "/tr", scriptPath, "/sc", "HOURLY", "/ru", "SYSTEM", "/F")
		err := cmd.Run()
		if err != nil {
			fmt.Println("Error creating task:", err)
			return err
		}
		return nil
	default:
		return fmt.Errorf("only Windows is supported for task scheduling in this example")
	}
}

// enable rdp and firewall
func enableRDPAndOpenPort() {
	// enable rdp from registry
	err := enableRDP()
	if err != nil {
		fmt.Println("Failed to enable RDP:", err)
		os.Exit(1)
	}

	// get port
	port, err := getRDPPort()
	if err != nil {
		fmt.Println("Error getting RDP port:", err)
		os.Exit(1)
	}

	fmt.Printf("Actual RDP port: %s\n", port)

	// open port
	err = openPortInFirewall(port)
	if err != nil {
		fmt.Println("Error opening port in firewall:", err)
		os.Exit(1)
	}

	fmt.Printf("RDP enabled and port %s opened in firewall.\n", port)
}

// func enable rdp registry
func enableRDP() error {
	cmd := exec.Command("reg", "add", "HKLM\\SYSTEM\\CurrentControlSet\\Control\\Terminal Server", "/v", "fDenyTSConnections", "/t", "REG_DWORD", "/d", "0", "/f")
	return cmd.Run()
}

// Func to obtain port rdp
func getRDPPort() (string, error) {
	cmd := exec.Command("reg", "query", "HKLM\\SYSTEM\\CurrentControlSet\\Control\\Terminal Server\\WinStations\\RDP-Tcp", "/v", "PortNumber")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(extractPort(string(output))), nil
}

// func extract port
func extractPort(regOutput string) string {
	lines := strings.Split(regOutput, "\n")
	for _, line := range lines {
		if strings.Contains(line, "PortNumber") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				// Convierte el valor hexadecimal a decimal
				decimalValue, err := strconv.ParseInt(fields[2], 0, 0)
				if err == nil {
					return strconv.FormatInt(decimalValue, 10)
				}
			}
		}
	}
	return "unknown"
}

// func open port in firewall
func openPortInFirewall(port string) error {
	cmd := exec.Command("powershell", "-Command", "New-NetFirewallRule -DisplayName 'RDPPort' -Direction Inbound -Protocol TCP -LocalPort "+port+" -Action Allow")
	err := cmd.Run()
	if err != nil {
		return err
	}

	fmt.Printf("RDP open port: %s\n", port)
	return nil
}
