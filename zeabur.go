package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// GraphQL query structure
type GraphQLRequest struct {
	Query string `json:"query"`
}

// GraphQL response structures
type GraphQLResponse struct {
	Data struct {
		Me struct {
			ID       string `json:"_id"`
			Name     string `json:"name"`
			Username string `json:"username"`
			Email    string `json:"email"`
		} `json:"me"`
		Servers []struct {
			ID   string `json:"_id"`
			Name string `json:"name"`
			IP   string `json:"ip"`
		} `json:"servers"`
		Projects struct {
			Edges []struct {
				Node struct {
					ID      string `json:"_id"`
					Name    string `json:"name"`
					Environments []struct {
						ID   string `json:"_id"`
						Name string `json:"name"`
					} `json:"environments"`
					Services []struct {
						ID      string `json:"_id"`
						Name    string `json:"name"`
						DNSName string `json:"dnsName"`
					} `json:"services"`
					Region struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"region"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"projects"`
	} `json:"data"`
}

func buildZeaburDnsMap(apiKey string, serverId string) (map[string]string, error) {
	// Define the GraphQL query
	query := `
	query {
		me {
			_id
			name
			username
			email
		}
		servers {
			_id
			name
			ip
		}
		projects(limit: 1024) {
			edges {
				node {
					_id
					name
					environments {
						_id
						name
					}
					services {
						_id
						name
						dnsName
					}
					region {
						id
						name
					}
				}
			}
		}
	}
	`

	// Create the request body
	requestBody, err := json.Marshal(GraphQLRequest{Query: query})
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", "https://api.zeabur.com/graphql", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	// Parse the response
	var graphQLResp GraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&graphQLResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	// Check if the serverId exists in the servers list
	serverExists := false
	for _, server := range graphQLResp.Data.Servers {
		if "server-"+server.ID == serverId {
			serverExists = true
			break
		}
	}
	if !serverExists {
		return nil, fmt.Errorf("server with ID %s does not exist", serverId)
	}

	// Build the DNS map
	dnsMap := make(map[string]string)
	for _, projectEdge := range graphQLResp.Data.Projects.Edges {
		project := projectEdge.Node

		// Filter projects by serverId
		if project.Region.ID != serverId {
			continue
		}

		// Get the project name in lowercase
		projectName := strings.ToLower(project.Name)

		// Get the environment ID (assuming there's at least one environment)
		if len(project.Environments) == 0 {
			continue
		}
		environmentID := project.Environments[0].ID

		// Process each service
		for _, service := range project.Services {
			// Create the key: dnsName.projectName (lowercase)
			key := strings.ToLower(service.DNSName) + "." + projectName

			// Create the value: service-serviceId.environment-environmentId
			value := "service-" + service.ID + ".environment-" + environmentID

			// Add to the map
			dnsMap[key] = value
		}
	}

	return dnsMap, nil
}
