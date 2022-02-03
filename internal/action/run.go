package action

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

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

	var variableUpdates []Variable

	if err := json.Unmarshal([]byte(inputs.Variables), &variableUpdates); err != nil {
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

		variableList, err := tfeClient.Variables.List(ctx, workspace.ID, tfe.VariableListOptions{
			ListOptions: tfe.ListOptions{},
		})
		if err != nil {
			return err
		}

		variableByKey := map[string]*tfe.Variable{}
		for _, variable := range variableList.Items {
			variableByKey[variable.Key] = variable
		}

		for _, variable := range variableUpdates {
			if existingVar, ok := variableByKey[variable.Key]; ok {
				if existingVar.Sensitive || existingVar.Value != variable.Value {
					log.Printf("Updating variable %s for workspace %s\n", variable.Key, workspace.Name)
					if _, err := tfeClient.Variables.Update(ctx, workspace.ID, existingVar.ID, tfe.VariableUpdateOptions{
						Value:       &variable.Value,
						Description: &variable.Description,
					}); err != nil {
						return fmt.Errorf("failed to update variable %s in workspace %s: %w", variable.Key, workspace.Name, err)
					}
				} else {
					log.Printf("No change for variable %s in workspace %s\n", variable.Key, workspace.Name)
				}
			} else {
				log.Printf("Creating variable %s for workspace %s\n", variable.Key, workspace.Name)
				if _, err := tfeClient.Variables.Create(ctx, workspace.ID, tfe.VariableCreateOptions{
					Key:         &variable.Key,
					Value:       &variable.Value,
					Description: &variable.Description,
					Category:    (*tfe.CategoryType)(&variable.Category),
					Sensitive:   &variable.Sensitive,
				}); err != nil {
					return fmt.Errorf("error creating variable %s in workspace %s: %w", variable.Key, workspace.Name, err)
				}

			}
		}
	}

	return nil
}
