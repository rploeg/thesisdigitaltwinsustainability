# Installation and configuration of infrastructure components


# Create Azure IoT Central application

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


# Create Azure Data Explorer

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

# Create Azure Digital Twins

The following steps are used to create Azure Digital Twins app

1. Create an empty Azure Digital Twin service
2. Turn on managed idenity
3. Upload your DTDL file of the production line


# Configuration in Azure IoT Central

To export the data from Azure IoT Central to Azure Data explorer the following steps need to be made:

1. Create a new destination in Export section of IoTC
2. Use managed idenity when selection Azure Data Explorer
3. Create new export with telemetry section
4. Use the following transformation of data:

import "iotc" as iotc;
{
    messageTimestamp: .telemetry | iotc::find(.name == "messageTimestamp").value,
    deviceId: .device.id,
    plantName: .telemetry | iotc::find(.name == "plantName").value,
    productionLine: .telemetry | iotc::find(.name == "productionLine").value,
    shiftNumber: .telemetry | iotc::find(.name == "shiftNumber").value,
    batchNumber: .telemetry | iotc::find(.name == "batchNumber").value,
    totalPartsMade: .telemetry | iotc::find(.name == "totalPartsMade").value,
    defectivePartsMade: .telemetry | iotc::find(.name == "defectivePartsMade").value,
    temperature: .telemetry | iotc::find(.name == "temperature").value,
    oilLevel: .telemetry | iotc::find(.name == "oilLevel").value,
    machineHealth: .telemetry | iotc::find(.name == "machineHealth").value,
    kwh: .telemetry | iotc::find(.name == "kwh").value,
    plannedkwh: .telemetry | iotc::find(.name == "plannedkwh").value
}

5. Safe the export and see if the export is running


# Configuration of Azure Data Explorer dashboard

1. Import [this file](https://github.com/rploeg/thesisdigitaltwinsustainability/blob/main/ADX/dashboard-OEEE%20Dashboard.json) in the dashboard of Azure Data Explorer

You should now see the following dashboard (select the right factory, productionline and timeframe):

![image](https://user-images.githubusercontent.com/49752333/171480270-1a180cb5-4df9-4978-b315-7b72ecee9a9e.png)

