package plugin

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"path"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/keyvault/keyvault"
	kvauth "github.com/Azure/azure-sdk-for-go/services/keyvault/auth"
	"github.com/Azure/go-autorest/autorest"
	log "github.com/sirupsen/logrus"
)

type AzureKeyVault struct {
	Authorizer *autorest.Authorizer
	BaseClient *keyvault.BaseClient
}

type Secret struct {
	Key       string
	Value     string
	VaultName string
	Tags      map[string]*string
}

const (
	VAULT_URL = "https://%v.vault.azure.net"
)

var (
	timeout = 5000 * time.Millisecond
)

func NewAzureKeyVault(debug bool) (*AzureKeyVault, error) {
	authorizer, err := kvauth.NewAuthorizerFromEnvironment()
	if err != nil {
		log.Errorf("Unable to create vault authorizer: %v", err)
		return nil, err
	}

	baseClient := keyvault.New()
	baseClient.Authorizer = authorizer

	// https://github.com/Azure/azure-sdk-for-go#inspecting-and-debugging
	if debug {
		baseClient.RequestInspector = logRequest()
		baseClient.ResponseInspector = logResponse()
	}

	azureKeyVault := &AzureKeyVault{
		Authorizer: &authorizer,
		BaseClient: &baseClient,
	}
	return azureKeyVault, nil
}

func (akv *AzureKeyVault) ListSecrets(vaultName string) ([]*Secret, error) {
	log.Debugf("Listing secrets from vault: %v", vaultName)
	secretsList := []*Secret{}

	contextDeadline := time.Now().Add(timeout)
	ctx, cancel := context.WithDeadline(context.Background(), contextDeadline)
	defer cancel()

	azureSecretList, err := akv.BaseClient.GetSecrets(ctx, fmt.Sprintf(VAULT_URL, vaultName), nil)
	if err != nil {
		log.Errorf("Unable to get list of secrets: %v", err)
		return nil, err
	}

	for _, secret := range azureSecretList.Values() {
		secName := path.Base(*secret.ID)
		secObject, err := akv.GetSecret(vaultName, secName)
		if err != nil {
			return nil, err
		}
		sec := &Secret{
			Key:       secName,
			Value:     secObject.Value,
			VaultName: vaultName,
			Tags:      secret.Tags,
		}
		secretsList = append(secretsList, sec)
	}
	return secretsList, nil
}

func (akv *AzureKeyVault) GetSecret(vaultName, secretName string) (*Secret, error) {
	contextDeadline := time.Now().Add(timeout)
	ctx, cancel := context.WithDeadline(context.Background(), contextDeadline)
	defer cancel()

	// Always fetch latest version of the secret
	secretVersion := ""

	secretResp, err := akv.BaseClient.GetSecret(ctx,
		fmt.Sprintf(VAULT_URL, vaultName), secretName, secretVersion)
	if err != nil {
		log.Errorf("Unable to get value for secret: %v", err)
		return nil, err
	}

	return &Secret{
		Key:   secretName,
		Value: fmt.Sprintf("%v", *secretResp.Value),
	}, nil
}

func logRequest() autorest.PrepareDecorator {
	return func(p autorest.Preparer) autorest.Preparer {
		return autorest.PreparerFunc(func(r *http.Request) (*http.Request, error) {
			r, err := p.Prepare(r)
			if err != nil {
				log.Errorf("Error request prepare: %v", err)
			}
			dump, _ := httputil.DumpRequestOut(r, true)
			log.Debugf(string(dump))
			return r, err
		})
	}
}

func logResponse() autorest.RespondDecorator {
	return func(p autorest.Responder) autorest.Responder {
		return autorest.ResponderFunc(func(r *http.Response) error {
			err := p.Respond(r)
			if err != nil {
				log.Errorf("Error request respond: %v", err)
			}
			dump, _ := httputil.DumpResponse(r, true)
			log.Debugf(string(dump))
			return err
		})
	}
}
