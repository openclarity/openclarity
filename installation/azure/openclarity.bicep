targetScope = 'subscription'

@description('Location for the OpenClarity resource group')
param location string = 'eastus'

@description('Username for the OpenClarity Server VM')
param adminUsername string

@description('SSH Public Key for the OpenClarity Server VM')
@secure()
param adminSSHKey string

@description('The size of the OpenClarity Server VM')
param serverVmSize string = 'Standard_D2s_v3'

@description('The VM architecture to VM size mapping of the Scanner VMs')
param scannerVmArchitectureToSizeMapping string = 'x86_64:Standard_D2s_v3,arm64:Standard_D2ps_v5'

@description('The architecture of the Scanner VMs')
@allowed([
  'x86_64'
  'arm64'
])
param scannerVmArchitecture string = 'x86_64'

@description('Security Type of the VMClartiy Server VM')
@allowed([
  'Standard'
  'TrustedLaunch'
])
param securityType string = 'TrustedLaunch'

@description ('OpenClarity APIServer Container Image')
param apiserverContainerImage string = 'ghcr.io/openclarity/openclarity-api-server:latest'

@description ('OpenClarity Orchestrator Container Image')
param orchestratorContainerImage string = 'ghcr.io/openclarity/openclarity-orchestrator:latest'

@description ('OpenClarity UI Container Image')
param uiContainerImage string = 'ghcr.io/openclarity/openclarity-ui:latest'

@description ('OpenClarity UIBackend Container Image')
param uibackendContainerImage string = 'ghcr.io/openclarity/openclarity-ui-backend:latest'

@description ('OpenClarity Scanner Container Image')
param scannerContainerImage string = 'ghcr.io/openclarity/openclarity-cli:latest'

@description ('Trivy Server Container Image')
param trivyServerContainerImage string = 'docker.io/aquasec/trivy:0.57.0'

@description ('Grype Server Container Image')
param grypeServerContainerImage string = 'ghcr.io/openclarity/grype-server:v0.7.5'

@description ('Exploit DB Container Image')
param exploitDBContainerImage string = 'ghcr.io/openclarity/exploit-db-server:v0.3.0'

@description ('Freshclam Mirror Container Image')
param freshclamMirrorContainerImage string = 'ghcr.io/openclarity/freshclam-mirror:v0.3.1'

@description ('Yara Rule Server Container Image')
param yaraRuleServerContainerImage string = 'ghcr.io/openclarity/yara-rule-server:v0.3.0'

@description('Postgres Container Image')
param postgresContainerImage string = 'docker.io/bitnami/postgresql:16.3.0-debian-12-r13'

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

@description('OpenClarity Deploy Postfix')
param deploypostfix string

var resourceGroupName = 'openclarity-${deploypostfix}'

resource openClarityResourceGroup 'Microsoft.Resources/resourceGroups@2024-03-01' = {
  name: resourceGroupName
  location: location
}

module openClarityManagedIdentity 'openclarityManagedIdentity.bicep' = {
  name: 'openclarity-managed-identity'
  scope: openClarityResourceGroup
  params: {
    location: location
  }
}

module openClarityScanRole 'openclarityScanRole.bicep' = {
  name: 'openclarity-${deploypostfix}-scan-role'
  scope: openClarityResourceGroup
  params: {
    resourceGroupName: resourceGroupName
    principalID: openClarityManagedIdentity.outputs.openClarityIdentityPrincipalId
  }
}

module openClarityDiscoverRole 'openclarityDiscoverRole.bicep' = {
  name: 'openclarity-${deploypostfix}-discover-role'
  scope: subscription()
  params: {
    resourceGroupName: resourceGroupName
    principalID: openClarityManagedIdentity.outputs.openClarityIdentityPrincipalId
  }
}

module openClarityDeploy 'openclarityDeployModule.bicep' = {
  name: 'openclarity-deploy'
  scope: openClarityResourceGroup
  params: {
    location: location
    adminSSHKey: adminSSHKey
    adminUsername: adminUsername
    serverVmSize: serverVmSize
    scannerVmArchitectureToSizeMapping: scannerVmArchitectureToSizeMapping
    scannerVmArchitecture: scannerVmArchitecture
    securityType: securityType
    openClarityIdentityID: openClarityManagedIdentity.outputs.openClarityIdentityId
    principalID: openClarityManagedIdentity.outputs.openClarityIdentityPrincipalId
    apiserverContainerImage: apiserverContainerImage
    orchestratorContainerImage: orchestratorContainerImage
    uiContainerImage: uiContainerImage
    uibackendContainerImage: uibackendContainerImage
    scannerContainerImage: scannerContainerImage
    trivyServerContainerImage: trivyServerContainerImage
    grypeServerContainerImage: grypeServerContainerImage
    exploitDBContainerImage: exploitDBContainerImage
    freshclamMirrorContainerImage: freshclamMirrorContainerImage
    yaraRuleServerContainerImage: yaraRuleServerContainerImage
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

output adminUsername string = openClarityDeploy.outputs.adminUsername
output hostname string = openClarityDeploy.outputs.hostname
output sshCommand string = openClarityDeploy.outputs.sshCommand
