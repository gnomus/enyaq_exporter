package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	promVersion "github.com/prometheus/common/version"
	"github.com/prometheus/client_golang/prometheus/collectors/version"

	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/vehicle/skoda"
	"github.com/evcc-io/evcc/vehicle/skoda/connect"
	"github.com/evcc-io/evcc/vehicle/vag/service"
)

func init() {
	promVersion.Version = "0.1.2"
	prometheus.MustRegister(version.NewCollector("enyaq_exporter"))
}

func main() {
	var (
		listenAddr   = flag.String("web.listen-address", ":9333", "The address to listen on for HTTP requests.")
		username     = flag.String("username", "user@example.com", "Login email for Skoda Connect")
		password     = flag.String("password", "secret1234", "Password for Skoda Connect account")
		vin          = flag.String("vin", "TM...", "VIN of vehicle to check")
		pollInterval = flag.Int("poll-interval", 60, "Interval in seconds between polls.")
		showVersion  = flag.Bool("version", false, "Print version information and exit.")
	)

	flag.Parse()

	if *showVersion {
		fmt.Printf("%s\n", promVersion.Print("enyaq_exporter"))
		os.Exit(0)
	}

	var evRange = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ev_range",
		Help: "Electric vehicle range",
	})
	var evSoc = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ev_soc",
		Help: "Electric vehicle state of charge",
	})
	var evStatus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ev_status",
		Help: "Electric vehicle status",
	})
	var evFinishTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ev_finish_time",
		Help: "Electric charging finish time",
	})
	var evOdometer = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ev_odometer",
		Help: "Electric odometer",
	})

	// Register the summary and the histogram with Prometheus's default registry
	prometheus.MustRegister(evRange)
	prometheus.MustRegister(evSoc)
	prometheus.MustRegister(evStatus)
	prometheus.MustRegister(evFinishTime)
	prometheus.MustRegister(evOdometer)

	// Add Go module build info
	prometheus.MustRegister(collectors.NewBuildInfoCollector())

	var err error
	logHandler := util.NewLogger("enyaq").Redact(*username, *password, *vin)

	// Poll inverter values
	go func() {
		for {
			// use Connect credentials to build provider
			var provider *skoda.Provider
			if err == nil {
				ts, err := service.TokenRefreshServiceTokenSource(logHandler, skoda.TRSParams, connect.AuthParams, *username, *password)
				if err != nil {
					log.Print("TokenRefresh error: ", err)
					time.Sleep(time.Duration(*pollInterval) * time.Second)

					continue
				}

				api := skoda.NewAPI(logHandler, ts)
				api.Client.Timeout = time.Second * 30

				provider = skoda.NewProvider(api, *vin, time.Second*30)
			}

			rangeKm, err := provider.Range()
			if err != nil {
				log.Print("Range Error: ", err)
			} else {
				evRange.Set(float64(rangeKm))
			}

			soc, err := provider.Soc()
			if err != nil {
				log.Print("SoC error: ", err)
			} else {
				evSoc.Set(soc)
			}

			statusString, err := provider.Status()
			if err != nil {
				log.Print("Status error: ", err)
			} else {
				switch statusString.String() {
				case "":
					evStatus.Set(0)
				case "A":
					evStatus.Set(1)
				case "B":
					evStatus.Set(2)
				case "C":
					evStatus.Set(3)
				case "D":
					evStatus.Set(4)
				case "E":
					evStatus.Set(5)
				case "F":
					evStatus.Set(6)
				default:
					log.Print("Unknown status: ", statusString)
				}
			}

			finishTime, err := provider.FinishTime()
			if err != nil {
				log.Print("Finish time error: ", err)
			} else {
				evFinishTime.Set(float64(finishTime.Unix()))
			}

			odometer, err := provider.Odometer()
			if err != nil {
				log.Print("Odometer error: ", err)
			} else {
				evOdometer.Set(odometer)
			}

			time.Sleep(time.Duration(*pollInterval) * time.Second)
		}
	}()

	// Expose the registered metrics via HTTP
	http.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{},
	))
	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
