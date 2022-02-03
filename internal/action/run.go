package action

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/go-tfe"
)

type Inputs struct {
	Token        string
	Address      string
	Organization string
	Variables    string
	WorkspaceTag string
}

type Variable struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Category    string `json:"category"`
	Sensitive   bool   `json:"sensitive"`
	Description string `json:"description"`
}

func Run(inputs Inputs) error {
	ctx := context.Background()

	tfeClient, err := tfe.NewClient(&tfe.Config{
		Token:   inputs.Token,
		Address: inputs.Address,
	})
	if err != nil {
		return err
	}

	var variables []Variable

	if err := json.Unmarshal([]byte(inputs.Variables), &variables); err != nil {
		return err
	}

	// TODO: handle pagination
	workspaceList, err := tfeClient.Workspaces.List(ctx, inputs.Organization, tfe.WorkspaceListOptions{
		ListOptions: tfe.ListOptions{},
	})
	if err != nil {
		return err
	}

	fmt.Println(workspaceList.Items, inputs.WorkspaceTag, inputs)
	var workspaces []*tfe.Workspace
	if inputs.WorkspaceTag != "" {
		for _, workspace := range workspaceList.Items {
			for _, tag := range workspace.Tags {
				fmt.Println(tag.Name, inputs.WorkspaceTag)
				if tag.Name == inputs.WorkspaceTag {
					workspaces = append(workspaces, workspace)
				}
			}
		}
	} else {
		workspaces = workspaceList.Items
	}

	for _, workspace := range workspaces {
		for _, variable := range variables {
			if _, err := tfeClient.Variables.Create(ctx, workspace.ID, tfe.VariableCreateOptions{
				Key:         &variable.Key,
				Value:       &variable.Value,
				Description: &variable.Description,
				Category:    (*tfe.CategoryType)(&variable.Category),
				Sensitive:   &variable.Sensitive,
			}); err != nil {
				fmt.Println(fmt.Errorf("error creating var: %w", err))
			}
		}
	}

	return nil
}
