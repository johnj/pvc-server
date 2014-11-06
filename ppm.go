package main

import (
	"math"
)

type PpmStats struct {
	Alive                int
	GlobalProcesses      int
	AppProcesses         int
	MaxAppProcesses      int
	AppWaitList          int
	ApplicationProcessed int
	PreviouslyProcessed  int
	SystemLoad           int
	Weight               int
	Fqdn                 string
}

func CalcPPMScore(stats *PpmStats) int {
	if stats.MaxAppProcesses == 0 {
		return 0
	}
	d := ((float64(stats.MaxAppProcesses) - float64(stats.AppProcesses)) / float64(stats.MaxAppProcesses)) * math.Pow(.9, float64(stats.SystemLoad))
	return int(d * 10000)
}

func BestAvailablePpm() string {
	row, err := db.Query("select fqdn from ppms where updated_at > current_timestamp - interval '90 seconds' order by score desc limit 1")

	if err != nil {
		panic(err)
	}

	var fqdn string

	for row.Next() {
		if err := row.Scan(&fqdn); err != nil {
			panic(err)
		}
		break
	}

	if err := row.Err(); err != nil {
		panic(err)
	}

	row.Close()

	return fqdn
}

func AddRunToPpm(ppm string) (bool, error) {
	stats := &PpmStats{}

	row, err := db.Query("select active_processes, allowed_processes, loadavg from ppms where fqdn=$1", ppm)

	if err != nil {
		panic(err)
	}

	for row.Next() {
		row.Scan(&stats.AppProcesses, &stats.MaxAppProcesses, &stats.SystemLoad)
	}

	stats.AppProcesses++
	stats.MaxAppProcesses++
	stats.SystemLoad++

	score := CalcPPMScore(stats)

	_, err = db.Exec("update ppms set active_processes=active_processes+1, loadavg=loadavg+1, score=$1 where fqdn=$2", score, ppm)

	if err != nil {
		panic(err)
	}

	if err == nil {
		return true, nil
	}
	return false, err
}

func RegisterPpmStats(stats *PpmStats) int {
	row, err := db.Query("select processed_requests, weighting from ppms where fqdn = $1", stats.Fqdn)

	if err != nil {
		panic(err)
	}

	var found bool

	for row.Next() {
		if err := row.Scan(&stats.PreviouslyProcessed, &stats.Weight); err != nil {
			panic(err)
		}
		found = true
		break
	}

	if err := row.Err(); err != nil {
		panic(err)
	}

	row.Close()

	if stats.Weight == 0 {
		stats.Weight = 1
	}
	if stats.SystemLoad == 0 {
		stats.SystemLoad = 1
	}

	score := CalcPPMScore(stats)

	var qry string

	if found {
		qry = `update ppms set active_processes = $1, waiting_processes = $2, loadavg = $3, allowed_processes = $4,
			processed_requests = $5, updated_at = current_timestamp, score = $6
			where fqdn = $7`
	} else {
		qry = `insert into ppms
			(active_processes, waiting_processes, loadavg, allowed_processes, processed_requests, score, fqdn, updated_at, created_at)
			values
			($1, $2, $3, $4, $5, $6, $7, current_timestamp, current_timestamp)`
	}

	q, err := db.Query(qry, stats.AppProcesses, stats.AppWaitList, stats.SystemLoad, stats.MaxAppProcesses, stats.ApplicationProcessed, score, stats.Fqdn)

	if err != nil {
		panic(err)
	}

	q.Close()

	return 200
}
