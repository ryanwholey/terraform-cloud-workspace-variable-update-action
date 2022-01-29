package action

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/hashicorp/go-tfe"
)

type Inputs struct {
	Token        string
	Address      string
	Organization string
	KeyID        string
	SecretsPath  string
}

type EncryptedSecretsSecrets struct {
	Secrets map[string]string
}

type EncryptedSecrets struct {
	Locals EncryptedSecretsSecrets
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

	b, err := ioutil.ReadFile(inputs.SecretsPath)
	if err != nil {
		return err
	}

	var secrets EncryptedSecrets

	if err := json.Unmarshal(b, &secrets); err != nil {
		return err
	}

	fmt.Println(string(b))
	fmt.Println(inputs.KeyID)

	plaintext := map[string]string{}
	for name, ciphertext := range secrets.Locals.Secrets {

		fmt.Println(ciphertext)

		out, err := kmsClient.Decrypt(ctx, &kms.DecryptInput{
			CiphertextBlob: []byte(ciphertext),
			KeyId:          &inputs.KeyID,
		})
		if err != nil {
			return err
		}

		plaintext[name] = string(out.Plaintext)
	}

	workspaceList, err := tfeClient.Workspaces.List(ctx, inputs.Organization, tfe.WorkspaceListOptions{
		ListOptions: tfe.ListOptions{},
	})
	if err != nil {
		return err
	}

	fmt.Println("next page: ", workspaceList.NextPage)

	for _, workspace := range workspaceList.Items {
		varList, err := tfeClient.Variables.List(ctx, workspace.ID, tfe.VariableListOptions{
			ListOptions: tfe.ListOptions{},
		})
		if err != nil {
			return err
		}

		for _, v := range varList.Items {
			if _, ok := plaintext[v.Key]; ok {
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
