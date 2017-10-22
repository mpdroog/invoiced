package main

import (
	"flag"
	"fmt"
	"os/exec"
	"github.com/getlantern/systray"
	//"github.com/getlantern/systray/icon"
	"github.com/skratchdot/open-golang/open"
	"bufio"
	"github.com/mitchellh/go-homedir"
	//"github.com/lextoumbourou/idle"
	"github.com/ctcpip/notifize"
	"sync"
)

var (
	verbose bool
	dbPath string
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

	systray.Run(onReady, nil)
	// WARN: Unreachable as systray takes over control
}

func onReady() {
	name := "../../invoiced"
	args := []string {
		"-c=../../config.toml",
		"-d="+dbPath,
	}

	if verbose {
		args = append(args, "-v")
		fmt.Printf("exec=%s %s\n", name, args)
	}
	cmd := exec.Command(name, args...)

	// stderr
	stderr, e := cmd.StderrPipe()
	if e != nil {
		panic(e)
	}
	// stdout
	stdout, e := cmd.StdoutPipe()
	if e != nil {
		panic(e)
	}

	stopChan := make(chan bool, 1)
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer stdout.Close()
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
	    for scanner.Scan() {
	        fmt.Printf("OUT>> %s\n", scanner.Text())
	    }
	    if e := scanner.Err(); e != nil {
	    	fmt.Printf("scanner.Std: %s\n", e.Error())
	    }
	}()

	wg.Add(1)
	go func() {
		defer stderr.Close()
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
	    for scanner.Scan() {
	    	txt := scanner.Text()
	        fmt.Printf("ERR>> %s\n", txt)
	        notifize.Display("InvoiceD", txt, false, "")
	    }
	    if e := scanner.Err(); e != nil {
	    	fmt.Printf("scanner.Err: %s\n", e.Error())
	    }
	}()

	//systray.SetIcon(icon.Data)
	systray.SetTitle("$$$")
	systray.SetTooltip("InvoiceD")

	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")
	mBrowser := systray.AddMenuItem("Open", "")

	go func() {
		if e := cmd.Start(); e != nil {
			panic(e)
		}
		wg.Wait()
		if e := cmd.Wait(); e != nil {
			fmt.Printf("cmd.Wait: %s\n", e.Error())
		}
		stopChan <- true
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
		}
	}

	// User timer
	/*go func() {
		for {
			select {
			case <-timer:
				if verbose {
					fmt.Println("user-timer")
				}
				d, e := idle.Get()
				if d > time.Duration("5m") {
					// TODO
				}
			}
		}
	}()*/
}
