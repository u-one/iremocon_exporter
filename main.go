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

	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer conn.Close()

	startExporter(context.Background())

	for {
		func() {
			res, err := iremocon.Se(conn)
			if err != nil {
				fmt.Printf("Error: [%s]\n", err)
				return
			}

			res = strings.TrimSpace(res)
			fmt.Println(res)

			strs := strings.Split(res, ";")
			if len(strs) < 5 {
				fmt.Printf("Error: invalid response [%s]\n", res)
				return
			}

			brightness, err := strconv.ParseFloat(strs[2], 64)
			if err != nil {
				fmt.Printf("Error parsing brightness: [%s]\n", err)
			} else {
				metBright.Set(brightness)
			}

			humidity, err := strconv.ParseFloat(strs[3], 64)
			if err != nil {
				fmt.Printf("Error parsing humidity: [%s]\n", err)
			} else {
				metHum.Set(humidity)
			}

			temperature, err := strconv.ParseFloat(strs[4], 64)
			if err != nil {
				fmt.Printf("Error parsing temperature: [%s]\n", err)
			} else {
				metTemp.Set(temperature)
			}
		}()
		time.Sleep(time.Second * 60)
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
