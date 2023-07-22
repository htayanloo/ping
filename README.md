# Go Network Ping

This is a simple tool written in Go to ping a list of hosts and display the results in a table format in the terminal. This tool also gives the average latency, packet loss percentage, and a sparkline representation of the last ten responses.

This project was developed by [Hadi Tayanloo](https://www.linkedin.com/in/htayanloo).

## Features

- Ping multiple hosts
- Sparkline for the last ten ping responses
- Average latency calculation
- Packet loss percentage
- Dynamically updating table view in the terminal
- Latency trend indicator (Upwards red arrow if increase is more than 20ms, downwards green arrow if decrease is more than 20ms)
- Unavailable hosts are highlighted in red

## Usage

1. Install [Go](https://golang.org/doc/install) if you haven't already done so.

2. Clone this repository:
   ```
   git clone https://github.com/htayanloo/ping.git
   ```

3. Navigate into the project directory:
   ```
   cd ping
   ```

4. Run the program with a list of IP addresses as arguments:
   ```
   go run main.go 8.8.8.8 1.1.1.1
   ```
   Replace `8.8.8.8` and `1.1.1.1` with the IP addresses of the hosts you want to ping.

## Output

The program will display an updating table that includes:

- Row number
- IP address
- Hostname
- Current ping response
- Average latency
- Latency trend indicator
- Packet loss percentage
- Sparkline for the last ten ping responses

If a host is unavailable, its row will be displayed in red.

## Notes

This tool is meant for simple network diagnostics. It is not meant to replace more comprehensive network monitoring solutions.

## Contributions

Contributions, issues, and feature requests are welcome! Feel free to check the [issues page](https://github.com/your-repo/your-project/issues).

## Developer

[Hadi Tayanloo](https://www.linkedin.com/in/htayanloo)

Contact - Skype: htayanloo

## License

Distributed under the MIT License. See `LICENSE` for more information.