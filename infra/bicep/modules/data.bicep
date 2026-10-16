param namePrefix string
param location string
param enableFreeTier bool = true

resource cosmos 'Microsoft.DocumentDB/databaseAccounts@2024-11-15' = {
  name: '${namePrefix}-cosmos'
  location: location
  kind: 'GlobalDocumentDB'
  properties: {
    databaseAccountOfferType: 'Standard'
    enableFreeTier: enableFreeTier
    consistencyPolicy: { defaultConsistencyLevel: 'Session' }
    locations: [{ locationName: location, failoverPriority: 0 }]
  }
}

resource database 'Microsoft.DocumentDB/databaseAccounts/sqlDatabases@2024-11-15' = {
  parent: cosmos
  name: 'job-application-assistant'
  properties: {
    resource: { id: 'job-application-assistant' }
    options: { autoscaleSettings: { maxThroughput: 1000 } }
  }
}

resource applications 'Microsoft.DocumentDB/databaseAccounts/sqlDatabases/containers@2024-11-15' = {
  parent: database
  name: 'applications'
  properties: {
    resource: {
      id: 'applications'
      partitionKey: { paths: ['/userId'], kind: 'Hash' }
    }
  }
}

resource drafts 'Microsoft.DocumentDB/databaseAccounts/sqlDatabases/containers@2024-11-15' = {
  parent: database
  name: 'drafts'
  properties: {
    resource: {
      id: 'drafts'
      partitionKey: { paths: ['/applicationId'], kind: 'Hash' }
    }
  }
}

resource profile 'Microsoft.DocumentDB/databaseAccounts/sqlDatabases/containers@2024-11-15' = {
  parent: database
  name: 'profile'
  properties: {
    resource: {
      id: 'profile'
      partitionKey: { paths: ['/userId'], kind: 'Hash' }
    }
  }
}

resource vault 'Microsoft.KeyVault/vaults@2024-11-01' = {
  name: '${namePrefix}-kv'
  location: location
  properties: {
    tenantId: tenant().tenantId
    sku: { family: 'A', name: 'standard' }
    enableRbacAuthorization: true
    enableSoftDelete: true
    softDeleteRetentionInDays: 7
  }
}

output cosmosEndpoint string = cosmos.properties.documentEndpoint
output keyVaultUri string = vault.properties.vaultUri
output cosmosName string = cosmos.name
output keyVaultName string = vault.name
