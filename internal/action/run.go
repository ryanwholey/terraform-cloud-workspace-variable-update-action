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

	fmt.Println(inputs.Variables)

	var variables []Variable

	if err := json.Unmarshal([]byte(inputs.Variables), &variables); err != nil {
		return err
	}

	varsMap := map[string]Variable{}
	for _, v := range variables {
		varsMap[v.Key] = v
	}

	// TODO: handle pagination
	workspaceList, err := tfeClient.Workspaces.List(ctx, inputs.Organization, tfe.WorkspaceListOptions{
		ListOptions: tfe.ListOptions{},
	})
	if err != nil {
		return err
	}

	for _, workspace := range workspaceList.Items {
		// TODO: handle pagination
		varList, err := tfeClient.Variables.List(ctx, workspace.ID, tfe.VariableListOptions{
			ListOptions: tfe.ListOptions{},
		})
		if err != nil {
			return err
		}

		for _, v := range varList.Items {
			if _, ok := varsMap[v.Key]; ok {
				log.Printf("Updating %s (%s): %s\n", workspace.Name, workspace.ID, v.Key)

				if _, err := tfeClient.Variables.Update(ctx, workspace.ID, v.ID, tfe.VariableUpdateOptions{
					Value: tfe.String(v.Value),
				}); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
