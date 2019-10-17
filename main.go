package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/u-one/go-iremocon/iremocon"
)

var (
	metTemp = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "myroom_temperature",
			Help: "temperature",
		},
	)
	metHum = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "myroom_humidity",
			Help: "humidity",
		},
	)
	metBright = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "myroom_lux",
			Help: "brightness",
		},
	)
)

func init() {
	prometheus.MustRegister(metTemp)
	prometheus.MustRegister(metHum)
	prometheus.MustRegister(metBright)
}

var (
	irHost       = flag.String("ir_host", "", "iRemocon host")
	irPort       = flag.String("ir_port", "51013", "iRemocon port (default: 51013)")
	exporterPort = flag.String("ex_port", "8080", "port for prometheus exporter")
)

func main() {
	flag.Parse()
	address := *irHost + ":" + *irPort
	fmt.Println("iRemocon address:", address)

	startExporter(context.Background())

connect:
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer conn.Close()

	for {
		err := func() error {
			res, err := iremocon.Se(conn)
			if err != nil {
				return fmt.Errorf("Se failed: [%s]\n", err)
			}

			res = strings.TrimSpace(res)
			fmt.Println(res)

			strs := strings.Split(res, ";")
			if len(strs) < 5 {
				return fmt.Errorf("Error: invalid response [%s]\n", res)
			}

			brightness, err := strconv.ParseFloat(strs[2], 64)
			if err != nil {
				return fmt.Errorf("Error parsing brightness: [%s]\n", err)
			}
			metBright.Set(brightness)

			humidity, err := strconv.ParseFloat(strs[3], 64)
			if err != nil {
				return fmt.Errorf("Error parsing humidity: [%s]\n", err)
			}
			metHum.Set(humidity)

			temperature, err := strconv.ParseFloat(strs[4], 64)
			if err != nil {
				return fmt.Errorf("Error parsing temperature: [%s]\n", err)
			}
			metTemp.Set(temperature)

			return nil
		}()
		time.Sleep(time.Second * 60)

		if err != nil {
			conn.Close()
			goto connect
		}
	}
}

func startExporter(ctx context.Context) {
	go func() {
		fmt.Println("startExporter port: ", *exporterPort)
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(":"+*exporterPort, nil)
		fmt.Println("exporter finished: ", err)
	}()
}
