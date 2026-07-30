using 'public-private-vm.bicep'

param location = 'denmarkeast'
param namePrefix = 'demo'
param adminUsername = 'azureuser'
param adminPublicKey = 'ssh-ed25519 ...'
param vmSize = 'Standard_B1s'
