package main

import (
	"flag"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/timesking/jstat_exporter/log"
)

const (
	namespace = "jstat"
)

var (
	listenAddress = flag.String("web.listen-address", ":9010", "Address on which to expose metrics and web interface.")
	metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	jstatPath     = flag.String("jstat.path", "/usr/bin/jstat", "jstat path")
	targetPid     = flag.String("target.pid", ":0", "target pid")
)

type Exporter struct {
	jstatPath    string
	targetPid    string
	newMax       prometheus.Gauge
	newCommit    prometheus.Gauge
	oldMax       prometheus.Gauge
	oldCommit    prometheus.Gauge
	metaMax      prometheus.Gauge
	metaCommit   prometheus.Gauge
	metaUsed     prometheus.Gauge
	oldUsed      prometheus.Gauge
	oldCap       prometheus.Gauge

	sv0Used      prometheus.Gauge
	sv0Cap       prometheus.Gauge
	sv1Used      prometheus.Gauge
	sv1Cap       prometheus.Gauge
	edenUsed     prometheus.Gauge
	edenCap      prometheus.Gauge

	fgcTimes     prometheus.Counter
	lastFgcTimes float64
	fgcSec       prometheus.Gauge
	ygcTimes     prometheus.Counter
	lastYgcTimes float64
	ygcSec       prometheus.Gauge
	gcSec        prometheus.Gauge
}

func NewExporter(jstatPath string, targetPid string) *Exporter {
	if strings.HasPrefix(targetPid, "/") && strings.HasSuffix(targetPid, ".pid") {
		if pid, err := ioutil.ReadFile(targetPid); err == nil {
			targetPid = strings.TrimSpace(string(pid))
			log.Info("Got PID from file: ", targetPid)
		}
	} else if strings.HasPrefix(targetPid, "#") {
		targetPidCmd := strings.Trim(targetPid, "#")
		out, err := exec.Command("/bin/bash", "-c", targetPidCmd).Output()
		if err != nil {
			log.Fatal(err)
		}
		targetPid = strings.TrimSpace(string(out))
		log.Info("Got PID", targetPid, ", from command: ", targetPidCmd)
	}
	return &Exporter{
		jstatPath: jstatPath,
		targetPid: targetPid,
		newMax: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "newMax",
			Help:      "newMax",
		}),
		newCommit: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "newCommit",
			Help:      "newCommit",
		}),
		oldMax: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "oldMax",
			Help:      "oldMax",
		}),
		oldCommit: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "oldCommit",
			Help:      "oldCommit",
		}),
		metaMax: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "metaMax",
			Help:      "metaMax",
		}),
		metaCommit: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "metaCommit",
			Help:      "metaCommit",
		}),
		metaUsed: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "metaUsed",
			Help:      "metaUsed",
		}),
		oldUsed: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "oldUsed",
			Help:      "oldUsed",
		}),
		oldCap: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "oldCap",
			Help:      "oldCap",
		}),		
		sv0Used: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "sv0Used",
			Help:      "sv0Used",
		}),
		sv0Cap: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "sv0Cap",
			Help:      "sv0Cap",
		}),		
		sv1Used: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "sv1Used",
			Help:      "sv1Used",
		}),
		sv1Cap: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "sv1Cap",
			Help:      "sv1Cap",
		}),		
		edenUsed: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "edenUsed",
			Help:      "edenUsed",
		}),
		edenCap: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "edenCap",
			Help:      "edenCap",
		}),

		fgcTimes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "fgcTimes",
			Help:      "fgcTimes",
		}),
		lastFgcTimes: 0.0,
		fgcSec: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "fgcSec",
			Help:      "fgcSec",
		}),
		ygcTimes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "ygcTimes",
			Help:      "ygcTimes",
		}),
		lastYgcTimes: 0.0,
		ygcSec: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "ygcSec",
			Help:      "ygcSec",
		}),
		gcSec: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "gcSec",
			Help:      "gcSec",
		}),
	}
}

// Describe implements the prometheus.Collector interface.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.newMax.Describe(ch)
	e.newCommit.Describe(ch)
	e.oldMax.Describe(ch)
	e.oldCommit.Describe(ch)
	e.metaMax.Describe(ch)
	e.metaCommit.Describe(ch)
	e.metaUsed.Describe(ch)
	e.oldUsed.Describe(ch)
	e.oldCap.Describe(ch)

	e.sv0Used.Describe(ch)
	e.sv0Cap.Describe(ch)
	e.sv1Used.Describe(ch)
	e.sv1Cap.Describe(ch)
	e.edenUsed.Describe(ch)
	e.edenCap.Describe(ch)

	e.fgcTimes.Describe(ch)
	e.fgcSec.Describe(ch)
	e.ygcTimes.Describe(ch)
	e.ygcSec.Describe(ch)
	e.gcSec.Describe(ch)
}

// Collect implements the prometheus.Collector interface.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.JstatGccapacity(ch)
	e.JstatGcold(ch)
	e.JstatGcnew(ch)
	e.JstatGc(ch)
}

func (e *Exporter) JstatGccapacity(ch chan<- prometheus.Metric) {

	out, err := exec.Command(e.jstatPath, "-gccapacity", e.targetPid).Output()
	if err != nil {
		log.Fatal(err)
	}

	for i, line := range strings.Split(string(out), "\n") {
		if i == 1 {
			parts := strings.Fields(line)
			newMax, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				log.Fatal(err)
			}
			e.newMax.Set(newMax)
			e.newMax.Collect(ch)
			newCommit, err := strconv.ParseFloat(parts[2], 64)
			if err != nil {
				log.Fatal(err)
			}
			e.newCommit.Set(newCommit)
			e.newCommit.Collect(ch)
			oldMax, err := strconv.ParseFloat(parts[7], 64)
			if err != nil {
				log.Fatal(err)
			}
			e.oldMax.Set(oldMax)
			e.oldMax.Collect(ch)
			oldCommit, err := strconv.ParseFloat(parts[8], 64)
			if err != nil {
				log.Fatal(err)
			}
			e.oldCommit.Set(oldCommit)
			e.oldCommit.Collect(ch)
			metaMax, err := strconv.ParseFloat(parts[11], 64)
			if err != nil {
				log.Fatal(err)
			}
			e.metaMax.Set(metaMax)
			e.metaMax.Collect(ch)
			metaCommit, err := strconv.ParseFloat(parts[12], 64)
			if err != nil {
				log.Fatal(err)
			}
			e.metaCommit.Set(metaCommit)
			e.metaCommit.Collect(ch)
		}
	}
}

func (e *Exporter) JstatGcold(ch chan<- prometheus.Metric) {

	out, err := exec.Command(e.jstatPath, "-gcold", e.targetPid).Output()
	if err != nil {
		log.Fatal(err)
	}

	for i, line := range strings.Split(string(out), "\n") {
		if i == 1 {
			parts := strings.Fields(line)
			metaUsed, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				log.Fatal(err)
			}
			e.metaUsed.Set(metaUsed) // MU: Metaspace utilization (kB).
			e.metaUsed.Collect(ch)
			oldUsed, err := strconv.ParseFloat(parts[5], 64)
			if err != nil {
				log.Fatal(err)
			}
			e.oldUsed.Set(oldUsed) // OU: Old space utilization (kB).
			e.oldUsed.Collect(ch)
			oldCap, err := strconv.ParseFloat(parts[4], 64)
			if err != nil {
				log.Fatal(err)
			}
			e.oldCap.Set(oldCap) // OC: Old space utilization (kB).
			e.oldCap.Collect(ch)			
		}
	}
}

func (e *Exporter) JstatGcnew(ch chan<- prometheus.Metric) {

	out, err := exec.Command(e.jstatPath, "-gcnew", e.targetPid).Output()
	if err != nil {
		log.Fatal(err)
	}

	for i, line := range strings.Split(string(out), "\n") {
		if i == 1 {
			parts := strings.Fields(line)
			sv0Used, err := strconv.ParseFloat(parts[2], 64)
			if err != nil {
				log.Fatal(err)
			}
			e.sv0Used.Set(sv0Used)
			e.sv0Used.Collect(ch)
			sv0Cap, err := strconv.ParseFloat(parts[0], 64)
			if err != nil {
				log.Fatal(err)
			}
			e.sv0Cap.Set(sv0Cap)
			e.sv0Cap.Collect(ch)

			sv1Used, err := strconv.ParseFloat(parts[3], 64)
			if err != nil {
				log.Fatal(err)
			}
			e.sv1Used.Set(sv1Used)
			e.sv1Used.Collect(ch)
			sv1Cap, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				log.Fatal(err)
			}
			e.sv1Cap.Set(sv1Cap)
			e.sv1Cap.Collect(ch)			

		
			edenUsed, err := strconv.ParseFloat(parts[8], 64)
			if err != nil {
				log.Fatal(err)
			}
			e.edenUsed.Set(edenUsed)
			e.edenUsed.Collect(ch)
			edenCap, err := strconv.ParseFloat(parts[7], 64)
			if err != nil {
				log.Fatal(err)
			}
			e.edenCap.Set(edenCap)
			e.edenCap.Collect(ch)				
		}
	}
}

func (e *Exporter) JstatGc(ch chan<- prometheus.Metric) {

	out, err := exec.Command(e.jstatPath, "-gc", e.targetPid).Output()
	if err != nil {
		log.Fatal(err)
	}

	for i, line := range strings.Split(string(out), "\n") {
		if i == 1 {
			parts := strings.Fields(line)
			//ygcTimes
			ygcTimes, err := strconv.ParseFloat(parts[12], 64)
			if err != nil {
				log.Fatal(err)
			}

			e.ygcTimes.Add(ygcTimes - e.lastYgcTimes)
			e.ygcTimes.Collect(ch)
			e.lastYgcTimes = ygcTimes
			ygcSec, err := strconv.ParseFloat(parts[13], 64)
			if err != nil {
				log.Fatal(err)
			}
			e.ygcSec.Set(ygcSec)
			e.ygcSec.Collect(ch)

			// fgcTimes
			fgcTimes, err := strconv.ParseFloat(parts[14], 64)
			if err != nil {
				log.Fatal(err)
			}

			e.fgcTimes.Add(fgcTimes - e.lastFgcTimes)
			e.fgcTimes.Collect(ch)
			e.lastFgcTimes = fgcTimes
			fgcSec, err := strconv.ParseFloat(parts[15], 64)
			if err != nil {
				log.Fatal(err)
			}
			e.fgcSec.Set(fgcSec)
			e.fgcSec.Collect(ch)

			//gcTime
			gcSec, err := strconv.ParseFloat(parts[16], 64)
			if err != nil {
				log.Fatal(err)
			}
			e.gcSec.Set(gcSec)
			e.gcSec.Collect(ch)
		}
	}
}

func main() {
	flag.Parse()

	exporter := NewExporter(*jstatPath, *targetPid)
	prometheus.MustRegister(exporter)

	log.Printf("Starting Server: %s", *listenAddress)
	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		<head><title>jstat Exporter</title></head>
		<body>
		<h1>jstat Exporter</h1>
		<p><a href="` + *metricsPath + `">Metrics</a></p>
		</body>
		</html>`))
	})
	err := http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		log.Fatal(err)
	}

}
