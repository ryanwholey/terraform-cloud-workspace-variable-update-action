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
		Variables:    githubactions.GetInput("variables"),
		WorkspaceTag: githubactions.GetInput("workspace_tag"),
	}); err != nil {
		githubactions.Fatalf("Error: %s", err)
	}
}
