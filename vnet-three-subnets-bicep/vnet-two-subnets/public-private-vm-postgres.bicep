// public-private-vm-postgres.bicep
// VNet with three subnets:
//   subnet1-public   -> VM1 (public IP, SSH from internet)
//   subnet2-private   -> VM2 (private IP only, SSH only from subnet1)
//   subnet3-private   -> Azure DB for PostgreSQL Flexible Server (delegated subnet,
//                         reachable only from subnet2 on 5432)
// Deploy: az deployment group create -g <rg> -f public-private-vm-postgres.bicep -p public-private-vm-postgres.bicepparam

@description('Azure region for all resources')
param location string = resourceGroup().location

@description('Name prefix used for all resources')
param namePrefix string = 'demo'

@description('Admin username for both VMs')
param adminUsername string

@description('SSH public key for both VMs')
param adminPublicKey string

@description('VM size for both VMs')
param vmSize string = 'Standard_B1s'

@description('PostgreSQL admin username')
param pgAdminUsername string

@secure()
@description('PostgreSQL admin password')
param pgAdminPassword string

@description('PostgreSQL Flexible Server SKU (Burstable B1ms is the cheapest usable tier)')
param pgSkuName string = 'Standard_B1ms'

@description('PostgreSQL version')
param pgVersion string = '16'

var vnetAddressSpace = '10.0.0.0/16'
var subnet1Prefix = '10.0.1.0/24'
var subnet2Prefix = '10.0.2.0/24'
var subnet3Prefix = '10.0.3.0/24'

// ---------- NSGs ----------
resource nsgPublic 'Microsoft.Network/networkSecurityGroups@2023-11-01' = {
  name: '${namePrefix}-subnet1-nsg'
  location: location
  properties: {
    securityRules: [
      {
        name: 'AllowSSHFromInternet'
        properties: {
          priority: 100
          direction: 'Inbound'
          access: 'Allow'
          protocol: 'Tcp'
          sourceAddressPrefix: 'Internet'
          sourcePortRange: '*'
          destinationAddressPrefix: '*'
          destinationPortRange: '22'
        }
      }
    ]
  }
}

resource nsgPrivate 'Microsoft.Network/networkSecurityGroups@2023-11-01' = {
  name: '${namePrefix}-subnet2-nsg'
  location: location
  properties: {
    securityRules: [
      {
        name: 'AllowSSHFromSubnet1Only'
        properties: {
          priority: 100
          direction: 'Inbound'
          access: 'Allow'
          protocol: 'Tcp'
          sourceAddressPrefix: subnet1Prefix
          sourcePortRange: '*'
          destinationAddressPrefix: '*'
          destinationPortRange: '22'
        }
      }
      {
        name: 'DenyAllOtherInbound'
        properties: {
          priority: 4096
          direction: 'Inbound'
          access: 'Deny'
          protocol: '*'
          sourceAddressPrefix: '*'
          sourcePortRange: '*'
          destinationAddressPrefix: '*'
          destinationPortRange: '*'
        }
      }
    ]
  }
}

resource nsgData 'Microsoft.Network/networkSecurityGroups@2023-11-01' = {
  name: '${namePrefix}-subnet3-nsg'
  location: location
  properties: {
    securityRules: [
      {
        name: 'AllowPostgresFromSubnet2Only'
        properties: {
          priority: 100
          direction: 'Inbound'
          access: 'Allow'
          protocol: 'Tcp'
          sourceAddressPrefix: subnet2Prefix
          sourcePortRange: '*'
          destinationAddressPrefix: '*'
          destinationPortRange: '5432'
        }
      }
      {
        name: 'DenyAllOtherInbound'
        properties: {
          priority: 4096
          direction: 'Inbound'
          access: 'Deny'
          protocol: '*'
          sourceAddressPrefix: '*'
          sourcePortRange: '*'
          destinationAddressPrefix: '*'
          destinationPortRange: '*'
        }
      }
    ]
  }
}

// ---------- VNet + Subnets ----------
resource vnet 'Microsoft.Network/virtualNetworks@2023-11-01' = {
  name: '${namePrefix}-vnet'
  location: location
  properties: {
    addressSpace: {
      addressPrefixes: [
        vnetAddressSpace
      ]
    }
    subnets: [
      {
        name: 'subnet1-public'
        properties: {
          addressPrefix: subnet1Prefix
          networkSecurityGroup: {
            id: nsgPublic.id
          }
        }
      }
      {
        name: 'subnet2-private'
        properties: {
          addressPrefix: subnet2Prefix
          networkSecurityGroup: {
            id: nsgPrivate.id
          }
        }
      }
      {
        name: 'subnet3-private'
        properties: {
          addressPrefix: subnet3Prefix
          networkSecurityGroup: {
            id: nsgData.id
          }
          delegations: [
            {
              name: 'postgresDelegation'
              properties: {
                serviceName: 'Microsoft.DBforPostgreSQL/flexibleServers'
              }
            }
          ]
        }
      }
    ]
  }
}

// ---------- Private DNS zone for Postgres ----------
resource pgDnsZone 'Microsoft.Network/privateDnsZones@2024-06-01' = {
  name: '${namePrefix}.private.postgres.database.azure.com'
  location: 'global'
}

resource pgDnsZoneLink 'Microsoft.Network/privateDnsZones/virtualNetworkLinks@2024-06-01' = {
  parent: pgDnsZone
  name: '${namePrefix}-vnet-link'
  location: 'global'
  properties: {
    virtualNetwork: {
      id: vnet.id
    }
    registrationEnabled: false
  }
}

// ---------- Public IP for VM1 ----------
resource publicIp 'Microsoft.Network/publicIPAddresses@2023-11-01' = {
  name: '${namePrefix}-vm1-pip'
  location: location
  sku: {
    name: 'Standard'
  }
  properties: {
    publicIPAllocationMethod: 'Static'
  }
}

// ---------- NICs ----------
resource nic1 'Microsoft.Network/networkInterfaces@2023-11-01' = {
  name: '${namePrefix}-vm1-nic'
  location: location
  properties: {
    ipConfigurations: [
      {
        name: 'ipconfig1'
        properties: {
          subnet: {
            id: vnet.properties.subnets[0].id
          }
          privateIPAllocationMethod: 'Dynamic'
          publicIPAddress: {
            id: publicIp.id
          }
        }
      }
    ]
  }
}

resource nic2 'Microsoft.Network/networkInterfaces@2023-11-01' = {
  name: '${namePrefix}-vm2-nic'
  location: location
  properties: {
    ipConfigurations: [
      {
        name: 'ipconfig1'
        properties: {
          subnet: {
            id: vnet.properties.subnets[1].id
          }
          privateIPAllocationMethod: 'Dynamic'
        }
      }
    ]
  }
}

// ---------- VM1 (public subnet) ----------
resource vm1 'Microsoft.Compute/virtualMachines@2024-07-01' = {
  name: '${namePrefix}-vm1'
  location: location
  properties: {
    hardwareProfile: {
      vmSize: vmSize
    }
    osProfile: {
      computerName: 'vm1'
      adminUsername: adminUsername
      linuxConfiguration: {
        disablePasswordAuthentication: true
        ssh: {
          publicKeys: [
            {
              path: '/home/${adminUsername}/.ssh/authorized_keys'
              keyData: adminPublicKey
            }
          ]
        }
      }
    }
    storageProfile: {
      imageReference: {
        publisher: 'Canonical'
        offer: 'ubuntu-24_04-lts'
        sku: 'server'
        version: 'latest'
      }
      osDisk: {
        createOption: 'FromImage'
        managedDisk: {
          storageAccountType: 'Standard_LRS'
        }
      }
    }
    networkProfile: {
      networkInterfaces: [
        {
          id: nic1.id
        }
      ]
    }
  }
}

// ---------- VM2 (private subnet) ----------
resource vm2 'Microsoft.Compute/virtualMachines@2024-07-01' = {
  name: '${namePrefix}-vm2'
  location: location
  properties: {
    hardwareProfile: {
      vmSize: vmSize
    }
    osProfile: {
      computerName: 'vm2'
      adminUsername: adminUsername
      linuxConfiguration: {
        disablePasswordAuthentication: true
        ssh: {
          publicKeys: [
            {
              path: '/home/${adminUsername}/.ssh/authorized_keys'
              keyData: adminPublicKey
            }
          ]
        }
      }
    }
    storageProfile: {
      imageReference: {
        publisher: 'Canonical'
        offer: 'ubuntu-24_04-lts'
        sku: 'server'
        version: 'latest'
      }
      osDisk: {
        createOption: 'FromImage'
        managedDisk: {
          storageAccountType: 'Standard_LRS'
        }
      }
    }
    networkProfile: {
      networkInterfaces: [
        {
          id: nic2.id
        }
      ]
    }
  }
}

// ---------- Azure Database for PostgreSQL Flexible Server ----------
resource postgres 'Microsoft.DBforPostgreSQL/flexibleServers@2024-08-01' = {
  name: '${namePrefix}-pg-${uniqueString(resourceGroup().id)}'
  location: location
  sku: {
    name: pgSkuName
    tier: 'Burstable'
  }
  properties: {
    version: pgVersion
    administratorLogin: pgAdminUsername
    administratorLoginPassword: pgAdminPassword
    network: {
      delegatedSubnetResourceId: vnet.properties.subnets[2].id
      privateDnsZoneArmResourceId: pgDnsZone.id
    }
    storage: {
      storageSizeGB: 32
    }
    backup: {
      backupRetentionDays: 7
      geoRedundantBackup: 'Disabled'
    }
    highAvailability: {
      mode: 'Disabled'
    }
  }
  dependsOn: [
    pgDnsZoneLink
  ]
}

// ---------- Outputs (grouped into one object for clean formatting) ----------
output connectionInfo object = {
  vm1PublicIp: publicIp.properties.ipAddress
  vm1PrivateIp: nic1.properties.ipConfigurations[0].properties.privateIPAddress
  vm2PrivateIp: nic2.properties.ipConfigurations[0].properties.privateIPAddress
  postgresFqdn: postgres.properties.fullyQualifiedDomainName
  postgresAdminUsername: pgAdminUsername
}
