param namePrefix string
param location string
param apiImage string
param cosmosEndpoint string
param keyVaultUri string
param entraTenantId string
param entraClientId string
param allowedPrincipalId string
param aiEndpoint string
param aiDeployment string
param redisSecretUri string

resource environment 'Microsoft.App/managedEnvironments@2024-03-01' = {
  name: '${namePrefix}-env'
  location: location
}

resource api 'Microsoft.App/containerApps@2024-03-01' = {
  name: '${namePrefix}-api'
  location: location
  identity: { type: 'SystemAssigned' }
  properties: {
    managedEnvironmentId: environment.id
    configuration: {
      ingress: {
        external: true
        targetPort: 8080
        transport: 'auto'
      }
      secrets: empty(redisSecretUri) ? [] : [
        {
          name: 'redis-url'
          keyVaultUrl: redisSecretUri
          identity: 'system'
        }
      ]
    }
    template: {
      containers: [{
        name: 'api'
        image: apiImage
        env: concat([
          { name: 'HTTP_ADDR', value: ':8080' }
          { name: 'AZURE_COSMOS_ENDPOINT', value: cosmosEndpoint }
          { name: 'COSMOS_DATABASE', value: 'job-application-assistant' }
          { name: 'KEY_VAULT_URI', value: keyVaultUri }
          { name: 'AZURE_AI_ENDPOINT', value: aiEndpoint }
          { name: 'AZURE_AI_DEPLOYMENT', value: aiDeployment }
          { name: 'REQUIRE_ENTRA_AUTH', value: 'true' }
          { name: 'ALLOWED_ENTRA_PRINCIPAL_ID', value: allowedPrincipalId }
        ], empty(redisSecretUri) ? [] : [{ name: 'REDIS_URL', secretRef: 'redis-url' }])
        resources: { cpu: json('0.25'), memory: '0.5Gi' }
      }]
      scale: { minReplicas: 0, maxReplicas: 1 }
    }
  }
}

resource auth 'Microsoft.App/containerApps/authConfigs@2024-03-01' = {
  parent: api
  name: 'current'
  properties: {
    globalValidation: {
      requireAuthentication: true
      unauthenticatedClientAction: 'Return401'
      excludedPaths: ['/healthz']
    }
    httpSettings: { requireHttps: true }
    identityProviders: {
      azureActiveDirectory: {
        enabled: true
        registration: {
          clientId: entraClientId
          openIdIssuer: 'https://login.microsoftonline.com/${entraTenantId}/v2.0'
        }
        validation: {
          allowedAudiences: ['api://${entraClientId}']
        }
      }
    }
  }
}

output apiUrl string = 'https://${api.properties.configuration.ingress.fqdn}'
output principalId string = api.identity.principalId
