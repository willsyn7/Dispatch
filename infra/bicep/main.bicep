targetScope = 'resourceGroup'

@description('Short lowercase name used as an Azure resource prefix.')
param namePrefix string

@description('Primary Azure region for the application.')
param location string = resourceGroup().location

@description('Existing container image for the API.')
param apiImage string

@description('Microsoft Entra tenant containing the single-user app registration.')
param entraTenantId string = tenant().tenantId

@description('Client ID of the Entra app registration accepted by the API.')
param entraClientId string

@description('Object ID of the only user allowed by the backend. Leave empty to accept any authenticated tenant user.')
param allowedPrincipalId string = ''

@description('Azure AI Foundry resource endpoint.')
param aiEndpoint string = ''

@description('Azure AI Foundry model deployment name.')
param aiDeployment string = ''

@description('Full Key Vault secret URI containing the Redis Cloud connection URL.')
param redisSecretUri string = ''

module data 'modules/data.bicep' = {
  name: '${namePrefix}-data'
  params: {
    namePrefix: namePrefix
    location: location
  }
}

module app 'modules/container-app.bicep' = {
  name: '${namePrefix}-api'
  params: {
    namePrefix: namePrefix
    location: location
    apiImage: apiImage
    cosmosEndpoint: data.outputs.cosmosEndpoint
    keyVaultUri: data.outputs.keyVaultUri
    entraTenantId: entraTenantId
    entraClientId: entraClientId
    allowedPrincipalId: allowedPrincipalId
    aiEndpoint: aiEndpoint
    aiDeployment: aiDeployment
    redisSecretUri: redisSecretUri
  }
}

resource cosmosAccount 'Microsoft.DocumentDB/databaseAccounts@2024-11-15' existing = {
  name: data.outputs.cosmosName
}

resource cosmosDataRole 'Microsoft.DocumentDB/databaseAccounts/sqlRoleAssignments@2024-11-15' = {
  parent: cosmosAccount
  name: guid(cosmosAccount.id, app.outputs.principalId, 'data-contributor')
  properties: {
    principalId: app.outputs.principalId
    roleDefinitionId: '${cosmosAccount.id}/sqlRoleDefinitions/00000000-0000-0000-0000-000000000002'
    scope: cosmosAccount.id
  }
}

resource vault 'Microsoft.KeyVault/vaults@2024-11-01' existing = {
  name: data.outputs.keyVaultName
}

resource vaultSecretsRole 'Microsoft.Authorization/roleAssignments@2022-04-01' = if (!empty(redisSecretUri)) {
  name: guid(vault.id, app.outputs.principalId, 'secrets-user')
  scope: vault
  properties: {
    principalId: app.outputs.principalId
    principalType: 'ServicePrincipal'
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', '4633458b-17de-408a-b874-0445c86b69e6')
  }
}

output apiUrl string = app.outputs.apiUrl
