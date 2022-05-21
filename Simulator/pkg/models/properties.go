package models

import "time"

type (
	Properties struct {
		IsMachineOn        bool `json:"isMachineOn"`
		TelemetryFrequency int  `json:"telemetryFrequency"`
		ShiftDurationHours int  `json:"shiftDurationHours"`
		BatchDurationHours int  `json:"batchDurationHours"`
	}

	ReportedProperties struct {
		HostName  string    `json:"hostName"`
		IPAddress string    `json:"ipAddress"`
		HostTime  time.Time `json:"hostTime"`
	}

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
	}

	BoltMachine struct {
		PlantName          string  `json:"plantName"`
		ProductionLine     string  `json:"productionLine"`
		ShiftNumber        int     `json:"shiftNumber"`
		BatchNumber        int     `json:"batchNumber"`
		TotalPartsMade     int     `json:"totalPartsMade"`
		DefectivePartsMade int     `json:"defectivePartsMade"`
		MachineHealth      string  `json:"machineHealth"`
		OilLevel           float64 `json:"oilLevel"`
		Temperature        float64 `json:"temperature"`
		Kwh 	           float64 `json:"kwh"`
		Format             string  `json:"format"`
	}
)
