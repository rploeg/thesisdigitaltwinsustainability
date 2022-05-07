# The transformation query specified here will be used to change each exported 
# message into a different format. You can get started using the example below,
# and learn more about the language in documentation:
# https://aka.ms/dataexporttransformation
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
    machineHealth: .telemetry | iotc::find(.name == "machineHealth").value
}