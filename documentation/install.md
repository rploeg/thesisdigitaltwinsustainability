# Installation of infrastructure components


# Azure IoT Central

The following steps are used to configure IoT Central
1. Create a custom Azure IoT Central application on https://apps.azureiotcentral.com 
2. [Import](https://github.com/rploeg/thesisdigitaltwinsustainability/blob/main/IoTC/BoltMaker.json) the boltmaker device template
4. Create the standard views in the device import screen

<b>Create app identity for IoT Central (used later for Azure Data Explorer)</b>:

remco@Azure:~$ az iot central app identity assign --name oeee \
>     --resource-group thesis \
>     --system-assigned
{
  "principalId": "YOURPRINCIPALID",
  "tenantId": "YOURTENANTID",
  "type": "SystemAssigned"
}


# Azure Data Explorer

The following steps are used for installation of Azure Data Explorer.

1. Create an ADE pool
2. Create a database called Boltmaker (or your own name)
3. Add a principal assignment to ADE with PowerShell:
az kusto database-principal-assignment create --cluster-name oeeethesis \
    --database-name YOURDATABASE    \
    --resource-group YOURRESOURCEGROUP \
    --principal-assignment-name NAME USED IN STEP OF IOT CENTRAL \
    --principal-id YOURPRINCIPALID \
    --principal-type App --role Admin \
    --tenant-id YOURTENANTID
  4. Create table in ADE
  
.create-merge table boltmaker (messageTimestamp:datetime, deviceId:string, plantName:string, productionLine:string, shiftNumber:long, batchNumber:long, totalPartsMade:long, defectivePartsMade:long, temperature:real, oilLevel:real, machineHealth:string, kwh:real, plannedkwh:real) 


