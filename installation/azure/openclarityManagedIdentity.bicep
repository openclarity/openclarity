targetScope = 'resourceGroup'

@description('Location where to create the resources')
param location string = resourceGroup().location

var openClarityIdentityName = 'openclarity-discoverer-deployer-${uniqueString(resourceGroup().id)}'

resource openClarityServerIdentity 'Microsoft.ManagedIdentity/userAssignedIdentities@2023-01-31' = {
  name: openClarityIdentityName
  location: location
}
output openClarityIdentityId string = openClarityServerIdentity.id
output openClarityIdentityPrincipalId string = openClarityServerIdentity.properties.principalId
