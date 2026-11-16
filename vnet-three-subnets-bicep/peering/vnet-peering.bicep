// vnet-peering.bicep
// Creates two VNets (hub + spoke), each with two subnets, and peers them bidirectionally.
// Deploy: az deployment group create -g <rg> -f vnet-peering.bicep

@description('Azure region for all resources')
param location string = resourceGroup().location

@description('Name prefix used for all resources')
param namePrefix string = 'demo'

// ---------- Hub VNet ----------
resource hubVnet 'Microsoft.Network/virtualNetworks@2023-11-01' = {
  name: '${namePrefix}-hub-vnet'
  location: location
  properties: {
    addressSpace: {
      addressPrefixes: [
        '10.0.0.0/16'
      ]
    }
    subnets: [
      {
        name: 'GatewaySubnet'
        properties: {
          addressPrefix: '10.0.0.0/27'
        }
      }
      {
        name: 'shared-services'
        properties: {
          addressPrefix: '10.0.1.0/24'
        }
      }
    ]
  }
}

// ---------- Spoke VNet ----------
resource spokeVnet 'Microsoft.Network/virtualNetworks@2023-11-01' = {
  name: '${namePrefix}-spoke-vnet'
  location: location
  properties: {
    addressSpace: {
      addressPrefixes: [
        '10.1.0.0/16'
      ]
    }
    subnets: [
      {
        name: 'workload'
        properties: {
          addressPrefix: '10.1.0.0/24'
        }
      }
      {
        name: 'data'
        properties: {
          addressPrefix: '10.1.1.0/24'
        }
      }
    ]
  }
}

// ---------- Peering: hub -> spoke ----------
resource hubToSpokePeering 'Microsoft.Network/virtualNetworks/virtualNetworkPeerings@2023-11-01' = {
  parent: hubVnet
  name: 'hub-to-spoke'
  properties: {
    remoteVirtualNetwork: {
      id: spokeVnet.id
    }
    allowVirtualNetworkAccess: true
    allowForwardedTraffic: true
    allowGatewayTransit: true
    useRemoteGateways: false
  }
}

// ---------- Peering: spoke -> hub ----------
resource spokeToHubPeering 'Microsoft.Network/virtualNetworks/virtualNetworkPeerings@2023-11-01' = {
  parent: spokeVnet
  name: 'spoke-to-hub'
  properties: {
    remoteVirtualNetwork: {
      id: hubVnet.id
    }
    allowVirtualNetworkAccess: true
    allowForwardedTraffic: true
    allowGatewayTransit: false
    useRemoteGateways: false
  }
}

// ---------- Outputs ----------
output hubVnetId string = hubVnet.id
output spokeVnetId string = spokeVnet.id
output hubSubnets array = [for subnet in hubVnet.properties.subnets: subnet.name]
output spokeSubnets array = [for subnet in spokeVnet.properties.subnets: subnet.name]
