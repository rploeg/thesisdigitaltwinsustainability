package simulating

type Plant struct {
	Name        string `json:"name"`
	BoltMachine struct {
		Count  int    `json:"count"`
		Format string `json:"format"`
	} `json:"boltMachine"`
}
