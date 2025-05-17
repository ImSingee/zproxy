package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wzshiming/socks5"
)

type Config struct {
	Listen string

	InDomainSuffix string
	ClusterDomain  string

	AuthUsername string
	AuthPassword string
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

	// Check if it's an IP address
	if net.ParseIP(host) != nil {
		return nil, fmt.Errorf("IP addresses are not allowed: %s", host)
	}

	// Check if the domain ends with the specified suffix
	suffix := "." + d.inDomainSuffix
	if !strings.HasSuffix(host, suffix) {
		return nil, fmt.Errorf("domain must end with %s: %s", suffix, host)
	}

	newHost := host[:len(host)-len(suffix)] + "." + d.clusterDomain
	newAddr := net.JoinHostPort(newHost, port)

	return net.Dial(network, newAddr)
}

// DialContext is the context-aware version of Dial
func (d *CustomDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return d.Dial(network, address)
}

func main() {
	config := &Config{
		Listen:         ":1080",
		InDomainSuffix: "cluster.local",
		ClusterDomain:  "cluster.local",
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

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runProxy(config *Config) {
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
