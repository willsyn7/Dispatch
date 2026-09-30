targetScope = 'resourceGroup'

@description('Short lowercase name used as an Azure resource prefix.')
param namePrefix string

@description('Primary Azure region for the application.')
param location string = resourceGroup().location

@description('Existing container image for the API.')
param apiImage string

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
  }
}

output apiUrl string = app.outputs.apiUrl

