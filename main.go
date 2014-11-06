package main

import (
	"code.google.com/p/gcfg"
	"database/sql"
	"flag"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/codegangsta/martini"
	"github.com/davecheney/profile"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

const MAX_CONNS int = 50
const MAX_POST int64 = 9000000

var config struct {
	Postgresql struct {
		Host     string
		Port     int
		User     string
		Dbname   string
		Password string
		MaxConns int
	}
	Server struct {
		Port            int
		Profile         int
		Puppetfqdn      string
		Reportprocessor string
		Factsprocessor  string
	}
}

var db *sql.DB

func PullJsonValue(rw http.ResponseWriter, js *simplejson.Json, field string, t string) (v interface{}) {
	var err error
	if val, ok := js.CheckGet(field); ok {
		switch t {
		default:
			v, err = val.String()
		case "string":
			v, err = val.String()
		case "int":
			v, err = val.Int()
		case "float":
			v, err = val.Float64()
		}

		if err != nil || v == nil {
			var s string
			pretty, eperr := js.EncodePretty()
			if eperr == nil {
				fmt.Sprintf(s, "Failed to parse '%s' in the JSON, I got the following JSON: ", pretty)
			} else {
				s = "Failed to parse JSON"
			}
			fmt.Fprint(rw, s)
		}
	} else {
		fmt.Fprint(rw, fmt.Sprintf("failed to find the field '%s' in the JSON", field))
	}
	return v
}

func main() {
	config_file := "/etc/pvc-server.conf"

	c := flag.String("config", config_file, "location of config file")
	flag.Parse()

	err := gcfg.ReadFileInto(&config, *c)

	log.Printf("config file being used is @ %s", *c)

	if err != nil {
		log.Fatalf("Failed to parse %s (%s)!", *c, err)
	}

	m := martini.Classic()

	db = GetDB()

	if config.Server.Puppetfqdn == "" {
		config.Server.Puppetfqdn = "puppet"
	}

	if config.Postgresql.MaxConns == 0 {
		config.Postgresql.MaxConns = MAX_CONNS
	}

	if config.Server.Profile != 0 {
		cfg := profile.Config{
			MemProfile:  true,
			CPUProfile:  true,
			ProfilePath: ".",
		}
		p := profile.Start(&cfg)
		defer p.Stop()
	}

	db.SetMaxIdleConns(config.Postgresql.MaxConns)
	db.SetMaxOpenConns(config.Postgresql.MaxConns)

	m.Post("/ppm", RecordPPMHealth)
	m.Get("/host/:name", HostRun)
	m.Post("/report/:name", RecordReport)
	m.Post("/facts/:name", RecordFacts)
	m.Use(martini.Static("assets"))
	{
		var port = 0
		if config.Server.Port == 0 {
			port = 8080
		} else {
			port = config.Server.Port
		}
		http.ListenAndServe(fmt.Sprintf(":%d", port), m)
	}
}

func RecordPPMHealth(rw http.ResponseWriter, r *http.Request) (int, string) {
	js, err := simplejson.NewFromReader(r.Body)

	if err != nil {
		return 400, "Failed to parse JSON"
	}

	stats := new(PpmStats)
	stats.Alive = PullJsonValue(rw, js, "passenger_active", "int").(int)
	stats.GlobalProcesses = PullJsonValue(rw, js, "global_process_count", "int").(int)
	stats.AppProcesses = PullJsonValue(rw, js, "application_active_processes", "int").(int)
	stats.MaxAppProcesses = PullJsonValue(rw, js, "application_enabled_process_count", "int").(int)
	stats.ApplicationProcessed = PullJsonValue(rw, js, "application_processed", "int").(int)
	stats.AppWaitList = PullJsonValue(rw, js, "application_get_wait_list_size", "int").(int)
	load := PullJsonValue(rw, js, "system_load", "float").(float64)
	stats.SystemLoad = int(load)
	pretty, _ := js.EncodePretty()
	fmt.Printf("%s\n", pretty)
	fmt.Printf("load: %d\n", stats.SystemLoad)
	stats.Fqdn = PullJsonValue(rw, js, "fqdn", "string").(string)
	return RegisterPpmStats(stats), ""
}

func RecordFacts(rw http.ResponseWriter, r *http.Request) (int, string) {
	if config.Server.Factsprocessor != "" {
		p, err := ioutil.ReadAll(http.MaxBytesReader(rw, r.Body, MAX_POST))

		if err != nil {
			return 400, string(err.Error())
		}

		a := []rune(string(p))
		var ret []string

		var res string
		for i, r := range a {
			res = res + string(r)
			if i > 0 && (i+1)%4000 == 0 {
				ret = append(ret, res)
				res = ""
			}
		}
		if res != "" {
			ret = append(ret, res)
		}

		if err == nil {
			go func(facts []byte) {
				cmd := exec.Command(config.Server.Factsprocessor)
				cmd.Stdin = strings.NewReader(fmt.Sprintf("%s\n", strings.Join(ret, "\n")))
				cmd.Run()
			}(p)
		}
	}

	return 200, "OK"
}

func RecordReport(rw http.ResponseWriter, r *http.Request) (int, string) {
	if config.Server.Reportprocessor != "" {
		p, err := ioutil.ReadAll(http.MaxBytesReader(rw, r.Body, MAX_POST))

		if err != nil {
			return 400, string(err.Error())
		}

		if err == nil {
			go func(report []byte) {
				cmd := exec.Command(config.Server.Reportprocessor)
				cmd.Stdin = strings.NewReader(fmt.Sprintf("%s\n", report))
				cmd.Run()
			}(p)
		}
	}

	return 200, "OK"
}

func HostRun(params martini.Params) (int, string) {
	certname := params["name"]
	return HostVars(certname)
}
