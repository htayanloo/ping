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
	Host              string
	Response          string
	History           []float64
	AvgLatency        []int
	LastLatencyMs     int // Renamed from PacketLoss
	HostName          string
	PacketLossPercent float64       // New field
	Jitter            time.Duration // New field
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
		// Update the call to pingAndGetResult to receive new values
		hostString, response, latencyMs, packetLossPercent, jitter := pingAndGetResult(host)
		history = append(history, latencyMs) // latencyMs is float64
		if len(history) > 10 {
			history = history[1:]
		}
		avgLatency := calculateAverageLatency(history)
		avgLatencyHistory = append(avgLatencyHistory, avgLatency)
		if len(avgLatencyHistory) > 10 {
			avgLatencyHistory = avgLatencyHistory[1:]
		}
		// Populate HostResult with new fields
		results <- HostResult{
			Host:              hostString, // This is the original host string
			Response:          response,
			History:           history,
			AvgLatency:        avgLatencyHistory,
			LastLatencyMs:     int(latencyMs), // Populate LastLatencyMs with current ping's latency
			HostName:          hostname,
			PacketLossPercent: packetLossPercent,
			Jitter:            jitter,
		}
		time.Sleep(time.Second)
	}
}

// Updated function signature
func pingAndGetResult(host string) (string, string, float64, float64, time.Duration) {
	if runtime.GOOS == "windows" {
		out, _ := exec.Command("ping", host, "-n", "1", "-w", "1000").Output()
		if strings.Contains(string(out), "Request timed out.") {
			return host, "unavailable", 10000, 100.0, 0 // host, response, latencyMs, packetLossPercent, jitter
		}
		latencyStr := strings.Split(strings.Split(string(out), "Average = ")[1], "ms")[0]
		latencyMs, _ := strconv.Atoi(latencyStr)
		return host, fmt.Sprintf("%d ms", latencyMs), float64(latencyMs), 0.0, 0 // host, response, latencyMs, packetLossPercent, jitter
	} else {
		pinger, err := ping.NewPinger(host)
		if err != nil {
			return host, "unavailable", 10000, 100.0, 0 // host, response, latencyMs, packetLossPercent, jitter
		}

		pinger.Count = 1
		pinger.Timeout = time.Second * 1
		pinger.SetPrivileged(true) // This might be needed on some systems

		err = pinger.Run() // Blocks until finished
		if err != nil {
			return host, "unavailable", 10000, 100.0, 0 // host, response, latencyMs, packetLossPercent, jitter
		}

		stats := pinger.Statistics() // get send/receive/rtt stats

		// Use stats.PacketLoss directly for packetLossPercent.
		// If stats.PacketsRecv == 0, it often means 100% loss or host is down.
		if stats.PacketsRecv == 0 { // More robust check for unavailability
			return host, "unavailable", 10000, 100.0, 0 // host, response, latencyMs, packetLossPercent, jitter
		}

		latencyMs := float64(stats.AvgRtt.Milliseconds()) // AvgRtt is time.Duration
		jitter := stats.StdDevRtt

		// The response string should reflect the latency in ms, typically as an integer.
		return host, fmt.Sprintf("%d ms", int64(latencyMs)), latencyMs, stats.PacketLoss, jitter // host, response, latencyMs, packetLossPercent, jitter
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

	table := tablewriter.NewTable(os.Stdout) // Changed NewWriter to NewTable for v1.0.x style
	// Updated header definition
	header := []string{"#", "Host", "Hostname", "Last Ping", "Avg Latency", "Trend", "Pkt Loss %", "Jitter", "Latency Sparkline"}
	// table.Header(header) // Header will be set inside the loop after Reset

	hostResults := make(map[string]HostResult)

	for _, host := range hosts {
		go pingHost(host, results)
	}

	go func() {
		for result := range results {
			hostResults[result.Host] = result
			table.Reset()      // Changed ClearRows to Reset
			table.Header(header) // Re-set header after Reset
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

					// Updated rowData population to include new fields and use correct existing fields
					rowData := []string{
						rowNumber,
						res.Host,
						res.HostName,
						res.Response, // This is the last ping status/latency string
						fmt.Sprintf("%d ms", res.AvgLatency[len(res.AvgLatency)-1]), // Average Latency
						avgLatencyChange, // Trend
						fmt.Sprintf("%.2f %%", res.PacketLossPercent),             // Packet Loss Percentage
						fmt.Sprintf("%s", res.Jitter),                           // Jitter
						sparkline, // Latency Sparkline
					}
					if res.Response == "unavailable" {
						coloredRowData := make([]string, len(rowData))
						for i, cell := range rowData {
							// avgLatencyChange and sparkline are already potentially colored by getColoredSparkline or the logic above.
							// Avoid double-coloring them.
							if cell == avgLatencyChange || cell == sparkline {
								coloredRowData[i] = cell
							} else {
								coloredRowData[i] = color.RedString(cell)
							}
						}
						table.Append(coloredRowData)
					} else {
						table.Append(rowData)
					}
				}
			}
			fmt.Print("\033[H\033[2J")
			table.Render()
		}
	}()

	select {}
}
