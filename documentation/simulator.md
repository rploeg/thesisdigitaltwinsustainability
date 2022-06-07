# Configurator simulator

The following steps are used to configure the simulator to send data to Azure IoT Central.

1. Install the simulator [code](https://github.com/rploeg/thesisdigitaltwinsustainability/tree/main/Simulator) on your machine. You need the language Go for this. Please install Visual Studio code where Go can be integrated
2. Configure the iotoee.json file with the information from Azure IoT Central
<code>
  {
    "logger": {
      "logLevel": "Debug",
      "logsDir": "./logs"
    },
    "application": {
      "provisioningUrl": "global.azure-devices-provisioning.net",
      "idScope": "YOURIDSCOPE
      "masterKey": "YOURMASTERKEY"
      "boltMachineModelID": "dtmi:parnellAerospace:BoltMakerV1;1"
    },
    "plant": [
      {
        "name": "FoodFactory",
        "boltMachine":{
          "count": 4,
          "format": "json"
        }
      }
    ]
  }
  </code>
3. Compile your code
4. Start the .exe file on your machine
5. You should see this <br>

![image](https://user-images.githubusercontent.com/49752333/171599352-8fcb4638-454e-4cdf-9b0d-41998bce7d69.png)



# Datapoints that simulator sends to Azure IoT Central

The following datapoints are send to to the sustainable digital twin solution

<code>
  
	BoltMachineTelemetryMessage struct {
		PlantName          string    `json:"plantName"`
		ProductionLine     string    `json:"productionLine"`
		ShiftNumber        int       `json:"shiftNumber"`
		BatchNumber        int       `json:"batchNumber"`
		MessageTimestamp   time.Time `json:"messageTimestamp"`
		TotalPartsMade     int       `json:"totalPartsMade"`
		DefectivePartsMade int       `json:"defectivePartsMade"`
		MachineHealth      string    `json:"machineHealth"`
		OilLevel           float64   `json:"oilLevel"`
		Temperature        float64   `json:"temperature"`
		Kwh		           float64   `json:"kwh"`
		PlannedKwH		   float64	 `json:"plannedkwh"`
	}
  </code>
  

# Import data

Also a dataset is provided that you can import into Azure Data Explorer. The simulated data set can be found in the [ADX directory](https://github.com/rploeg/thesisdigitaltwinsustainability/ADX/). 