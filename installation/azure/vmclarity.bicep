targetScope = 'subscription'

@description('Location for the VMClarity resource group')
param location string = 'eastus'

@description('Username for the VMClarity Server VM')
param adminUsername string

@description('SSH Public Key for the VMClarity Server VM')
@secure()
param adminSSHKey string

@description('The size of the VMClarity Server VM')
param serverVmSize string = 'Standard_D2s_v3'

@description('The size of the Scanner VMs')
param scannerVmSize string = 'Standard_D2s_v3'

@description('Security Type of the VMClartiy Server VM')
@allowed([
  'Standard'
  'TrustedLaunch'
])
param securityType string = 'TrustedLaunch'

@description ('VMClarity APIServer Container Image')
param apiserverContainerImage string = 'ghcr.io/openclarity/vmclarity-apiserver:latest'

@description ('VMClarity Orchestrator Container Image')
param orchestratorContainerImage string = 'ghcr.io/openclarity/vmclarity-orchestrator:latest'

@description ('VMClarity UI Container Image')
param uiContainerImage string = 'ghcr.io/openclarity/vmclarity-ui:latest'

@description ('VMClarity UIBackend Container Image')
param uibackendContainerImage string = 'ghcr.io/openclarity/vmclarity-ui-backend:latest'

@description ('VMClarity Scanner Container Image')
param scannerContainerImage string = 'ghcr.io/openclarity/vmclarity-cli:latest'

@description ('Trivy Server Container Image')
param trivyServerContainerImage string = 'docker.io/aquasec/trivy:0.41.0'

@description ('Grype Server Container Image')
param grypeServerContainerImage string = 'ghcr.io/openclarity/grype-server:v0.7.0'

@description ('Exploit DB Container Image')
param exploitDBContainerImage string = 'ghcr.io/openclarity/exploit-db-server:v0.2.4'

@description ('Freshclam Mirror Container Image')
param freshclamMirrorContainerImage string = 'ghcr.io/openclarity/freshclam-mirror:v0.3.0'

@description('Postgres Container Image')
param postgresContainerImage string = 'docker.io/bitnami/postgresql:12.14.0-debian-11-r28'

@description('Asset Scan Delete Policy')
@allowed([
  'Always'
  'OnSuccess'
  'Never'
])
param assetScanDeletePolicy string = 'Always'

@description('Database to Use')
@allowed([
  'Postgresql'
  'External Postgresql'
  'SQLite'
])
param databaseToUse string = 'SQLite'

@description('Password to configure Postgresql with on first install. Required if Postgres is selected as the Database To Use. Do not change this on stack update.')
@secure()
param postgresDBPassword string = ''

@description('Hostname or IP address of the External DB to connect to. Required if an external database type is selected as the Database To Use.')
param externalDBHost string = ''

@description('Network Port of the External DB to connect to. Required if an external database type is selected as the Database To Use.')
@minValue(0)
param externalDBPort int = 0

@description('Name of the Database to use on the External DB. Required if an external database type is selected as the Database To Use.')
param externalDBName string = ''

@description('Username to use to connect to the External DB. Required if an external database type is selected as the Database To Use.')
param externalDBUsername string = ''

@description('Password to use to connect to the External DB. Required if an external database type is selected as the Database To Use.')
@secure()
param externalDBPassword string = ''

@description('VMClarity Deploy Postfix')
param deploypostfix string

var resourceGroupName = 'vmclarity-${deploypostfix}'

resource vmClarityResourceGroup 'Microsoft.Resources/resourceGroups@2022-09-01' = {
  name: resourceGroupName
  location: location
}

module vmClarityManagedIdentity 'vmclarityManagedIdentity.bicep' = {
  name: 'vmclarity-managed-identity'
  scope: vmClarityResourceGroup
  params: {
    location: location
  }
}

module vmClarityScanRole 'vmclarityScanRole.bicep' = {
  name: 'vmclarity-${deploypostfix}-scan-role'
  scope: vmClarityResourceGroup
  params: {
    principalID: vmClarityManagedIdentity.outputs.vmClarityIdentityPrincipalId
  }
}

module vmClarityDiscoverRole 'vmclarityDiscoverRole.bicep' = {
  name: 'vmclarity-${deploypostfix}-discover-role'
  scope: subscription()
  params: {
    resourceGroupName: resourceGroupName
    principalID: vmClarityManagedIdentity.outputs.vmClarityIdentityPrincipalId
  }
}

module vmClarityDeploy 'vmclarityDeployModule.bicep' = {
  name: 'vmclarity-deploy'
  scope: vmClarityResourceGroup
  params: {
    location: location
    adminSSHKey: adminSSHKey
    adminUsername: adminUsername
    serverVmSize: serverVmSize
    scannerVmSize: scannerVmSize
    securityType: securityType
    vmClarityIdentityID: vmClarityManagedIdentity.outputs.vmClarityIdentityId
    principalID: vmClarityManagedIdentity.outputs.vmClarityIdentityPrincipalId
    apiserverContainerImage: apiserverContainerImage
    orchestratorContainerImage: orchestratorContainerImage
    uiContainerImage: uiContainerImage
    uibackendContainerImage: uibackendContainerImage
    scannerContainerImage: scannerContainerImage
    trivyServerContainerImage: trivyServerContainerImage
    grypeServerContainerImage: grypeServerContainerImage
    exploitDBContainerImage: exploitDBContainerImage
    freshclamMirrorContainerImage: freshclamMirrorContainerImage
    postgresContainerImage: postgresContainerImage
    assetScanDeletePolicy: assetScanDeletePolicy
    databaseToUse: databaseToUse
    postgresDBPassword: postgresDBPassword
    externalDBHost: externalDBHost
    externalDBPort: externalDBPort
    externalDBName: externalDBName
    externalDBUsername: externalDBUsername
    externalDBPassword: externalDBPassword
  }
}

output adminUsername string = vmClarityDeploy.outputs.adminUsername
output hostname string = vmClarityDeploy.outputs.hostname
output sshCommand string = vmClarityDeploy.outputs.sshCommand
