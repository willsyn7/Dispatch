# Infrastructure

The Bicep skeleton establishes the resource boundaries for Cosmos DB, Key Vault, and the Go API on Azure Container Apps. It intentionally leaves these release decisions for the integration phase:

- Azure AI Foundry project/model deployment
- Entra ID app registrations and Container Apps authentication policy
- Key Vault secrets and role assignments
- Redis Cloud connection details
- Gmail OAuth callback URLs and secrets
- container registry and CI federation

Validate from the repository root with `az bicep build --file infra/bicep/main.bicep` when the Azure CLI is available.

