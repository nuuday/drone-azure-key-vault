// Copyright 2019 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"context"
	"fmt"

	"github.com/drone/drone-go/drone"
	"github.com/drone/drone-go/plugin/secret"

	log "github.com/sirupsen/logrus"
)

// New returns a new secret plugin.
func New(debug bool) secret.Plugin {
	azureKeyVault, err := NewAzureKeyVault(debug)
	if err != nil {
		log.Fatalf("Error creating the AzureKeyVault: %v", err)
	}
	return &plugin{
		KVClient: azureKeyVault,
	}
}

type plugin struct {
	KVClient *AzureKeyVault
}

func (p *plugin) Find(ctx context.Context, req *secret.Request) (*drone.Secret, error) {
	keyVaultName := req.Path
	secretName := req.Name
	if secretName == "" {
		// If there's no secret name, assume there's only one secret
		// with the key name called "value"
		secretName = "value"
	}

	log.Debugf("Secret request: Key Vault - %v, Name: %v", keyVaultName, secretName)

	// makes an api call to the Azure Key Vault and attempts
	// to retrieve the secret at the requested Key Vault
	secretsList, err := p.KVClient.ListSecrets(keyVaultName)
	if err != nil {
		return nil, fmt.Errorf("key vault not found: %v", err)
	}

	secretsMap := make(map[string]string)
	for _, sec := range secretsList {
		secretsMap[sec.Key] = sec.Value
	}

	value, ok := secretsMap[secretName]
	if !ok {
		return nil, fmt.Errorf("secret key not found")
	}

	// the user can filter out requests based on event type
	// using the X-Drone-Events secret key. Check for this
	// user-defined filter logic.
	events := extractEvents(secretsMap)
	if !match(req.Build.Event, events) {
		return nil, fmt.Errorf("access denied: event does not match")
	}

	// the user can filter out requets based on repository
	// using the X-Drone-Repos secret key. Check for this
	// user-defined filter logic.
	repos := extractRepos(secretsMap)
	if !match(req.Repo.Slug, repos) {
		return nil, fmt.Errorf("access denied: repository does not match")
	}

	// the user can filter out requets based on repository
	// branch using the X-Drone-Branches secret key. Check
	// for this user-defined filter logic.
	branches := extractBranches(secretsMap)
	if !match(req.Build.Target, branches) {
		return nil, fmt.Errorf("access denied: branch does not match")
	}

	return &drone.Secret{
		Name:            secretName,
		Data:            value,
		PullRequest:     true, // always true. use X-Drone-Events to prevent pull requests.
		PullRequestPush: true, // always true. use X-Drone-Events to prevent pull requests.
	}, nil
}
