package models

type (
	CentralApplication struct {
		ProvisioningURL    string `json:"provisioningUrl"`    // DPS provisioning URL.
		IDScope            string `json:"idScope"`            // the id scope of the provisioning endpoint.
		MasterKey          string `json:"masterKey"`          // the master SAS key of the provisioning endpoint.
		BoltMachineModelID string `json:"boltMachineModelID"` // the bolt machine device model ID.
	}
)
