targetScope = 'resourceGroup'

@description('OpenClarity Resource Group Name')
param resourceGroupName string

@description('OpenClarity Managed Identity Principal ID')
param principalID string

var scanRoleID = guid(resourceGroup().id, 'openclarity-scanner')
var scanRoleName = 'OpenClarity Scanner for ${resourceGroupName}'
var scanRoleDescription = 'IAM Role to allow OpenClarity to deploy virtual machines that mount and scan snapshots.'

resource openClarityScanRole 'Microsoft.Authorization/roleDefinitions@2022-04-01' = {
  name: scanRoleID
  properties: {
    roleName: scanRoleName
    description: scanRoleDescription
    type: 'customRole'
    assignableScopes: [
      resourceGroup().id
    ]
    permissions: [
      {
        actions: [
          'Microsoft.Compute/virtualMachines/read'
          'Microsoft.Compute/virtualMachines/write'
          'Microsoft.Compute/virtualMachines/delete'
          'Microsoft.Compute/snapshots/read'
          'Microsoft.Compute/snapshots/write'
          'Microsoft.Compute/snapshots/delete'
          'Microsoft.Compute/disks/read'
          'Microsoft.Compute/disks/write'
          'Microsoft.Compute/disks/delete'
          'Microsoft.Network/networkInterfaces/write'
          'Microsoft.Network/networkInterfaces/read'
          'Microsoft.Network/networkInterfaces/delete'
          'Microsoft.Network/networkSecurityGroups/join/action'
          'Microsoft.Network/virtualNetworks/subnets/join/action'
          'Microsoft.Network/networkInterfaces/join/action'
          'Microsoft.Compute/snapshots/beginGetAccess/action'
          'Microsoft.Compute/snapshots/endGetAccess/action'
          'Microsoft.Storage/storageAccounts/listkeys/action'
        ]
      }
    ]
  }
}

resource scanRoleAssignment 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(resourceGroup().id, 'openclarity-server', openClarityScanRole.id)
  properties: {
    roleDefinitionId: openClarityScanRole.id
    principalId: principalID
    principalType: 'ServicePrincipal'
  }
}
