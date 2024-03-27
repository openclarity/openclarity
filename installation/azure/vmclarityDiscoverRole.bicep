targetScope = 'subscription'

@description('VMClarity Resource Group Name')
param resourceGroupName string

@description('VMClarity Managed Identity Principal ID')
param principalID string

var discoverRoleID = guid(subscription().id, resourceGroupName, 'vmclarity-discoverer-snapshotter')
var discoverRoleName = 'VMClarity Discoverer Snapshotter for ${resourceGroupName}'
var discoverRoleDescription = 'IAM Role to allow VMClarity to discover and snapshot virtual machines.'

resource vmClarityDiscoverRole 'Microsoft.Authorization/roleDefinitions@2022-04-01' = {
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
  name: guid(subscription().id, resourceGroupName, 'vmclarity-server', discoverRoleName)
  properties: {
    roleDefinitionId: vmClarityDiscoverRole.id
    principalId: principalID
    principalType: 'ServicePrincipal'
  }
}
