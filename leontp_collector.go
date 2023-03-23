package main

import (
	"sync"
	"strconv"
	"net"
	"encoding/binary"


	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

const prefix = "leontp_"

var (
	upDesc   *prometheus.Desc
	satellitesDesc *prometheus.Desc
	uptimeDesc *prometheus.Desc
	lockTimeDesc *prometheus.Desc
	ntpRequestsDesc *prometheus.Desc
	ntpTimeDesc *prometheus.Desc

	statRequest = []byte{0x27, 0x00, 0x10, 0x01, 0x00, 0x00, 0x00, 0x00}
)


func init() {
	upDesc = prometheus.NewDesc(prefix+"up", "Scrape was successful", []string{"host"}, nil)
	satellitesDesc = prometheus.NewDesc(prefix+"satellites_count", "Active satellites", []string{"host", "serial"}, nil)
	uptimeDesc = prometheus.NewDesc(prefix+"uptime_seconds", "Uptime", []string{"host", "serial"}, nil)
	lockTimeDesc = prometheus.NewDesc(prefix+"lock_time_seconds", "GPS lock time", []string{"host", "serial"}, nil)
	ntpRequestsDesc = prometheus.NewDesc(prefix+"ntp_requests_count", "NTP requests served", []string{"host", "serial"}, nil)
	ntpTimeDesc = prometheus.NewDesc(prefix+"ntp_time", "NTP time", []string{"host", "serial"}, nil)
}

type leontpCollector struct {
}

func (c leontpCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- upDesc
	ch <- satellitesDesc
	ch <- uptimeDesc
	ch <- lockTimeDesc
	ch <- ntpRequestsDesc
	ch <- ntpTimeDesc
}

func (c *leontpCollector) collectForNode(host string, ch chan<- prometheus.Metric, wg *sync.WaitGroup) {
	defer wg.Done()

	l := []string{host}

	udpServer, err := net.ResolveUDPAddr("udp", host+":123")
	if err != nil {
		log.Errorln("ResolveUDPAddr failed: " + err.Error())
		ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 0, l...)
		return
	}

	conn, err := net.DialUDP("udp", nil, udpServer)
	if err != nil {
		log.Errorln("Listen failed: " + err.Error())
		ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 0, l...)
		return
	}
	defer conn.Close()

	_, err = conn.Write(statRequest)
	if err != nil {
		log.Errorln("Write data failed: " + err.Error())
		ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 0, l...)
		return
	}

	received := make([]byte, 1024)
	_, err = conn.Read(received)
	if err != nil {
		log.Errorln("Read data failed: " + err.Error())
		ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 0, l...)
		return
	}

	refTs0 := float64(binary.LittleEndian.Uint32(received[16:20])) / 4294967296.0 // fractional part of the NTP timestamp
	refTs1 := binary.LittleEndian.Uint32(received[20:24]) // full seconds of NTP timestamp
	uptime := binary.LittleEndian.Uint32(received[24:28])
	ntpServed := binary.LittleEndian.Uint32(received[28:32])
	//cmdServed := binary.LittleEndian.Uint32(received[32:36])
	lockTime := binary.LittleEndian.Uint32(received[36:40])
	// flags := uint8(received[40])
	satellites := uint8(received[41])
	serial := binary.LittleEndian.Uint16(received[42:44])
	// firmwareVersion := binary.LittleEndian.Uint32(received[44:48])

	ntptime := refTs0 + float64(refTs1)

	ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 1, l...)

	l = append(l, strconv.FormatUint(uint64(serial), 10))
	ch <- prometheus.MustNewConstMetric(satellitesDesc, prometheus.GaugeValue, float64(satellites), l...)
	ch <- prometheus.MustNewConstMetric(uptimeDesc, prometheus.GaugeValue, float64(uptime), l...)
	ch <- prometheus.MustNewConstMetric(lockTimeDesc, prometheus.GaugeValue, float64(lockTime), l...)
	ch <- prometheus.MustNewConstMetric(ntpRequestsDesc, prometheus.GaugeValue, float64(ntpServed), l...)
	ch <- prometheus.MustNewConstMetric(ntpTimeDesc, prometheus.GaugeValue, ntptime, l...)
}

func (c leontpCollector) Collect(ch chan<- prometheus.Metric) {
	nodes := config.Nodes
	wg := &sync.WaitGroup{}

	wg.Add(len(nodes))
	for _, node := range nodes {
		go c.collectForNode(node, ch, wg)
	}

	wg.Wait()
}
