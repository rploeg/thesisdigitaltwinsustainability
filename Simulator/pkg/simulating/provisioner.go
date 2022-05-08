package simulating

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/iot-for-all/iiot-oee/pkg/models"
	"github.com/iot-for-all/iiot-oee/pkg/util"
	"github.com/rs/zerolog/log"
)

type (
	// ProvisioningRequest represents a request to provision a device in IoT Central.
	ProvisioningRequest struct {
		DeviceID    string                     // device deviceID to register the device as.
		Context     context.Context            // context of the provision request.
		Application *models.CentralApplication // simulation that is requesting the device provision.
		ModelID     string                     // device model id
	}

	// ProvisioningResponse represents the response for a given provision request.
	ProvisioningResponse struct {
		*ProvisioningRequest        // the request to which this response is generated.
		ConnectionString     string // Result of the provision request.
	}

	// DeviceProvisioner responsible for provisioning devices via DPS.
	DeviceProvisioner struct {
		context context.Context            // the context of the provisioner.
		app     *models.CentralApplication // simulating configuration.
		client  *http.Client               // http client used to interact with DPS
	}

	// registrationRequest is the registration request sent to DPS
	registrationRequest struct {
		RegistrationID string                 `json:"registrationId"`
		Payload        map[string]interface{} `json:"payload"`
	}

	// registrationResponse is the registration response sent by DPS
	registrationResponse struct {
		OperationID string `json:"operationId"`
		Status      string `json:"status"`
	}

	// registrationResult is the result of the registration request
	registrationResult struct {
		RegistrationState struct {
			AssignedHub string `json:"assignedHub"`
			DeviceID    string `json:"deviceId"`
			Status      string `json:"status"`
		} `json:"registrationState"`
	}
)

// NewProvisioner creates a new deviceProvisioner.
func NewProvisioner(ctx context.Context, app *models.CentralApplication) *DeviceProvisioner {
	p := DeviceProvisioner{
		context: ctx,
		app:     app,
		client: &http.Client{
			Timeout: time.Duration(10000) * time.Millisecond,
		},
	}

	return &p
}

// Provision provisions a device in IoT Central
func (p *DeviceProvisioner) Provision(req *ProvisioningRequest) *ProvisioningResponse {
	log.Trace().Str("deviceID", req.DeviceID).Msg("provisioning device")

	key, err := util.ComputeHmac(req.Application.MasterKey, req.DeviceID)
	if err != nil {
		log.Error().Err(err).Str("deviceId", req.DeviceID).Msg("failed to compute device key for device")
		return nil
	}

	keyRes := fmt.Sprintf("%s/registrations/%s", req.Application.IDScope, req.DeviceID)
	token, err := util.CreateSasToken(key, keyRes, "registration", 1*time.Minute)
	if err != nil {
		log.Error().Err(err).Str("deviceId", req.DeviceID).Msg("failed to compute sas key for device")
		return nil
	}

	start := time.Now()

	opdID, err := p.sendRegisterRequest(
		req.Application.ProvisioningURL,
		req.Application.IDScope,
		req.DeviceID,
		req.ModelID,
		token)

	if err != nil {
		log.Error().Err(err).Str("deviceId", req.DeviceID).Msg("failed to register device")
		return nil
	}

	log.Trace().Str("deviceID", req.DeviceID).Msg("checking registration status")

	reg, err := p.getRegistrationStatus(
		req.Application.ProvisioningURL,
		req.Application.IDScope,
		req.DeviceID,
		opdID, token)

	if err != nil {
		log.Error().Err(err).Str("deviceId", req.DeviceID).Msg("failed to get device registration result")
		return nil
	}

	connStr := fmt.Sprintf("HostName=%s;DeviceId=%s;SharedAccessKey=%s",
		reg.RegistrationState.AssignedHub,
		req.DeviceID,
		key)

	now := time.Now()
	latency := float64(now.UnixNano()-start.UnixNano()) / float64(time.Second)

	log.Trace().Str("deviceID", req.DeviceID).Float64("latency", latency).Msg("Got provisioning response")
	return &ProvisioningResponse{
		ProvisioningRequest: req,
		ConnectionString:    connStr,
	}
}

// sendRegisterRequest sends the registration request to DPS for registering the device
// host is the target DPS host to send the request to.
// scopeID is the DPS scope to register the device with.
// deviceID is the id of the device to register.
// modelID is the id of the model to register the device as.
// token is the shared access token used for authorization.
func (p *DeviceProvisioner) sendRegisterRequest(
	host string,
	idScope string,
	deviceID string,
	modelID string,
	token string) (string, error) {
	// todo: handle error conditions
	// todo: handle retry
	path := fmt.Sprintf("https://%s/%s/registrations/%s/register?api-version=2019-03-31", host, idScope, deviceID)

	reqData, err := json.Marshal(registrationRequest{
		RegistrationID: deviceID,
		Payload: map[string]interface{}{
			"modelId": modelID,
		},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(p.context, "PUT", path, bytes.NewReader(reqData))
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Encoding", "utf-8")
	req.Header.Add("Authorization", token)

	res, err := p.client.Do(req)
	if err != nil {
		return "", err
	}

	var resData registrationResponse
	err = json.NewDecoder(res.Body).Decode(&resData)
	if err != nil {
		return "", err
	}

	return resData.OperationID, nil
}

// getRegistrationStatus get the registration status of a registration request
func (p *DeviceProvisioner) getRegistrationStatus(
	host string,
	idScope string,
	deviceID string,
	operationID string,
	token string) (*registrationResult, error) {

	path := fmt.Sprintf("https://%s/%s/registrations/%s/operations/%s?api-version=2019-03-31", host, idScope, deviceID, operationID)
	req, err := http.NewRequestWithContext(p.context, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Encoding", "utf-8")
	req.Header.Add("Authorization", token)

	maxRegistrationAttempts := 3
	for i := 1; i <= maxRegistrationAttempts; i++ {
		select {
		case <-p.context.Done():
			return nil, fmt.Errorf("operation cancelled")
		default:
			res, err := p.client.Do(req)
			if err != nil {
				return nil, err
			}

			if res.StatusCode == http.StatusAccepted {
				backoff, err := strconv.Atoi(res.Header.Get("Retry-After"))
				if err != nil {
					backoff = 3
				}

				log.Trace().Str("deviceID", deviceID).Int("backoff", backoff).Int("attempt", i).Msg("Registration status")
				time.Sleep(time.Duration(backoff) * time.Second)
				continue
			}

			var resData registrationResult
			err = json.NewDecoder(res.Body).Decode(&resData)
			if err != nil {
				return nil, err
			}

			return &resData, nil
		}
	}

	return nil, fmt.Errorf("failed to evaluate registration status. all retry attempts failed")
}
