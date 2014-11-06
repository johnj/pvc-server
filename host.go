package main

import (
	"fmt"
	"strings"
)

func CreateHost(certname string) error {
	q, err := db.Query("insert into hosts (certname, created_at, updated_at) values ($1, current_timestamp, current_timestamp)", certname)
	q.Close()
	return err
}

func HostVars(certname string) (int, string) {
	row, err := db.Query(`select force_run, get_facts, check_interval, files_monitored,
		coalesce(extract(epoch from last_run_finished), 0) as last_run, max_runinterval
		from hosts where certname = $1`, certname)

	if err != nil {
		panic(err)
	}

	var found bool
	var force_run, fact_run, check_interval, last_run, max_runinterval int
	var files_monitored string

	for row.Next() {
		row.Scan(&force_run, &fact_run, &check_interval, &files_monitored, &last_run, &max_runinterval)
		found = true
		break
	}
	row.Close()

	if found == false {
		err := CreateHost(certname)
		if err != nil {
			return 500, string(err.Error())
		}
		force_run = 1
	}

	var ret []string

	ret = append(ret, "PVC_RETURN=0")

	if force_run != 0 || last_run == 0 {
		ret = append(ret, "PVC_RUN=1")
		ppm := BestAvailablePpm()
		if ppm != "" {
			_, err := AddRunToPpm(ppm)
			if err == nil {
				ret = append(ret, fmt.Sprintf("PVC_PPM_HOST=%s", ppm))
				ret = append(ret, fmt.Sprintf("PVC_PPM_NAME=%s", config.Server.Puppetfqdn))
			}
		}
	} else if fact_run != 0 {
		ret = append(ret, "PVC_FACT_RUN=1")
	}

	if len(files_monitored) > 0 {
		ret = append(ret, fmt.Sprintf("PVC_FILES_MONITORED='%s'", strings.Replace(files_monitored, "'", "", -1)))
	}

	if check_interval > 0 {
		ret = append(ret, fmt.Sprintf("PVC_CHECK_INTERVAL=%d", check_interval))
	}

	return 200, strings.Join(ret, "\n") + "\n"
}
