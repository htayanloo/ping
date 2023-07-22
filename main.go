package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/go-ping/ping"
	"github.com/olekukonko/tablewriter"
)

type HostResult struct {
	Host       string
	Response   string
	History    []float64
	AvgLatency []int
	PacketLoss int
	HostName   string
}

func pingHost(host string, results chan<- HostResult) {
	history := make([]float64, 0, 10)
	avgLatencyHistory := make([]int, 0, 10)
	hostnames, _ := net.LookupAddr(host)
	hostname := host
	if len(hostnames) > 0 {
		hostname = hostnames[0]
	}
	for {
		result, response, ms := pingAndGetResult(host)
		history = append(history, ms)
		if len(history) > 10 {
			history = history[1:]
		}
		avgLatency := calculateAverageLatency(history)
		avgLatencyHistory = append(avgLatencyHistory, avgLatency)
		if len(avgLatencyHistory) > 10 {
			avgLatencyHistory = avgLatencyHistory[1:]
		}
		results <- HostResult{Host: result, Response: response, History: history, AvgLatency: avgLatencyHistory, PacketLoss: int(ms), HostName: hostname}
		time.Sleep(time.Second)
	}
}

func pingAndGetResult(host string) (string, string, float64) {
	if runtime.GOOS == "windows" {
		out, _ := exec.Command("ping", host, "-n", "1", "-w", "1000").Output()
		if strings.Contains(string(out), "Request timed out.") {
			return host, "unavailable", 10000
		}
		latency := strings.Split(strings.Split(string(out), "Average = ")[1], "ms")[0]
		latencyMs, _ := strconv.Atoi(latency)
		return host, fmt.Sprintf("%d ms", latencyMs), float64(latencyMs)
	} else {
		pinger, err := ping.NewPinger(host)
		if err != nil {
			return host, "unavailable", 10000 // large value for 'unavailable'
		}

		pinger.Count = 1
		pinger.Timeout = time.Second * 1

		err = pinger.Run()
		if err != nil {
			return host, "unavailable", 10000 // large value for 'unavailable'
		}

		stats := pinger.Statistics()
		if stats.PacketLoss == 100 {
			return host, "unavailable", 10000 // large value for 'unavailable'
		}

		latencyMs := stats.AvgRtt.Milliseconds()

		return host, fmt.Sprintf("%d ms", latencyMs), float64(latencyMs)
	}
}

func getColoredSparkline(values []float64) string {
	sparkline := ""
	for _, value := range values {
		if value >= 100 {
			sparkline += color.RedString("▄")
		} else if value >= 50 {
			sparkline += color.YellowString("▄")
		} else {
			sparkline += color.GreenString("▄")
		}
	}
	return sparkline
}

func calculateAverageLatency(latencies []float64) int {
	sum := 0.0
	for _, latency := range latencies {
		sum += latency
	}
	return int(sum) / len(latencies)
}

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("Usage: go run main.go <ip_address_1> <ip_address_2> ...")
		return
	}

	hosts := args[1:]
	results := make(chan HostResult)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"#", "Host", "Hostname", "Ping Response", "Average Latency", "Latency Change", "Packet Loss %", "Last 10 Responses (Sparkline)"})

	hostResults := make(map[string]HostResult)

	for _, host := range hosts {
		go pingHost(host, results)
	}

	go func() {
		for result := range results {
			hostResults[result.Host] = result
			table.ClearRows()
			for i, host := range hosts {
				if res, ok := hostResults[host]; ok {
					sparkline := getColoredSparkline(res.History)
					rowNumber := fmt.Sprintf("%d", i+1)
					avgLatencyChange := ""
					if len(res.AvgLatency) > 1 {
						diff := res.AvgLatency[len(res.AvgLatency)-1] - res.AvgLatency[len(res.AvgLatency)-2]
						if diff > 1 {
							avgLatencyChange = color.RedString("↑")
						} else if diff < -1 {
							avgLatencyChange = color.GreenString("↓")
						}
					}
					if res.Response == "unavailable" {
						table.Rich([]string{rowNumber, res.Host, res.HostName, res.Response, fmt.Sprintf("%d ms", res.AvgLatency[len(res.AvgLatency)-1]), avgLatencyChange, fmt.Sprintf("%d %%", res.PacketLoss), sparkline}, []tablewriter.Colors{{tablewriter.FgRedColor}, {tablewriter.FgRedColor}, {tablewriter.FgRedColor}, {tablewriter.FgRedColor}, {tablewriter.FgRedColor}, {tablewriter.FgRedColor}, {tablewriter.FgRedColor}})
					} else {
						table.Append([]string{rowNumber, res.Host, res.HostName, res.Response, fmt.Sprintf("%d ms", res.AvgLatency[len(res.AvgLatency)-1]), avgLatencyChange, fmt.Sprintf("%d %%", res.PacketLoss), sparkline})
					}
				}
			}
			fmt.Print("\033[H\033[2J")
			table.Render()
		}
	}()

	select {}
}
