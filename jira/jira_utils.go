package jira

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Issue struct {
	Title       string `json:"summary"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Assignee    string `json:"assignee"`
}

type JiraUtils struct {
	jiraHost string
	apiToken string
}

func New(jiraHost string, apiToken string) *JiraUtils {
	return &JiraUtils{
		jiraHost,
		apiToken,
	}
}

func (ut *JiraUtils) GetIssue(id string) (*Issue, error) {
	// Set up the HTTP client and request
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s/rest/api/2/issue/%s", ut.jiraHost, id), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	// Add your Jira API token here
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", ut.apiToken))

	// Make the request and get the response
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON response into the Issue struct
	var issue Issue
	err = json.Unmarshal(body, &issue)
	if err != nil {
		return nil, err
	}

	return &issue, nil
}
