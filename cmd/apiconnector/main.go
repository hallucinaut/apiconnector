package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
)

type ConnectionTest struct {
	Service     string
	URL         string
	Status      string
	Latency     time.Duration
	Headers     map[string]string
	Error       string
}

func main() {
	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived shutdown signal, cancelling...")
		cancel()
	}()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	fmt.Println(color.CyanString("\n=== API CONNECTIVITY TEST ===\n"))

	var tests []ConnectionTest
	for _, arg := range os.Args[1:] {
		test := parseTestConfig(arg)
		tests = append(tests, test)
	}

	// Run tests with context
	if err := runConnectionTestsWithContext(ctx, tests); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(color.CyanString("apiconnector - API Connectivity Tester"))
	fmt.Println()
	fmt.Println("Usage: apiconnector <service1> <service2> ...")
	fmt.Println("Format: name=http://url[:port]")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  apiconnector api=http://localhost:8080/health")
	fmt.Println("  db=postgres://localhost:5432")
}

func parseTestConfig(config string) ConnectionTest {
	test := ConnectionTest{}
	parts := strings.SplitN(config, "=", 2)
	if len(parts) == 2 {
		test.Service = parts[0]
		test.URL = parts[1]
	}
	return test
}

func runConnectionTests(tests []ConnectionTest) error {
	return runConnectionTestsWithContext(context.Background(), tests)
}

func runConnectionTestsWithContext(ctx context.Context, tests []ConnectionTest) error {
	var success, failure int

	for i := range tests {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled")
		default:
		}

		test := &tests[i]
		test.Status, test.Latency, test.Error = testConnect(ctx, test.URL)

		if test.Error == "" {
			success++
			fmt.Printf("%-20s %s (%s)\n", test.Service, color.GreenString("OK"), formatDuration(test.Latency))
		} else {
			failure++
			fmt.Printf("%-20s %s (%s)\n", test.Service, color.RedString("FAIL"), test.Error)
		}
	}

	fmt.Println()
	fmt.Printf("Summary: %d OK, %d FAIL\n", success, failure)

	if failure > 0 {
		return fmt.Errorf("%d connection failures", failure)
	}

	return nil
}

func testConnect(ctx context.Context, url string) (string, time.Duration, string) {
	start := time.Now()

	// Check context cancellation
	select {
	case <-ctx.Done():
		return "ERROR", 0, "context cancelled"
	default:
	}

	// Parse URL
	parsedURL := parseURL(url)
	if parsedURL == "" {
		return "ERROR", 0, "Invalid URL"
	}

	// Check port connectivity
	port := getPort(url)
	if port != "" {
		conn, err := net.DialTimeout("tcp", parsedURL+":"+port, 5*time.Second)
		if err != nil {
			return "FAIL", 0, fmt.Sprintf("Port %s unreachable: %v", port, err)
		}
		conn.Close()
	}

	// Check HTTP endpoint if it's an HTTP URL
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		client := &http.Client{
			Timeout: 5 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		// Create request with context
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return "ERROR", 0, fmt.Sprintf("Request creation error: %v", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return "FAIL", 0, fmt.Sprintf("HTTP error: %v", err)
		}
		defer resp.Body.Close()

		latency := time.Since(start)
		status := "OK"
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			status = "OK"
		} else {
			status = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}

		return status, latency, ""
	}

	return "OK", time.Since(start), ""
}

func parseURL(url string) string {
	// Remove protocol
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	
	// Get hostname
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return url
}

func getPort(url string) string {
	// Extract port from URL
	parts := strings.Split(url, ":")
	if len(parts) >= 2 {
		for i, part := range parts {
			if i > 0 && i < len(parts)-1 {
				// Check if this looks like a port
				if part != "" && part != "http" && part != "https" {
					if _, err := strconv.Atoi(part); err == nil {
						return part
					}
				}
			}
		}
	}
	return ""
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	return fmt.Sprintf("%dms", d.Milliseconds())
}