package main

import (
	"flag"
	"fmt"
	"github.com/getlantern/systray"
	"os/exec"
	//"github.com/getlantern/systray/icon"
	"github.com/lextoumbourou/idle"
	"github.com/mitchellh/go-homedir"
	"github.com/skratchdot/open-golang/open"
	//"github.com/ctcpip/notifize"
	//"sync"
	"bufio"
	"os"
	"strings"
	"time"
)

var (
	verbose    bool
	dbPath     string
	timerStart *time.Time
	minute     time.Duration

	mStart     *systray.MenuItem
	curProject *map[string]string
)

func main() {
	home, e := homedir.Dir()
	if e != nil {
		panic(e)
	}
	flag.BoolVar(&verbose, "v", false, "Verbose-mode (log more)")
	flag.StringVar(&dbPath, "d", home+"/invoiced", "Path to database")
	flag.Parse()
	if verbose {
		fmt.Printf("Verbose-mode\n")
	}

	minute, e = time.ParseDuration("1m")
	if e != nil {
		panic(e)
	}

	systray.Run(onReady, nil)
	// WARN: Unreachable as systray takes over control
}

func onReady() {
	name := "../../invoiced"
	args := []string{
		"-c=../../config.toml",
		"-d=" + dbPath,
	}

	if verbose {
		args = append(args, "-v")
		fmt.Printf("exec=%s %s\n", name, args)
	}
	cmd := exec.Command(name, args...)
	stdout, e := cmd.StdoutPipe()
	if e != nil {
		panic(e)
	}
	cmd.Stderr = os.Stderr

	stopChan := make(chan bool, 2)

	systray.SetTitle("$$$")
	systray.SetTooltip("InvoiceD")

	mBrowser := systray.AddMenuItem("Open", "Open browser")
	mStart = systray.AddMenuItem("Start", "Start timer")
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")
	mStart.Disable()

	go func() {
		if e := cmd.Start(); e != nil {
			panic(e)
		}
		if e := cmd.Wait(); e != nil {
			fmt.Printf("cmd.Wait: %s\n", e.Error())
		}
		stopChan <- true
	}()

	go func() {
		// Line reader
		scanner := bufio.NewScanner(stdout)
		fmt.Printf("stdout.Wait()\n")
		for scanner.Scan() {
			line := scanner.Text()
			if verbose {
				fmt.Printf("stdout.Line=%s\n", line)
			}
			if strings.HasPrefix(line, "cmd ") {
				// Proxy-cmd
				args := make(map[string]string)
				for _, item := range strings.Split(line, " ") {
					if item == "cmd" {
						continue
					}
					tok := strings.Split(item, "=")
					args[tok[0]] = tok[1]
				}
				fmt.Printf("args=%+v\n", args)
				curProject = &args
				systray.SetTitle(fmt.Sprintf("$%s$%s$%s", args["entity"], args["year"], args["hour"]))
				mStart.Enable()
			}
		}
	}()

	// User timer
	ticker := time.NewTicker(time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				if verbose {
					fmt.Println("interval.sec")
				}
				if timerStart != nil {
					n := time.Now().Sub(*timerStart)
					systray.SetTitle(fmt.Sprintf("$%02d:%02d:%02d", int(n.Hours())%60, int(n.Minutes())%60, int(n.Seconds())%60))
				}
				d, e := idle.Get()
				if e != nil {
					panic(e)
				}
				if d > minute { //time.Duration("5m") {
					fmt.Printf("1min idle\n")
				}
			case <-stopChan:
				ticker.Stop()
				return
			}
		}
	}()

	for {
		select {
		case <-stopChan:
			systray.Quit()
			// WARN: All code below is unreachable

		case <-mQuit.ClickedCh:
			if verbose {
				fmt.Println("cmd=quit")
			}
			if e := cmd.Process.Kill(); e != nil {
				fmt.Printf("cmd.Kill: %s\n", e.Error())
			}
			stopChan <- true

		case <-mBrowser.ClickedCh:
			if verbose {
				fmt.Println("cmd=open")
			}
			if e := open.Run("http://localhost:9999/static/"); e != nil {
				panic(e)
			}

		case <-mStart.ClickedCh:
			if verbose {
				fmt.Println("cmd=start")
			}
			if timerStart != nil {
				// Stop the timer!
				n := time.Now().Sub(*timerStart)
				fmt.Printf("Duration=%s\n", n.String())
				// TODO: Now use API to load..add..save the change

				mStart.SetTitle("Start timer")
				timerStart = nil
			}
			mStart.SetTitle("Stop timer")
			n := time.Now()
			timerStart = &n
		}
	}
}
