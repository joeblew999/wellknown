package pdfform

import (
	"fmt"
	"net"
)

// NetworkInfo holds information about network connectivity
type NetworkInfo struct {
	LocalIPs []string // Local network IP addresses
	Port     string   // Port number
}

// GetLocalIPs returns all non-loopback IPv4 addresses
func GetLocalIPs() ([]string, error) {
	var ips []string

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		// Check if it's an IP address (not a network)
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			// Only get IPv4 addresses
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}

	return ips, nil
}

// GetServerURLs returns all URLs where the server can be accessed
func GetServerURLs(port string, https bool) ([]string, error) {
	var urls []string
	protocol := "http"
	if https {
		protocol = "https"
	}

	// Add localhost
	urls = append(urls, fmt.Sprintf("%s://localhost:%s", protocol, port))

	// Add 127.0.0.1
	urls = append(urls, fmt.Sprintf("%s://127.0.0.1:%s", protocol, port))

	// Add LAN IPs
	ips, err := GetLocalIPs()
	if err != nil {
		return urls, err // Return localhost URLs even if we can't get LAN IPs
	}

	for _, ip := range ips {
		urls = append(urls, fmt.Sprintf("%s://%s:%s", protocol, ip, port))
	}

	return urls, nil
}

// PrintServerInfo displays all URLs where the server can be accessed
func PrintServerInfo(port string, https bool) {
	urls, err := GetServerURLs(port, https)
	if err != nil {
		fmt.Printf("⚠️  Warning: Could not detect all network interfaces: %v\n", err)
	}

	protocol := "HTTP"
	if https {
		protocol = "HTTPS"
	}

	fmt.Printf("\n🌐 %s Server started on port %s\n\n", protocol, port)
	fmt.Println("📱 Access from your devices:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for i, url := range urls {
		var label string
		switch {
		case i == 0:
			label = "🖥️  Local (this computer)"
		case i == 1:
			label = "🔗 Local (IP address)"
		default:
			label = "📡 LAN (from other devices)"
		}
		fmt.Printf("  %s\n     %s\n\n", label, url)
	}

	if https {
		fmt.Println("🔒 HTTPS is enabled - certificates in .data/certs/")
		fmt.Println("")
		fmt.Println("📱 iOS Setup:")
		fmt.Println("   1. Open Safari on your iPhone/iPad")
		fmt.Println("   2. Visit any URL above")
		fmt.Println("   3. Tap 'Show Details' > 'visit this website'")
		fmt.Println("   4. Done! Certificate is now trusted")
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("\nPress Ctrl+C to stop the server")
	fmt.Println()
}
