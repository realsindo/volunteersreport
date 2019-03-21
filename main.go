package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/micro/go-config"
)

//Used constant
const cTimeOut = "TimeOut"
const cErrorCreatingRequest = "ErrorCreatingRequest"
const cErrorRequest = "ErrorRequest"
const cReporterUser = "ReporterUser"
const cReporterPassword = "ReporterPassword"
const cURLTeam = "UrlTeam"
const cErrorFile = "ErrorFile"
const cOutputFile = "OutputFile"
const cSeparator = "Separator"
const cLineEnd = "LineEnd"
const cUsageLine1 = "Config file parameter missing"
const cUsageLine2 = "Usage: %s config_file\n"
const cUsageLine3 = "config_file - path to config file\n"

//Team struct is representation of Team
type Team struct {
	ID              uint             `json:"id"`
	Identifier      string           `json:"identifier"`
	Name            string           `json:"name"`
	VolunteerEmails []VolunteerEmail `json:"volunteeremails"`
}

//VolunteerEmail struct is representation of volunter assigned to the team
type VolunteerEmail struct {
	TeamID         uint
	VolunteerEmail string `json:"volunteeremail"`
}

func main() {

	//Checks if has argument for config
	if len(os.Args) < 2 {
		fmt.Println(cUsageLine1)
		progname, _ := os.Executable()
		fmt.Printf(cUsageLine2, path.Base(progname))
		fmt.Println(cUsageLine3)
		os.Exit(0)
	}

	//Loading config
	config.LoadFile(os.Args[1])

	//get Teams
	teamURL := getCfgString(cURLTeam)
	data := getReportData(teamURL)
	var tms []Team
	json.Unmarshal(data, &tms)

	writeLogFile(&tms)
}

//get string value from config
func getCfgString(name ...string) string {
	return getCfgStringDefault("", name...)
}

//get string value from config with default value
func getCfgStringDefault(def string, name ...string) string {
	return config.Get(name...).String(def)
}

//get int value from config
func getCfgInt(name ...string) int {
	return getCfgIntDefault(0, name...)
}

//get int value from config with default value
func getCfgIntDefault(def int, name ...string) int {
	return config.Get(name...).Int(def)
}

//set up basic authentification
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

//call rest api for retreivenig data
func getReportData(resturl string) []byte {
	//sets timeout
	tmOut := time.Duration(getCfgIntDefault(180, cTimeOut)) * time.Second
	client := &http.Client{
		Timeout: tmOut,
	}

	req, err := http.NewRequest("GET", resturl, nil)
	if err != nil {
		log.Fatalf(getCfgString(cErrorCreatingRequest), err)
	}
	req.Header.Add("Authorization", "Basic "+basicAuth(getCfgString(cReporterUser), getCfgString(cReporterPassword)))

	var resp *http.Response
	resp, err = client.Do(req)
	if err != nil {
		log.Fatalf(getCfgString(cErrorRequest), err)
	}
	data, _ := ioutil.ReadAll(resp.Body)

	return data
}

//Error checking
func checkError(err error) {
	if err != nil {
		log.Panicf(getCfgString(cErrorFile), err)
	}
}

//Writing log File
func writeLogFile(tms *[]Team) {
	if tms == nil {
		return
	}
	sep := getCfgStringDefault(",", cSeparator)
	eln := getCfgStringDefault("\n", cLineEnd)
	f, err := os.Create(getCfgString(cOutputFile))
	checkError(err)
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, tm := range *tms {
		_, err = w.WriteString(tm.Name)
		checkError(err)
		_, err = w.WriteString(sep)
		checkError(err)
		_, err = w.WriteString(strconv.Itoa(len(tm.VolunteerEmails)))
		checkError(err)
		_, err = w.WriteString(eln)
		checkError(err)
	}
	w.Flush()
}
