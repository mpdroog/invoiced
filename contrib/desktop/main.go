package main

import (
	"flag"
	"fmt"
	"os/exec"
	"github.com/getlantern/systray"
	//"github.com/getlantern/systray/icon"
	"github.com/skratchdot/open-golang/open"
	"bufio"
	"os"
)

var (
	verbose bool
	dbPath string
)

func onReady() {
	// Start invoicing service
	arg := ""
	if verbose {
		arg = "-v"
	}
	cmd := exec.Command("../../invoiced", arg, "-d="+dbPath)

	// stderr
	stderr, e := cmd.StderrPipe()
	if e != nil {
		panic(e)
	}
	defer stderr.Close()
	go func() {
		scanner := bufio.NewScanner(stderr)
	    for scanner.Scan() {
	        fmt.Printf("ERR>> %s\n", scanner.Text())
	    }
	}()

	// stdout
	stdout, e := cmd.StdoutPipe()
	if e != nil {
		panic(e)
	}
	defer stdout.Close()
	go func() {
		scanner := bufio.NewScanner(stdout)
	    for scanner.Scan() {
	        fmt.Printf("OUT>> %s\n", scanner.Text())
	    }
	}()

	if e := cmd.Start(); e != nil {
		panic(e)
	}

	//systray.SetIcon(icon.Data)
	systray.SetTitle("$$$")
	systray.SetTooltip("InvoiceD")

	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")
	mBrowser := systray.AddMenuItem("Open", "")

	go func() {
		for {
			select {
			case <-mQuit.ClickedCh:
				if verbose {
					fmt.Println("cmd=quit")
				}
				systray.Quit()

				if e := cmd.Process.Kill(); e != nil {
					panic(e)
				}
				// stop the for..
				return

			case <-mBrowser.ClickedCh:
				if verbose {
					fmt.Println("cmd=open")
				}
				if e := open.Run("http://localhost:9999/static/"); e != nil {
					panic(e)
				}
			}
		}
	}()

	if e := cmd.Wait(); e != nil {
		if e.Error() != "signal: killed" {
			panic(e)
		}
	}

	// TODO: Hacky force quit
	os.Exit(0)
}

func main() {
	flag.BoolVar(&verbose, "v", false, "Verbose-mode (log more)")
	flag.StringVar(&dbPath, "d", "billingdb", "Path to database")
	flag.Parse()
	if verbose {
		fmt.Printf("Verbose-mode\n")
	}
	systray.Run(onReady)
}