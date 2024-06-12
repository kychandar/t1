package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()

	// Create a list with an "install nginx" item and a "Quit" item
	list := tview.NewList().
		AddItem("install nginx", "", 'a', func() {
			showInstallPage(app)
		}).
		AddItem("Quit", "Press to exit", 'q', func() {
			app.Stop()
		})

	// Create a Flex layout to hold the list and set a border
	flex := tview.NewFlex().
		AddItem(list, 0, 1, true)
	flex.SetBorder(true).SetTitle("Cloudlyte Installer")

	// Create a Pages container to manage multiple pages
	pages := tview.NewPages().
		AddPage("list", flex, true, true)

	// Set the root and run the application
	if err := app.SetRoot(pages, true).SetFocus(flex).Run(); err != nil {
		panic(err)
	}
}

func showInstallPage(app *tview.Application) {
	logsTextView := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetChangedFunc(func() {
			app.Draw()
		})

	// Create a Flex layout to hold the logs and set a border
	logsFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(logsTextView, 0, 1, true)
	logsFlex.SetBorder(true).SetTitle("Installing Nginx")

	pages := tview.NewPages().
		AddPage("logs", logsFlex, true, true)

	app.SetRoot(pages, true).SetFocus(logsFlex)

	// Execute the installation script and stream logs
	go func() {
		cmd := exec.Command("/bin/bash", "install_nginx.sh")
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error obtaining stdout: %v\n", err)
			return
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error obtaining stderr: %v\n", err)
			return
		}

		if err := cmd.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting command: %v\n", err)
			return
		}

		reader := io.MultiReader(stdout, stderr)
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()
			app.QueueUpdateDraw(func() {
				fmt.Fprintf(logsTextView, "%s\n", line)
			})
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading output: %v\n", err)
		}

		if err := cmd.Wait(); err != nil {
			app.QueueUpdateDraw(func() {
				fmt.Fprintf(logsTextView, "[red]Installation failed: %v\n", err)
			})
		} else {
			app.QueueUpdateDraw(func() {
				fmt.Fprintf(logsTextView, "[green]Nginx installed successfully!\n")
			})

			// Delay before showing the success page
			time.Sleep(1 * time.Second)

			app.QueueUpdateDraw(func() {
				showSuccessPage(app)
			})
		}
	}()
}

func showSuccessPage(app *tview.Application) {
	// Obtain the local IP address
	ipAddress, err := getLocalIP()
	if err != nil {
		ipAddress = "Unknown"
	}

	// Create TextView for success message
	successText := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[green]Nginx installed successfully!")

	// Create TextView for IP address
	ipText := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(fmt.Sprintf("[white]IP Address: %s", ipAddress))

	// Create TextView for status
	statusText := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[white]Status: Running")

	// Create a Flex layout to hold the success message, IP address, and status
	successFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(successText, 1, 1, false).
		AddItem(ipText, 1, 1, false).
		AddItem(statusText, 1, 1, false)
	successFlex.SetBorder(true).SetTitle("Installation Complete")

	pages := tview.NewPages().
		AddPage("success", successFlex, true, true)

	app.QueueUpdateDraw(func() {
		app.SetRoot(pages, true).SetFocus(successFlex)
	})
}

// Helper function to get the local IP address
func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no IP address found")
}
