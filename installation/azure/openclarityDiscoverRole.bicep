targetScope = 'subscription'

@description('OpenClarity Resource Group Name')
param resourceGroupName string

@description('OpenClarity Managed Identity Principal ID')
param principalID string

var discoverRoleID = guid(subscription().id, resourceGroupName, 'openclarity-discoverer-snapshotter')
var discoverRoleName = 'OpenClarity Discoverer Snapshotter for ${resourceGroupName}'
var discoverRoleDescription = 'IAM Role to allow OpenClarity to discover and snapshot virtual machines.'

resource openClarityDiscoverRole 'Microsoft.Authorization/roleDefinitions@2022-04-01' = {
  name: discoverRoleID
  properties: {
    roleName: discoverRoleName
    description: discoverRoleDescription
    type: 'customRole'
    assignableScopes: [
      subscription().id
    ]
    permissions: [
      {
        actions: [
          'Microsoft.Resources/subscriptions/resourceGroups/read'
          'Microsoft.Resources/subscriptions/resourceGroups/moveResources/action'
          'Microsoft.Resources/subscriptions/resourceGroups/validateMoveResources/action'
          'Microsoft.Compute/virtualMachines/read'
          'Microsoft.Compute/disks/read'
          'Microsoft.Compute/snapshots/read'
          'Microsoft.Compute/snapshots/write'
          'Microsoft.Compute/snapshots/delete'
          'Microsoft.Compute/disks/beginGetAccess/action'
        ]
      }
    ]
  }
}

resource discoverRoleAssignment 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(subscription().id, resourceGroupName, 'openclarity-server', discoverRoleName)
  properties: {
    roleDefinitionId: openClarityDiscoverRole.id
    principalId: principalID
    principalType: 'ServicePrincipal'
  }
}
