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
		Tags:        &inputs.WorkspaceTag,
		ListOptions: tfe.ListOptions{},
	})
	if err != nil {
		return err
	}

	for _, workspace := range workspaceList.Items {
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
