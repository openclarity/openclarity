@description('Username for the VMClarity Server VM')
param adminUsername string

@description('SSH Public Key for the VMClarity Server VM')
@secure()
param adminSSHKey string

@description('The size of the VMClarity Server VM')
param serverVmSize string = 'Standard_D2s_v3'

@description('The size of the Scanner VMs')
param scannerVmSize string = 'Standard_D2s_v3'

@description('Location where to create the resources')
param location string = resourceGroup().location

@description('Public IP DNS prefix')
param dnsLabelPrefix string = toLower('vmclarity-server-${uniqueString(resourceGroup().id)}')

@description('Security Type of the VMClartiy Server VM')
@allowed([
  'Standard'
  'TrustedLaunch'
])
param securityType string = 'TrustedLaunch'

@description('VMClarity Server Identity ID')
param vmClarityIdentityID string

@description('VMClarity Managed Identity Principal ID')
param principalID string

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

@description ('Yara Rule Server Container Image')
param yaraRuleServerContainerImage string = 'ghcr.io/openclarity/yara-rule-server:v0.1.0'

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
param databaseToUse string = 'Postgresql'

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

var vmClarityNetName = 'vmclarity-server-net'
var addressPrefix = '10.1.0.0/16'

var vmClarityServerSubnetName = 'vmclarity-server-subnet'
var vmClarityServerSecurityGroupName = 'vmclarity-server-security-group'
var vmClarityServerSubnetAddressPrefix = '10.1.0.0/24'

var vmClarityScannerSubnetName = 'vmclarity-scanner-subnet'
var vmClarityScannerSecurityGroupName = 'vmclarity-scanner-security-group'
var vmClarityScannerSubnetAddressPrefix = '10.1.1.0/24'

var vmclarityServerVMName = 'vmclarity-server'
var publicIPAddressName = '${vmclarityServerVMName}-public-ip'
var networkInterfaceName = '${vmclarityServerVMName}-net-int'

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
  ScannerInstanceType: scannerVmSize
  AssetScanDeletePolicy: assetScanDeletePolicy
  DatabaseToUse: databaseToUse
  PostgresDBPassword: postgresDBPassword
  ExternalDBHost: externalDBHost
  ExternalDBPort: string(externalDBPort)
  ExternalDBName: externalDBName
  ExternalDBUsername: externalDBUsername
  ExternalDBPassword: externalDBPassword
  AZURE_SUBSCRIPTION_ID: subscription().subscriptionId
  AZURE_SCANNER_LOCATION: location
  AZURE_SCANNER_RESOURCE_GROUP: resourceGroup().name
  AZURE_SCANNER_SUBNET_ID: vmClarityNet::vmClarityScannerSubnet.id
  AZURE_SCANNER_PUBLIC_KEY: base64(adminSSHKey)
  AZURE_SCANNER_VM_SIZE: scannerVmSize
  AZURE_SCANNER_IMAGE_PUBLISHER: imageReference.publisher
  AZURE_SCANNER_IMAGE_OFFER: imageReference.offer
  AZURE_SCANNER_IMAGE_SKU: imageReference.sku
  AZURE_SCANNER_IMAGE_VERSION: imageReference.version
  AZURE_SCANNER_SECURITY_GROUP: vmClarityScannerSecurityGroup.id
  AZURE_SCANNER_STORAGE_ACCOUNT_NAME: storageAccountName
  AZURE_SCANNER_STORAGE_CONTAINER_NAME: snapshotContainerName
}

var scriptTemplate = loadTextContent('vmclarity-install.sh')

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

var vmClarityGuestAttestationName = 'VmClarityServerGuestAttestation'
var extensionName = 'GuestAttestation'
var extensionPublisher = 'Microsoft.Azure.Security.LinuxAttestation'
var extensionVersion = '1.0'
var maaTenantName = 'GuestAttestation'
var maaEndpoint = substring('emptystring', 0, 0)

var vmClarityServerCustomScriptName = 'VmClarityServerCustomScript'

resource networkInterface 'Microsoft.Network/networkInterfaces@2021-05-01' = {
  name: networkInterfaceName
  location: location
  properties: {
    ipConfigurations: [
      {
        name: 'ipconfig1'
        properties: {
          subnet: {
            id: vmClarityNet::vmClarityServerSubnet.id
          }
          privateIPAllocationMethod: 'Dynamic'
          publicIPAddress: {
            id: publicIPAddress.id
          }
        }
      }
    ]
    networkSecurityGroup: {
      id: vmClarityServerSecurityGroup.id
    }
  }
}

resource vmClarityServerSecurityGroup 'Microsoft.Network/networkSecurityGroups@2021-05-01' = {
  name: vmClarityServerSecurityGroupName
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
resource vmClarityNet 'Microsoft.Network/virtualNetworks@2021-05-01' = {
  name: vmClarityNetName
  location: location
  properties: {
    addressSpace: {
      addressPrefixes: [
        addressPrefix
      ]
    }
    subnets:[
      {
        name: vmClarityServerSubnetName
        properties: {
          addressPrefix: vmClarityServerSubnetAddressPrefix
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
        name: vmClarityScannerSubnetName
        properties: {
          addressPrefix: vmClarityScannerSubnetAddressPrefix
          privateEndpointNetworkPolicies: 'Enabled'
          privateLinkServiceNetworkPolicies: 'Enabled'
        }
      }
    ]
  }

  resource vmClarityServerSubnet 'subnets' existing = {
    name: vmClarityServerSubnetName
  }

  resource vmClarityScannerSubnet 'subnets' existing = {
    name: vmClarityScannerSubnetName
  }
}

resource publicIPAddress 'Microsoft.Network/publicIPAddresses@2021-05-01' = {
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

resource vmClarityServer 'Microsoft.Compute/virtualMachines@2021-11-01' = {
  name: vmclarityServerVMName
  location: location
  identity: {
    type: 'UserAssigned'
    userAssignedIdentities:{
      '${vmClarityIdentityID}': {}
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
      computerName: vmclarityServerVMName
      adminUsername: adminUsername
      linuxConfiguration: linuxConfiguration
    }
    securityProfile: ((securityType == 'TrustedLaunch') ? securityProfileJson : null)
  }
}

resource vmclarityServerGuestAttestation 'Microsoft.Compute/virtualMachines/extensions@2022-03-01' = if ((securityType == 'TrustedLaunch') && ((securityProfileJson.uefiSettings.secureBootEnabled == true) && (securityProfileJson.uefiSettings.vTpmEnabled == true))) {
  parent: vmClarityServer
  name: vmClarityGuestAttestationName
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

resource vmclarityServerCustomScript 'Microsoft.Compute/virtualMachines/extensions@2023-03-01' = {
  parent: vmClarityServer
  dependsOn: [
    vmclarityServerGuestAttestation
  ]
  name: vmClarityServerCustomScriptName
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

resource vmClarityScannerSecurityGroup 'Microsoft.Network/networkSecurityGroups@2021-05-01' = {
  name: vmClarityScannerSecurityGroupName
  location: location
  properties: {
    securityRules: [
      {
        name: 'SSH-From-VMClarity-Server'
        properties: {
          priority: 1000
          protocol: 'Tcp'
          access: 'Allow'
          direction: 'Inbound'
          sourceAddressPrefix: vmClarityServerSubnetAddressPrefix
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
// var subnetRef = '${vmClarityNet.id}/subnets/${vmClarityServerSubnet.name}'

resource storageAccount 'Microsoft.Storage/storageAccounts@2022-09-01' = {
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
          id: vmClarityNet::vmClarityServerSubnet.id
          action: 'Allow'
        }
      ]
      defaultAction: 'Deny'
    }
  }
}

resource blobService 'Microsoft.Storage/storageAccounts/blobServices@2021-02-01' = {
  parent: storageAccount
  name: 'default'
}

var snapshotContainerName = 'snapshots'

resource snapshotContainer 'Microsoft.Storage/storageAccounts/blobServices/containers@2022-09-01' = {
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
  name: guid(storageAccount.id, vmClarityIdentityID, blobContributorRoleDefinition.id)
  properties: {
    roleDefinitionId: blobContributorRoleDefinition.id
    principalId: principalID
    principalType: 'ServicePrincipal'
  }
}

output adminUsername string = adminUsername
output hostname string = publicIPAddress.properties.dnsSettings.fqdn
output sshCommand string = 'ssh ${adminUsername}@${publicIPAddress.properties.dnsSettings.fqdn}'
