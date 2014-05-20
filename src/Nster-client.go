package main

import (
	//"bufio"
	"bytes"
	"data"
	"flag"
	"fmt"
	//"goconf/conf"
	"net/http"
	"log"
	"encoding/json"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	//"strconv"
	"syscall"
	"time"
)

type config struct {
	intervalSeconds uint32
	pluginsFolder string
	namespacePrefix string
}

func parseArgs() {
	flag.String("config", "/etc/Paoge.conf", "Config file location")
	flag.Parse()
	fmt.Println("tail:", flag.Args())
}

func getConfigs(configFile string) config {
	fmt.Println(configFile)
	var configuration config
	configuration.intervalSeconds = 10
	configuration.pluginsFolder = "/opt/paoge"
	configuration.namespacePrefix = ""
	return configuration
}

func reloadConfig() {
	fmt.Println("Loading Configs")
}

func signalReload() {
	ch := make(chan os.Signal)
	for {
		signal.Notify(ch, syscall.SIGHUP)
		<-ch
		reloadConfig()
	}
}

func signalCatcher() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGKILL)
	<-ch
	log.Println("CTRL-C; exiting")
	os.Exit(0)
}

func getChecks(Path string, checkChannel chan string) {
	walkFn := func(path string, info os.FileInfo, err error) error {
		stat, err := os.Stat(path)
		if err != nil {	return err	}
		if stat.Mode().IsRegular() { checkChannel <- path }
		return nil
	}
	for {
		err := filepath.Walk(Path, walkFn)
		if err != nil { log.Fatal(err) }
		time.Sleep(60 * time.Second)
	}
}

//pass args?
func runChecks(rootPath string, checks chan string, checkResults chan data.ServiceResults) {
	for {
		check := <-checks
		var results data.ServiceResults
		results.Time = time.Now()
		results.CheckName = strings.Split(check, rootPath)[1]

		cmd := exec.Command(check)
		output, err := cmd.CombinedOutput()
		if err != nil {	println(err.Error()) }

		results.Stdout = string(output)
		checkResults <- results
	}
}

func sendChecks(checkResults chan data.ServiceResults) {
	client := &http.Client{}
	for {
		result := <-checkResults
		//fmt.Println("sending:", result.checkName, result.time, result.returnCode, result.stdout, result.stderr)
		jsonResult, err := json.Marshal(result)
		if err != nil { println(err) }
		fmt.Println("json: ", string(jsonResult))
		r, _ := http.NewRequest("POST", "http://localhost:9000/nster/receive",  bytes.NewBufferString(string(jsonResult)))
		resp, err := client.Do(r)
		if err != nil { fmt.Println(err) } else { fmt.Println(resp.Status) }
	}
}

func main() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
	//Checks Path channel
	checks := make(chan string, 100)
	checkResults := make(chan data.ServiceResults, 100)

	// Handle SIGINT and SIGTERM.
	go signalCatcher()
	go signalReload()

	parseArgs()
	checkRoot := "/home/james/opt/paoge/checks"
	go getChecks(checkRoot, checks)
	go runChecks(checkRoot, checks, checkResults)
	go sendChecks(checkResults)
	select {}
}
