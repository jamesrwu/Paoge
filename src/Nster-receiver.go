package main

import (
	"data"
	"fmt"
	//"goconf/conf"
	"net/http"
	"log"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"upper.io/db"
	"upper.io/db/mysql"
)

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

func addService(check) {

}

func receiveChecks(rw http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var check data.ServiceResults
	err := decoder.Decode(&check)
	if err != nil {	fmt.Println("Error")	}
	addService(check)
}

func main() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)

	go signalCatcher()
	go signalReload()

	var settings = db.Settings{
		Host:     "localhost",  // MySQL server IP or name.
		Database: "Paoge",    // Database name.
		User:     "root",     // Optional user name.
		Password: "test",     // Optional user password.
	}
	session, err := db.Open("mysql", settings)
	defer session.Close()
	if err != nil { fmt.Println("DB Connection Error")}
	serviceCollection, err := session.Collection("service_results")
	if err != nil { fmt.Println("DB Table Connection Error")}
	serviceCollection.Append(data.ServiceResults{})

	http.HandleFunc("/nster/receive", receiveChecks).Name()
	log.Fatal(http.ListenAndServe(":9000", nil))
	select {}
}
