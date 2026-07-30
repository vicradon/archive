// public-private-vm.bicep
// VNet with two subnets: subnet1 (public, SSH from internet) and subnet2 (private, SSH only from subnet1).
// Deploy: az deployment group create -g <rg> -f public-private-vm.bicep -p public-private-vm.bicepparam

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

var vnetAddressSpace = '10.0.0.0/16'
var subnet1Prefix = '10.0.1.0/24'
var subnet2Prefix = '10.0.2.0/24'

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
    ]
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
resource nic1  'Microsoft.Network/networkInterfaces@2023-11-01' = {
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

// ---------- Outputs ----------
output vm1PublicIp string = publicIp.properties.ipAddress
output vm1PrivateIp string = nic1.properties.ipConfigurations[0].properties.privateIPAddress
output vm2PrivateIp string = nic2.properties.ipConfigurations[0].properties.privateIPAddress
