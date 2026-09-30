param namePrefix string
param location string
param apiImage string
param cosmosEndpoint string
param keyVaultUri string

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
    }
    template: {
      containers: [{
        name: 'api'
        image: apiImage
        env: [
          { name: 'HTTP_ADDR', value: ':8080' }
          { name: 'COSMOS_ENDPOINT', value: cosmosEndpoint }
          { name: 'KEY_VAULT_URI', value: keyVaultUri }
        ]
        resources: { cpu: json('0.25'), memory: '0.5Gi' }
      }]
      scale: { minReplicas: 0, maxReplicas: 1 }
    }
  }
}

output apiUrl string = 'https://${api.properties.configuration.ingress.fqdn}'

