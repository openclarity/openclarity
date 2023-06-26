targetScope = 'resourceGroup'

@description('Location where to create the resources')
param location string = resourceGroup().location

var vmClarityIdentityName = 'vmclarity-discoverer-deployer-${uniqueString(resourceGroup().id)}'

resource vmClarityServerIdentity 'Microsoft.ManagedIdentity/userAssignedIdentities@2018-11-30' = {
  name: vmClarityIdentityName
  location: location
}
output vmClarityIdentityId string = vmClarityServerIdentity.id
output vmClarityIdentityPrincipalId string = vmClarityServerIdentity.properties.principalId
