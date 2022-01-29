package main

import (
	"github.com/ryanwholey/terraform-cloud-update-workspace-variables/internal/action"
	"github.com/sethvargo/go-githubactions"
)

func main() {
	if err := action.Run(action.Inputs{
		Organization: githubactions.GetInput("organization"),
		Token:        githubactions.GetInput("token"),
		Address:      githubactions.GetInput("address"),
		SecretsPath:  githubactions.GetInput("secrets_path"),
	}); err != nil {
		githubactions.Fatalf("Error: %s", err)
	}
}
