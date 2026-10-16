# Infrastructure

The Bicep deployment creates Cosmos DB with the application containers, Key Vault, an Azure Container Apps environment, the API container, Cosmos managed-identity access, optional Key Vault access for a Redis Cloud URL, and an Entra authentication gate.

It intentionally leaves these release decisions for the integration phase:

- Azure AI Foundry project/model deployment
- creation of the Entra ID app registration passed as `entraClientId`
- creation of the Redis Cloud database and its Key Vault secret
- the Azure AI Foundry project/model deployment and managed-identity role assignment
- Gmail OAuth callback URLs and secrets (Gmail is not implemented yet)
- container registry and CI federation

Validate from the repository root with `az bicep build --file infra/bicep/main.bicep` when the Azure CLI is available.
