package Test3

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type loadavg struct {
	oneMin      float32
	fiveMin     float32
	fifteenMin  float32
	cpuCount    uint16
	procRunning uint32
	procTotal   uint32
}

type memstats struct {
	memTotal  uint64
	memFree   uint64
	swapTotal uint64
	swapFree  uint64
}

type filesystem struct {
	fileType string
}

func getfilesystem(fstype int64) filesystem {
	var cur_filesystem filesystem
	if fstype == 61267 {
		cur_filesystem.fileType = "ext3"
	} else {
		cur_filesystem.fileType = "unknown"
	}
	return cur_filesystem
}

func getloadavg() loadavg {
	var load loadavg
	b, err := os.Open("/proc/loadavg")
	if err != nil {
		panic(err)
	}
	defer b.Close()
	scanner := bufio.NewScanner(b)
	scanner.Scan()
	line := scanner.Text()
	sp_line := strings.Split(line, " ")
	oneMin, _ := strconv.ParseFloat(sp_line[0], 32)
	fiveMin, _ := strconv.ParseFloat(sp_line[1], 32)
	fifteenMin, _ := strconv.ParseFloat(sp_line[2], 32)
	load.oneMin, load.fiveMin, load.fifteenMin = float32(oneMin), float32(fiveMin), float32(fifteenMin)

	proc_line := strings.Split(sp_line[3], "/")
	procRunning, _ := strconv.Atoi(proc_line[0])
	procTotal, _ := strconv.Atoi(proc_line[1])
	load.procRunning, load.procTotal = uint32(procRunning), uint32(procTotal)
	return load
}

func getstatfs(path string) syscall.Statfs_t {
	s := syscall.Statfs_t{}
	syscall.Statfs(path, &s)
	return s
}

func getmegabytes(bytes uint64) uint64 {
	megabytes := bytes / 1024 / 1024
	return megabytes
}

func getmemstats() memstats {
	var mem memstats
	b, err := os.Open("/proc/meminfo")
	if err != nil {
		panic(err)
	}
	defer b.Close()

	scanner := bufio.NewScanner(b)

	// scanner.Scan() advances to the next token returning false if an error was encountered
	for scanner.Scan() {
		line := scanner.Text()
		sp_line := strings.Split(line, ":")
		if sp_line[0] == "MemTotal" {
			memTotal, _ := strconv.Atoi(strings.TrimRight(strings.Trim(sp_line[1], " "), " kB"))
			mem.memTotal = uint64(memTotal)
		} else if sp_line[0] == "MemFree" {
			memFree, _ := strconv.Atoi(strings.TrimRight(strings.Trim(sp_line[1], " "), " kB"))
			mem.memFree = uint64(memFree)
		} else if sp_line[0] == "SwapTotal" {
			swapTotal, _ := strconv.Atoi(strings.TrimRight(strings.Trim(sp_line[1], " "), " kB"))
			mem.swapTotal = uint64(swapTotal)
		} else if sp_line[0] == "SwapFree" {
			swapFree, _ := strconv.Atoi(strings.TrimRight(strings.Trim(sp_line[1], " "), " kB"))
			mem.swapFree = uint64(swapFree)
		}
	}

	// When finished scanning if any error other than io.EOF occured
	// it will be returned by scanner.Err().
	if err := scanner.Err(); err != nil {
		log.Fatal(scanner.Err())
	}
	return mem
}

func signalReload() {
	ch := make(chan os.Signal)
	for {
		signal.Notify(ch, syscall.SIGHUP)
		<-ch
		log.Println("SIGHUP")
	}
}

func signalCatcher() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGKILL)
	<-ch
	log.Println("CTRL-C; exiting")
	os.Exit(0)
}

func main() {
	// Handle SIGINT and SIGTERM.
	go signalCatcher()
	go signalReload()
	for {
		t0 := time.Now()
		fs := getstatfs("/")
		bytesLeft := fs.Bavail * uint64(fs.Bsize)
		bytesTotal := fs.Blocks * uint64(fs.Bsize)
		fmt.Println("Bytes left:", bytesLeft, ", Bytes total:", bytesTotal)

		megabytesLeft := getmegabytes(bytesLeft)
		megabytesTotal := getmegabytes(bytesTotal)
		fmt.Println("MB left:", megabytesLeft, ", MB total:", megabytesTotal)
		//fmt.Printf(getdiskspace(path))

		fmt.Println("Filesystem type:", getfilesystem(fs.Type).fileType)

		memstats := getmemstats()
		fmt.Println("Total memory:", memstats.memTotal)
		fmt.Println("Free memory:", memstats.memFree)
		fmt.Println("Total swap:", memstats.swapTotal)
		fmt.Println("Free swap:", memstats.swapFree)

		loadstats := getloadavg()
		fmt.Println("Load avg:", loadstats.oneMin, loadstats.fiveMin, loadstats.fifteenMin)
		fmt.Println("Process running: ", loadstats.procRunning, "/", loadstats.procTotal)

		t1 := time.Now()
		fmt.Printf("This call took %v to run.\n", t1.Sub(t0))
		time.Sleep(10 * time.Second)
	}
}
