A secret extension to A secret extension that provides optional support for sourcing secrets from Azure Key Vault. _Please note this project requires Drone server version 1.4 or higher._

## Installation

Create a shared secret:

```console
$ openssl rand -hex 16
bea26a2221fd8090ea38720fc445eca6
```

Download and run the plugin:

```
$ docker run -d \
  --publish=3000:3000 \
  --env=DRONE_DEBUG=true \
  --env=DRONE_SECRET=bea26a2221fd8090ea38720fc445eca6 \
  --env=AZURE_TENANT_ID=$AZURE_TENANT_ID \
  --env=AZURE_CLIENT_ID=$AZURE_CLIENT_ID \
  --env=AZURE_CLIENT_SECRET=$AZURE_CLIENT_SECRET \
  --restart=always \
  --name=secrets <docker_repo>/drone-azure-key-vault
```

Update your runner configuration to include the plugin address and the shared secret.

```text
DRONE_SECRET_PLUGIN_ENDPOINT=http://1.2.3.4:3000
DRONE_SECRET_PLUGIN_TOKEN=bea26a2221fd8090ea38720fc445eca6
```

## Azure Key Vault

[Azure Key Vault](https://azure.microsoft.com/en-us/services/key-vault/) is a
tool for securely storing and accessing secrets. The Azure Key Vault extension
provides your pipeline with access to Azure Key Vault secrets.

### Required Azure environment variables

- `AZURE_TENANT_ID`: Specifies the Tenant to which to authenticate.

- `AZURE_CLIENT_ID`: Specifies the app client ID to use.

- `AZURE_CLIENT_SECRET`: Specifies the app secret to use.

The app client specified in the environment variables needs to have `READ` access
to the Key Vaults which are going to be accessed in the pipelines.

### Creating secrets

Use the [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli)
to create secrets in the Key Vault.
In the below example we store the Docker username and password.

```
$ az keyvault secret set --vault-name vault-dev --name docker-username --value user
$ az keyvault secret set --vault-name vault-dev --name docker-password --value pass
```

### Accessing the secrets

Once the secrets are stored in Azure key vault, we can update the yaml
configuration to use those secrets.
To access them, first we need to define a secret resource for each external
secret:

```
---
kind: secret
name: docker-username
get:
  path: vault-dev
  name: docker-username

---
kind: secret
name: docker-password
get:
  path: vault-dev
  name: docker-password
```

The `path` to the secret is the `Azure Key Vault` name, and `name` is the
secret name we want to fetch from the Key Vault.

Referencing them in the yaml configuration:

```
kind: pipeline
name: default

steps:
- name: build
  image: alpine
  environment:
    DOCKER_USERNAME:
      from_secret: docker-username
    DOCKER_PASSWORD:
      from_secret: docker-password

---
kind: secret
name: docker-username
get:
  path: vault-dev
  name: docker-username

---
kind: secret
name: docker-password
get:
  path: vault-dev
  name: docker-password

...
```

### Limiting access

Secrets are available to all repositories and all build events by default.

Limiting the access works in the same way as for the other Drone external secrets
plugin. More details can be found [here](https://docs.drone.io/secret/external/vault/#limiting-access).

**Note**: The access is limited at a Vault level currently.

Example: Limiting the Key Vault secrets to be used in a single repository

```
$ az keyvault secret set --vault-name vault-dev --name x-drone-repos --value octocat/hello-world
```
