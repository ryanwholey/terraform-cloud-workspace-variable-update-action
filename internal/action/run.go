package action

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/hashicorp/go-tfe"
)

type Inputs struct {
	Token        string
	Address      string
	Organization string
	Secrets      string
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

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	kmsClient := kms.NewFromConfig(cfg)

	var secrets map[string]string

	if err := json.Unmarshal([]byte(inputs.Secrets), &secrets); err != nil {
		return err
	}

	plaintext := map[string]string{}
	for name, ciphertext := range secrets {

		decoded, err := base64.StdEncoding.DecodeString(ciphertext)
		if err != nil {
			return fmt.Errorf("failed to base64 decode ciphertext: %w", err)
		}

		out, err := kmsClient.Decrypt(ctx, &kms.DecryptInput{
			CiphertextBlob: []byte(decoded),
		})
		if err != nil {
			return err
		}

		plaintext[name] = string(out.Plaintext)
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
			if _, ok := plaintext[v.Key]; ok {
				log.Printf("Updating %s (%s): %s\n", workspace.Name, workspace.ID, v.Key)

				if _, err := tfeClient.Variables.Update(ctx, workspace.ID, v.ID, tfe.VariableUpdateOptions{
					Value: tfe.String(plaintext[v.Key]),
				}); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
