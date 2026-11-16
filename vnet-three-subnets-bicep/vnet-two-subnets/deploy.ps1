az group create --name demo_vnet-rg --location denmarkeast --tags environment=dev
az deployment group create -g demo_vnet-rg -f public-private-vm.bicep -p public-private-vm.bicepparam
# az deployment group create -g demo_vnet-rg -f public-private-vm-postgres.bicep -p public-private-vm-postgres.bicepparam