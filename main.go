package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/wzshiming/socks5"
)

var zeaburDnsStore = NewZeaburDnsStore()

type Config struct {
	Listen string

	InDomainSuffix string
	ClusterDomain  string

	AuthUsername string
	AuthPassword string

	// Zeabur related configuration
	ZeaburAPIKey         string
	ZeaburServerID       string
	ZeaburUpdateInterval time.Duration
}

type CustomDialer struct {
	inDomainSuffix string
	clusterDomain  string
}

// Dial connects to the target address, potentially modifying the domain name
func (d *CustomDialer) Dial(network, address string) (net.Conn, error) {
	// Parse the address to get host and port
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	host = strings.ToLower(host)

	// Check if the domain ends with the specified suffix
	suffix := "." + d.inDomainSuffix
	if !strings.HasSuffix(host, suffix) {
		return nil, fmt.Errorf("domain must end with %s: %s", suffix, host)
	}

	var newHost string

	// Check if it's a Zeabur domain (.zeabur.{inDomainSuffix})
	zeaburSuffix := ".zeabur" + suffix
	if strings.HasSuffix(host, zeaburSuffix) {
		// Extract the service name and project name
		key := host[:len(host)-len(zeaburSuffix)]

		// Get the value from the DNS store
		value, exists := zeaburDnsStore.Get(key)

		if !exists {
			return nil, fmt.Errorf("zeabur service not found: %s", key)
		}

		// Construct the new host: {value}.svc.{clusterDomain}
		newHost = value + ".svc." + d.clusterDomain
	} else {
		// Regular domain transformation
		newHost = host[:len(host)-len(suffix)] + "." + d.clusterDomain
	}

	newAddr := net.JoinHostPort(newHost, port)
	return net.Dial(network, newAddr)
}

// DialContext is the context-aware version of Dial
func (d *CustomDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return d.Dial(network, address)
}

func main() {
	config := &Config{
		Listen:               ":1080",
		InDomainSuffix:       "cluster.local",
		ClusterDomain:        "cluster.local",
		ZeaburUpdateInterval: 5 * time.Minute, // Default update interval: 5 minutes
	}

	// Read environment variables
	if port := os.Getenv("PORT"); port != "" {
		config.Listen = fmt.Sprintf(":%s", port)
	}
	if domainSuffix := os.Getenv("IN_DOMAIN_SUFFIX"); domainSuffix != "" {
		config.InDomainSuffix = domainSuffix
	}
	if clusterDomain := os.Getenv("CLUSTER_DOMAIN"); clusterDomain != "" {
		config.ClusterDomain = clusterDomain
	}
	config.AuthUsername = os.Getenv("AUTH_USERNAME")
	config.AuthPassword = os.Getenv("AUTH_PASSWORD")

	// Read Zeabur-related environment variables
	config.ZeaburAPIKey = os.Getenv("ZEABUR_API_KEY")
	config.ZeaburServerID = os.Getenv("ZEABUR_SERVER_ID")
	if updateInterval := os.Getenv("ZEABUR_UPDATE_INTERVAL"); updateInterval != "" {
		duration, err := time.ParseDuration(updateInterval)
		if err != nil {
			fmt.Printf("Warning: Invalid ZEABUR_UPDATE_INTERVAL format: %v. Using default: %v\n", err, config.ZeaburUpdateInterval)
		} else {
			config.ZeaburUpdateInterval = duration
		}
	}

	// Create root command
	rootCmd := &cobra.Command{
		Use:   "zproxy",
		Short: "A SOCKS5 proxy for accessing internal services deployed by Zeabur Dedicated Server",
		Run: func(cmd *cobra.Command, args []string) {
			runProxy(config)
		},
	}

	// Add command line flags
	flags := rootCmd.Flags()
	flags.StringVarP(&config.Listen, "listen", "l", config.Listen, "Proxy listening address")
	flags.StringVarP(&config.InDomainSuffix, "in-domain-suffix", "s", config.InDomainSuffix, "Domain suffix to replace")
	flags.StringVarP(&config.ClusterDomain, "cluster-domain", "c", config.ClusterDomain, "Cluster domain to use as replacement")
	flags.StringVarP(&config.AuthUsername, "username", "u", config.AuthUsername, "Authentication username")
	flags.StringVarP(&config.AuthPassword, "password", "p", config.AuthPassword, "Authentication password")

	// Add Zeabur-related flags
	flags.StringVar(&config.ZeaburAPIKey, "zeabur-api-key", config.ZeaburAPIKey, "Zeabur API key")
	flags.StringVar(&config.ZeaburServerID, "zeabur-server-id", config.ZeaburServerID, "Zeabur server ID")
	flags.DurationVar(&config.ZeaburUpdateInterval, "zeabur-update-interval", config.ZeaburUpdateInterval, "Interval for updating Zeabur DNS map")

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// updateZeaburDnsStore updates the global DNS store using the Zeabur API
func updateZeaburDnsStore(apiKey, serverId string) error {
	// Call the buildZeaburDnsMap function to get the DNS data
	dnsMap, err := buildZeaburDnsMap(apiKey, serverId)
	if err != nil {
		return fmt.Errorf("failed to build Zeabur DNS store: %w", err)
	}

	// Update the global DNS store
	zeaburDnsStore.Set(dnsMap)

	return nil
}

func runProxy(config *Config) {
	// Initialize the Zeabur DNS store if API key and server ID are provided
	if config.ZeaburAPIKey != "" && config.ZeaburServerID != "" {
		// Update the DNS store
		err := updateZeaburDnsStore(config.ZeaburAPIKey, config.ZeaburServerID)
		if err != nil {
			log.Fatalf("Failed to initialize Zeabur DNS store: %v", err)
		}

		log.Printf("Zeabur DNS store initialized with update interval: %v", config.ZeaburUpdateInterval)

		// Start a goroutine to periodically update the DNS store
		go func() {
			ticker := time.NewTicker(config.ZeaburUpdateInterval)
			defer ticker.Stop()

			for range ticker.C {
				err := updateZeaburDnsStore(config.ZeaburAPIKey, config.ZeaburServerID)
				if err != nil {
					log.Printf("Warning: Failed to update Zeabur DNS store: %v", err)
				}
			}
		}()
	} else {
		log.Println("Zeabur DNS store disabled: API key or server ID not provided")
	}

	// Create custom dialer
	dialer := &CustomDialer{
		inDomainSuffix: config.InDomainSuffix,
		clusterDomain:  config.ClusterDomain,
	}

	// Create a new SOCKS5 server
	server := socks5.NewServer()

	// Set the custom dialer for the server
	server.ProxyDial = dialer.DialContext

	// Set up authentication if credentials are provided
	authEnabled := config.AuthUsername != "" && config.AuthPassword != ""
	if authEnabled {
		// Create a user/pass authenticator
		auth := socks5.UserAuth(config.AuthUsername, config.AuthPassword)
		server.Authentication = auth

		fmt.Printf("Authentication enabled for user: %s\n", config.AuthUsername)
	} else {
		fmt.Println("Warning: No authentication credentials provided. Running without authentication.")
	}

	fmt.Printf("Starting SOCKS5 proxy on %s\n", config.Listen)
	fmt.Printf("Domain suffix: %s, Cluster domain: %s\n", config.InDomainSuffix, config.ClusterDomain)

	// Start the server
	err := server.ListenAndServe("tcp", config.Listen)
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
}
