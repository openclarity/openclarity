@description('Username for the OpenClarity Server VM')
param adminUsername string

@description('SSH Public Key for the OpenClarity Server VM')
@secure()
param adminSSHKey string

@description('The size of the OpenClarity Server VM')
param serverVmSize string = 'Standard_D2s_v3'

@description('The VM architecture to VM size mapping of the Scanner VMs')
param scannerVmArchitectureToSizeMapping string = 'x86_64:Standard_D2s_v3,arm64:Standard_D2ps_v5'

@description('The architecture to image sku mapping of the Scanner VMs')
param scannerVMArchitectureToImageSkuMapping string = 'x86_64:20_04-lts-gen2,arm64:20_04-lts-arm64'

@description('The architecture of the Scanner VMs')
@allowed([
  'x86_64'
  'arm64'
])
param scannerVmArchitecture string

@description('Location where to create the resources')
param location string = resourceGroup().location

@description('Public IP DNS prefix')
param dnsLabelPrefix string = toLower('openclarity-server-${uniqueString(resourceGroup().id)}')

@description('Security Type of the VMClartiy Server VM')
@allowed([
  'Standard'
  'TrustedLaunch'
])
param securityType string = 'TrustedLaunch'

@description('OpenClarity Server Identity ID')
param openClarityIdentityID string

@description('OpenClarity Managed Identity Principal ID')
param principalID string

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
param trivyServerContainerImage string = 'docker.io/aquasec/trivy:0.56.2'

@description ('Grype Server Container Image')
param grypeServerContainerImage string = 'ghcr.io/openclarity/grype-server:v0.7.4'

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

var imageReference = {
    publisher: 'Canonical'
    offer: '0001-com-ubuntu-server-focal'
    sku: '20_04-lts-gen2'
    version: 'latest'
}

var openClarityNetName = 'openclarity-server-net'
var addressPrefix = '10.1.0.0/16'

var openClarityServerSubnetName = 'openclarity-server-subnet'
var openClarityServerSecurityGroupName = 'openclarity-server-security-group'
var openClarityServerSubnetAddressPrefix = '10.1.0.0/24'

var openClarityScannerSubnetName = 'openclarity-scanner-subnet'
var openClarityScannerSecurityGroupName = 'openclarity-scanner-security-group'
var openClarityScannerSubnetAddressPrefix = '10.1.1.0/24'

var openClarityServerVMName = 'openclarity-server'
var publicIPAddressName = '${openClarityServerVMName}-public-ip'
var networkInterfaceName = '${openClarityServerVMName}-net-int'

var params = {
  APIServerContainerImage: apiserverContainerImage
  OrchestratorContainerImage: orchestratorContainerImage
  UIContainerImage: uiContainerImage
  UIBackendContainerImage: uibackendContainerImage
  ScannerContainerImage: scannerContainerImage
  TrivyServerContainerImage: trivyServerContainerImage
  GrypeServerContainerImage: grypeServerContainerImage
  YaraRuleServerContainerImage: yaraRuleServerContainerImage
  ExploitDBServerContainerImage: exploitDBContainerImage
  FreshclamMirrorContainerImage: freshclamMirrorContainerImage
  PostgresqlContainerImage: postgresContainerImage
  ScannerVmArchitectureToSizeMapping: scannerVmArchitectureToSizeMapping
  ScannerVmArchitecture: scannerVmArchitecture
  AssetScanDeletePolicy: assetScanDeletePolicy
  DatabaseToUse: databaseToUse
  PostgresDBPassword: postgresDBPassword
  ExternalDBHost: externalDBHost
  ExternalDBPort: string(externalDBPort)
  ExternalDBName: externalDBName
  ExternalDBUsername: externalDBUsername
  ExternalDBPassword: externalDBPassword
  AdminUsername: adminUsername
  AZURE_SUBSCRIPTION_ID: subscription().subscriptionId
  AZURE_SCANNER_LOCATION: location
  AZURE_SCANNER_RESOURCE_GROUP: resourceGroup().name
  AZURE_SCANNER_SUBNET_ID: openClarityNet::openClarityScannerSubnet.id
  AZURE_SCANNER_PUBLIC_KEY: base64(adminSSHKey)
  AZURE_SCANNER_VM_ARCHITECTURE_TO_SIZE_MAPPING: scannerVmArchitectureToSizeMapping
  AZURE_SCANNER_VM_ARCHITECTURE: scannerVmArchitecture
  AZURE_SCANNER_IMAGE_PUBLISHER: imageReference.publisher
  AZURE_SCANNER_IMAGE_OFFER: imageReference.offer
  AZURE_SCANNER_VM_ARCHITECTURE_TO_IMAGE_SKU_MAPPING: scannerVMArchitectureToImageSkuMapping
  AZURE_SCANNER_IMAGE_VERSION: imageReference.version
  AZURE_SCANNER_SECURITY_GROUP: openClarityScannerSecurityGroup.id
  AZURE_SCANNER_STORAGE_ACCOUNT_NAME: storageAccountName
  AZURE_SCANNER_STORAGE_CONTAINER_NAME: snapshotContainerName
}

var scriptTemplate = loadTextContent('openclarity-install.sh')

var renderedScript = reduce(
  items(params),
  {value: scriptTemplate},
  (curr, next) => {value: replace(curr.value, '__${next.key}__', next.value)}
).value

var osDiskType = 'StandardSSD_LRS'
var linuxConfiguration = {
  disablePasswordAuthentication: true
  ssh: {
    publicKeys: [
      {
        path: '/home/${adminUsername}/.ssh/authorized_keys'
        keyData: adminSSHKey
      }
    ]
  }
}
var securityProfileJson = {
  uefiSettings: {
    secureBootEnabled: true
    vTpmEnabled: true
  }
  securityType: securityType
}

var openClarityGuestAttestationName = 'OpenClarityServerGuestAttestation'
var extensionName = 'GuestAttestation'
var extensionPublisher = 'Microsoft.Azure.Security.LinuxAttestation'
var extensionVersion = '1.0'
var maaTenantName = 'GuestAttestation'
var maaEndpoint = substring('emptystring', 0, 0)

var openClarityServerCustomScriptName = 'OpenClarityServerCustomScript'

resource networkInterface 'Microsoft.Network/networkInterfaces@2024-01-01' = {
  name: networkInterfaceName
  location: location
  properties: {
    ipConfigurations: [
      {
        name: 'ipconfig1'
        properties: {
          subnet: {
            id: openClarityNet::openClarityServerSubnet.id
          }
          privateIPAllocationMethod: 'Dynamic'
          publicIPAddress: {
            id: publicIPAddress.id
          }
        }
      }
    ]
    networkSecurityGroup: {
      id: openClarityServerSecurityGroup.id
    }
  }
}

resource openClarityServerSecurityGroup 'Microsoft.Network/networkSecurityGroups@2024-01-01' = {
  name: openClarityServerSecurityGroupName
  location: location
  properties: {
    securityRules: [
      {
        name: 'SSH'
        properties: {
          priority: 1000
          protocol: 'Tcp'
          access: 'Allow'
          direction: 'Inbound'
          sourceAddressPrefix: '*'
          sourcePortRange: '*'
          destinationAddressPrefix: '*'
          destinationPortRange: '22'
        }
      }
    ]
  }
}

// Declare subnets inside of virtualNet so that they don't get deleted when
// re-applying the template
// https://github.com/Azure/bicep/issues/4653
resource openClarityNet 'Microsoft.Network/virtualNetworks@2024-01-01' = {
  name: openClarityNetName
  location: location
  properties: {
    addressSpace: {
      addressPrefixes: [
        addressPrefix
      ]
    }
    subnets:[
      {
        name: openClarityServerSubnetName
        properties: {
          addressPrefix: openClarityServerSubnetAddressPrefix
          serviceEndpoints: [
            {
              service: 'Microsoft.Storage'
            }
          ]
          privateEndpointNetworkPolicies: 'Enabled'
          privateLinkServiceNetworkPolicies: 'Enabled'
        }
      }
      {
        name: openClarityScannerSubnetName
        properties: {
          addressPrefix: openClarityScannerSubnetAddressPrefix
          privateEndpointNetworkPolicies: 'Enabled'
          privateLinkServiceNetworkPolicies: 'Enabled'
        }
      }
    ]
  }

  resource openClarityServerSubnet 'subnets' existing = {
    name: openClarityServerSubnetName
  }

  resource openClarityScannerSubnet 'subnets' existing = {
    name: openClarityScannerSubnetName
  }
}

resource publicIPAddress 'Microsoft.Network/publicIPAddresses@2024-01-01' = {
  name: publicIPAddressName
  location: location
  sku: {
    name: 'Basic'
  }
  properties: {
    publicIPAllocationMethod: 'Dynamic'
    publicIPAddressVersion: 'IPv4'
    dnsSettings: {
      domainNameLabel: dnsLabelPrefix
    }
    idleTimeoutInMinutes: 4
  }
}

resource openClarityServer 'Microsoft.Compute/virtualMachines@2024-03-01' = {
  name: openClarityServerVMName
  location: location
  identity: {
    type: 'UserAssigned'
    userAssignedIdentities:{
      '${openClarityIdentityID}': {}
    }
  }
  properties: {
    hardwareProfile: {
      vmSize: serverVmSize
    }
    storageProfile: {
      osDisk: {
        createOption: 'FromImage'
        managedDisk: {
          storageAccountType: osDiskType
        }
      }
      imageReference: imageReference
    }
    networkProfile: {
      networkInterfaces: [
        {
          id: networkInterface.id
        }
      ]
    }
    osProfile: {
      computerName: openClarityServerVMName
      adminUsername: adminUsername
      linuxConfiguration: linuxConfiguration
    }
    securityProfile: ((securityType == 'TrustedLaunch') ? securityProfileJson : null)
  }
}

resource openClarityServerGuestAttestation 'Microsoft.Compute/virtualMachines/extensions@2024-03-01' = if ((securityType == 'TrustedLaunch') && ((securityProfileJson.uefiSettings.secureBootEnabled == true) && (securityProfileJson.uefiSettings.vTpmEnabled == true))) {
  parent: openClarityServer
  name: openClarityGuestAttestationName
  location: location
  properties: {
    publisher: extensionPublisher
    type: extensionName
    typeHandlerVersion: extensionVersion
    autoUpgradeMinorVersion: true
    enableAutomaticUpgrade: true
    settings: {
      AttestationConfig: {
        MaaSettings: {
          maaEndpoint: maaEndpoint
          maaTenantName: maaTenantName
        }
      }
    }
  }
}

resource openClarityServerCustomScript 'Microsoft.Compute/virtualMachines/extensions@2024-03-01' = {
  parent: openClarityServer
  dependsOn: [
    openClarityServerGuestAttestation
  ]
  name: openClarityServerCustomScriptName
  location: location
  properties: {
    publisher: 'Microsoft.Azure.Extensions'
    type: 'CustomScript'
    typeHandlerVersion: '2.1'
    autoUpgradeMinorVersion: true
    protectedSettings: {
      script: base64(renderedScript)
    }
  }
}

resource openClarityScannerSecurityGroup 'Microsoft.Network/networkSecurityGroups@2024-01-01' = {
  name: openClarityScannerSecurityGroupName
  location: location
  properties: {
    securityRules: [
      {
        name: 'SSH-From-OpenClarity-Server'
        properties: {
          priority: 1000
          protocol: 'Tcp'
          access: 'Allow'
          direction: 'Inbound'
          sourceAddressPrefix: openClarityServerSubnetAddressPrefix
          sourcePortRange: '*'
          destinationAddressPrefix: '*'
          destinationPortRange: '22'
        }
      }
    ]
  }
}

var storageAccountName = toLower('store${uniqueString(resourceGroup().id)}')
var storageAccountType = 'Standard_LRS'
// var subnetRef = '${openClarityNet.id}/subnets/${openClarityServerSubnet.name}'

resource storageAccount 'Microsoft.Storage/storageAccounts@2023-05-01' = {
  name: storageAccountName
  location: location
  sku: {
    name: storageAccountType
  }
  kind: 'StorageV2'
  properties: {
    accessTier: 'Hot'
    networkAcls: {
      bypass: 'AzureServices'
      virtualNetworkRules: [
        {
          id: openClarityNet::openClarityServerSubnet.id
          action: 'Allow'
        }
      ]
      defaultAction: 'Deny'
    }
  }
}

resource blobService 'Microsoft.Storage/storageAccounts/blobServices@2023-05-01' = {
  parent: storageAccount
  name: 'default'
}

var snapshotContainerName = 'snapshots'

resource snapshotContainer 'Microsoft.Storage/storageAccounts/blobServices/containers@2023-05-01' = {
  parent: blobService
  name: snapshotContainerName
}

@description('This is the built-in Blob Contributor role. See https://learn.microsoft.com/en-gb/azure/role-based-access-control/built-in-roles#storage-blob-data-contributor')
resource blobContributorRoleDefinition 'Microsoft.Authorization/roleDefinitions@2022-04-01' existing = {
  scope: subscription()
  name: 'ba92f5b4-2d11-453d-a403-e96b0029c9fe'
}

resource blocContributorRoleAssignment 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  scope: storageAccount
  name: guid(storageAccount.id, openClarityIdentityID, blobContributorRoleDefinition.id)
  properties: {
    roleDefinitionId: blobContributorRoleDefinition.id
    principalId: principalID
    principalType: 'ServicePrincipal'
  }
}

output adminUsername string = adminUsername
output hostname string = publicIPAddress.properties.dnsSettings.fqdn
output sshCommand string = 'ssh ${adminUsername}@${publicIPAddress.properties.dnsSettings.fqdn}'
