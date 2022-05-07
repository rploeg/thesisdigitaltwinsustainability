package simulating

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/amenzhinsky/iothub/iotdevice"
	iotmqtt "github.com/amenzhinsky/iothub/iotdevice/transport/mqtt"
	"github.com/amenzhinsky/iothub/logger"
	"github.com/hashicorp/go-uuid"
	"github.com/iot-for-all/iiot-oee/pkg/models"
	"github.com/rs/zerolog/log"
)

type (
	centralDevice struct {
		deviceID                    string // unique id of the device.
		app                         *models.CentralApplication
		context                     context.Context
		cancel                      context.CancelFunc
		isMachineOn                 bool                    // Twin property - is the machine on so that the device can send telemetry
		telemetryFrequency          int                     // Twin property - how often this device should send telemetry
		reportedPropertiesFrequency int                     // Twin property - how often this device should send reported properties
		shiftDurationHours          int                     // Twin property - how many hours are there in an employee shift
		batchDurationHours          int                     // Twin property - how many hours are there in batch
		boltMachine                 *models.BoltMachine     // bolt machine state
		provisioner                 *DeviceProvisioner      // provisioner used to provision the device in DPS
		connectionString            string                  // IoT Hub connectionString of the device.
		isConnected                 bool                    // is the device connected.
		isConnecting                bool                    // is the device connecting now.
		sendingTelemetry            bool                    // is the device sending telemetry now.
		sendingReportedProps        bool                    // is the device sending reported properties now.
		iotHubClient                *iotdevice.Client       // IoT Hub connection MQTT client.
		twinSub                     *iotdevice.TwinStateSub // subscription to listen for twin updates.
		c2dSub                      *iotdevice.EventSub     // subscription to listen for c2d commands
		retryCount                  int                     // number of retries for sending telemetry
		telemetrySequenceNumber     int                     // telemetry message number since process startup
		telemetryWaitContext        context.Context
		telemetryWaitCancel         context.CancelFunc
		reportedWaitContext         context.Context
		reportedWaitCancel          context.CancelFunc
	}
)

func NewDevice(ctx context.Context, app *models.CentralApplication, deviceID string,
	boltMachine *models.BoltMachine) *centralDevice {
	deviceCtx, cancel := context.WithCancel(ctx)
	twCtx, twCancel := context.WithCancel(deviceCtx)
	rwCtx, rwCancel := context.WithCancel(deviceCtx)

	return &centralDevice{
		deviceID:                    deviceID,
		app:                         app,
		context:                     deviceCtx,
		cancel:                      cancel,
		isMachineOn:                 true,
		telemetryFrequency:          60,
		reportedPropertiesFrequency: 60 * 60 * 2,
		shiftDurationHours:          8,
		batchDurationHours:          1,
		boltMachine:                 boltMachine,
		provisioner:                 NewProvisioner(deviceCtx, app),
		connectionString:            "",
		isConnected:                 false,
		isConnecting:                false,
		sendingTelemetry:            false,
		sendingReportedProps:        false,
		iotHubClient:                nil,
		twinSub:                     nil,
		c2dSub:                      nil,
		retryCount:                  0,
		telemetryWaitContext:        twCtx,
		telemetryWaitCancel:         twCancel,
		reportedWaitContext:         rwCtx,
		reportedWaitCancel:          rwCancel,
	}
}

func (d *centralDevice) Start() {
	// give some time for starling to startup and settle down
	log.Debug().Str("deviceID", d.deviceID).Msg("starting device")
	select {
	case <-d.context.Done():
		return
	default:
		time.Sleep(10 * time.Second)
	}

	log.Debug().Str("deviceID", d.deviceID).Msg("provisioning device")
	if d.connectDevice() {
		// telemetry and reported property pumps are started after applying twin states
		d.applyInitialTwinState()
	}
}

func (d *centralDevice) Stop() {
	d.cancel()
}

func (d *centralDevice) startTelemetryPump() {
	log.Debug().Str("deviceID", d.deviceID).Msg("telemetry generator pump starting")
	for {
		select {
		case <-d.context.Done():
			return
		default:
			// send telemetry
			log.Trace().Str("deviceID", d.deviceID).Msg("send telemetry")

			// send telemetry only if the machine is ON
			if d.isMachineOn {
				telemetry, err := d.getTelemetryMessage()
				if err != nil {
					log.Error().Err(err).Str("deviceID", d.deviceID).Msg("error preparing telemetry from host")
				} else {
					if d.sendTelemetryMessage(telemetry) {
						log.Debug().Str("payload", string(telemetry)).Msg("sent telemetry")
					}
				}
			} else {
				log.Debug().Str("deviceID", d.deviceID).Msg("ignoring telemetry as the machine is OFF")
			}

			// sleep for some time between each telemetry sends
			select {
			case <-d.telemetryWaitContext.Done():
				return
			case <-time.After(time.Second * time.Duration(d.telemetryFrequency)):
			}
		}
	}
}

func (d *centralDevice) startReportedPropsPump() {
	log.Debug().Msg("reported properties update pump starting")
	for {
		select {
		case <-d.context.Done():
			return
		default:
			// send reported properties
			log.Debug().Msg("send reported properties")
			props, err := d.getReportedProperties()
			if err != nil {
				log.Error().Err(err).Msg("error preparing reported properties from host")
			} else {
				d.sendReportedProperties(props)
			}

			// sleep for some time between each reported property sends
			select {
			case <-d.reportedWaitContext.Done():
				return
			case <-time.After(time.Second * time.Duration(d.reportedPropertiesFrequency)):
			}
		}
	}
}

func (d *centralDevice) getTelemetryMessage() ([]byte, error) {

	now := time.Now().UTC()
	shiftNumber := now.Hour() / d.shiftDurationHours
	if now.After(time.Date(now.Year(), now.Month(), now.Day(), shiftNumber*d.shiftDurationHours, 0, 0, 0, time.UTC)) {
		shiftNumber++
	}

	batchNumber := now.Hour() / d.batchDurationHours
	if now.After(time.Date(now.Year(), now.Month(), now.Day(), batchNumber*d.batchDurationHours, 0, 0, 0, time.UTC)) {
		batchNumber++
	}

	totalPartsMade := 90 + rand.Intn(10)
	defectivePartsMade := 0
	if rand.Intn(100) > 80 {
		defectivePartsMade = rand.Intn(10)
	}

	d.boltMachine.OilLevel -= 0.1
	if d.boltMachine.OilLevel <= 0.0 {
		d.boltMachine.OilLevel = 100.0
	}

	if d.boltMachine.OilLevel < 10.0 {
		d.boltMachine.MachineHealth = "Error"
		totalPartsMade = 0
		defectivePartsMade = 0
	} else if d.boltMachine.OilLevel < 25.0 {
		d.boltMachine.MachineHealth = "Warning"
		totalPartsMade -= 50
	} else {
		d.boltMachine.MachineHealth = "Healthy"
	}

	if d.boltMachine.Temperature >= 90 {
		d.boltMachine.Temperature -= 0.5
	} else if d.boltMachine.Temperature <= 50 {
		d.boltMachine.Temperature += 0.5
	} else {
		if rand.Intn(100) > 50 {
			d.boltMachine.Temperature += 0.5
		} else {
			d.boltMachine.Temperature -= 0.5
		}
	}
//added Remco
	if d.boltMachine.Kwh >= 95 {
		d.boltMachine.Kwh -= 0.5
	} else if d.boltMachine.Kwh <= 40 {
		d.boltMachine.Kwh += 0.5
	} else {
		if rand.Intn(100) > 40 {
			d.boltMachine.Kwh += 0.5
		} else {
			d.boltMachine.Kwh -= 0.5
		}
	}

	telemetry := models.BoltMachineTelemetryMessage{
		PlantName:          d.boltMachine.PlantName,
		ProductionLine:     d.boltMachine.ProductionLine,
		ShiftNumber:        shiftNumber,
		BatchNumber:        batchNumber,
		MessageTimestamp:   time.Now().UTC(),
		TotalPartsMade:     totalPartsMade,
		DefectivePartsMade: defectivePartsMade,
		MachineHealth:      d.boltMachine.MachineHealth,
		OilLevel:           d.boltMachine.OilLevel,
		Temperature:        d.boltMachine.Temperature,
		Kwh:       		    d.boltMachine.Kwh,
	}

	return d.getBoltTelemetryPayload(&telemetry, d.boltMachine.Format == "opcua")
}

func (d *centralDevice) sendTelemetryMessage(body []byte) bool {
	// if the device is in the middle of sending a telemetry, skip this request
	if d.sendingTelemetry {
		log.Trace().
			Str("deviceID", d.deviceID).
			Msg("skipping telemetry as it is already sending one")
		return false
	}

	d.sendingTelemetry = true

	// if there are too many retries, device might have disconnected or failed over; provision it again
	failureDetected := false
	if d.retryCount > 1 {
		// && req.device.isConnected == true {
		d.disconnectDevice()
		d.connectionString = ""
		// clear device from cache
		log.Debug().Str("deviceID", d.deviceID).Int("retryCount", d.retryCount).Msg("device might have been moved so will be re-provisioned")
		failureDetected = true
	}

	// make sure that the device is connected
	if d.isConnected == false {
		if d.connectDevice() == false {
			d.sendingTelemetry = false
			return false
		}

		// device failed over successfully
		if failureDetected {
			log.Debug().Str("deviceID", d.deviceID).Msg("device failed over successfully")
		}
	}

	// send telemetry to IoT Central
	log.Trace().Str("payload", string(body)).Int("size", len(body)).Msg("about to send telemetry message")
	correlationID, _ := uuid.GenerateUUID()
	messageID, _ := uuid.GenerateUUID()
	timeoutCtx, cancel := context.WithTimeout(d.context, time.Millisecond*time.Duration(10000))
	defer cancel()
	err := d.iotHubClient.SendEvent(timeoutCtx, body,
		iotdevice.WithSendCorrelationID(correlationID),
		iotdevice.WithSendMessageID(messageID),
		iotdevice.WithSendProperties(map[string]string{
			"iothub-creation-time-utc":    time.Now().Format(time.RFC3339),
			"iothub-connection-device-id": d.deviceID,
			"iothub-interface-id":         "",
		}))
	if err != nil {
		log.Error().
			Str("deviceID", d.deviceID).
			Err(err).
			Msg("error sending telemetry to hub")
		d.retryCount++
		d.sendingTelemetry = false
		return false
	} else {
		d.retryCount = 0
		d.sendingTelemetry = false
		//log.Debug().Str("payload", string(body)).Msg("sent telemetry")
	}
	return true
}

func (d *centralDevice) sendReportedProperties(reportedProps *models.ReportedProperties) {
	// if the device is in the middle of sending a reported property update, skip this request
	if d.sendingReportedProps {
		log.Trace().
			Str("deviceID", d.deviceID).
			Msg("skipping reported properties as it is already sending one")
		return
	}

	d.sendingReportedProps = true

	// make sure that the device is connected
	if d.isConnected == false {
		if d.connectDevice() == false {
			d.sendingReportedProps = false
			return
		}
	}

	//desired, reported, _ := req.device.iotHubClient.RetrieveTwinState(req.device.context)
	//log.Debug().Str("deviceID", req.device.deviceID).Msg(fmt.Sprintf("current desired: %v, reported: %v", desired, reported))

	// generate reported properties
	reportedTwin := make(iotdevice.TwinState)
	reportedTwin["hostName"] = reportedProps.HostName
	reportedTwin["ipAddress"] = reportedProps.IPAddress
	reportedTwin["hostTime"] = reportedProps.HostTime

	// send the reported properties to IoT Central
	timeoutCtx, cancel := context.WithTimeout(d.context, time.Millisecond*time.Duration(10000))
	defer cancel()
	_, err := d.iotHubClient.UpdateTwinState(timeoutCtx, reportedTwin)
	if err != nil {
		log.Debug().Err(err).Str("deviceID", d.deviceID).Msg("error sending reported properties update")
		d.retryCount++
	} else {
		payload, _ := json.Marshal(reportedTwin)
		log.Debug().
			Str("deviceID", d.deviceID).
			Str("payload", string(payload)).
			Int("numGoroutines", runtime.NumGoroutine()).
			Msg("sent reported properties")
		d.retryCount = 0
	}

	d.sendingReportedProps = false
}

func (d *centralDevice) connectDevice() bool {
	// provision the device for the first time
	if d.provisionDevice() == false {
		return false
	}

	// if the device is in the middle of connecting, ignore this request
	if d.isConnecting {
		return false
	}

	d.isConnecting = true
	var err error
	// connect the device to IoT Central
	d.iotHubClient, err = iotdevice.NewFromConnectionString(iotmqtt.New(), d.connectionString,
		iotdevice.WithLogger(logger.New(logger.LevelDebug, func(lvl logger.Level, s string) {
			log.Trace().Msg(s)
		})))
	if err != nil {
		d.isConnecting = false
		log.Error().Err(err).Str("deviceID", d.deviceID).Str("connectionString", d.connectionString).Msg("error parsing connection string")
		return false
	}

	log.Trace().Str("deviceID", d.deviceID).Str("connectionString", d.connectionString).Msg("trying to connect to iothub")
	timeoutCtx, cancel := context.WithTimeout(d.context, time.Millisecond*time.Duration(10000))
	defer cancel()
	if err = d.iotHubClient.Connect(timeoutCtx); err != nil {
		d.isConnecting = false
		log.Error().Err(err).Str("deviceID", d.deviceID).Msg("error connecting to IoT Hub")

		// device might have moved to a different hub, provision and connect to hub again
		errMsg := strings.ToLower(err.Error())
		if errMsg == "not authorized" || errMsg == "server unavailable" || strings.Contains(errMsg, "network error") {
			log.Trace().Str("deviceID", d.deviceID).Msg("detected hub fail over, re-provisioning device")

			if d.provisionDevice() == false {
				return false
			}

			// close existing hub connections
			_ = d.iotHubClient.Close()
			d.iotHubClient = nil

			d.iotHubClient, _ = iotdevice.NewFromConnectionString(iotmqtt.New(), d.connectionString,
				iotdevice.WithLogger(logger.New(logger.LevelDebug, func(lvl logger.Level, s string) {
					log.Trace().Msg(s)
				})))
			timeoutCtx, cancel := context.WithTimeout(d.context, time.Millisecond*time.Duration(10000))
			defer cancel()
			if err = d.iotHubClient.Connect(timeoutCtx); err != nil {
				log.Error().Err(err).Str("deviceID", d.deviceID).Str("connectionString", d.connectionString).Msg("error connecting to IoT Hub")
				return false
			}
			log.Debug().Str("deviceID", d.deviceID).Msg("detected hub fail over, reconnected to IoT Hub")
		} else {
			return false
		}
	}
	log.Trace().Err(err).Str("deviceID", d.deviceID).Msg("device connected to IoT Hub")

	// register for twin updates
	if d.subscribeTwinUpdates() == false {
		d.isConnecting = false
		return false
	}

	// register for c2d commands
	if d.subscribeCommands() == false {
		d.isConnecting = false
		return false
	}

	d.isConnected = true
	d.isConnecting = false

	return true
}

// disconnectDevice disconnects a given device from IoT Central
func (d *centralDevice) disconnectDevice() bool {
	if d.iotHubClient != nil {
		// stop all go functions e.g.: twin update acknowledgements, command acknowledgements
		d.cancel()

		// unregister for twin updates
		d.unsubscribeTwinUpdates()

		// unregister from c2d commands and direct methods
		d.unsubscribeCommands()

		_ = d.iotHubClient.Close()
		d.iotHubClient = nil
		d.context, d.cancel = context.WithCancel(d.context)
	}
	log.Trace().Str("deviceID", d.deviceID).Msg("disconnected device from IoT Hub")

	// do not reset connection string
	// we reuse the connection string until we get a failure

	d.isConnected = false

	return true
}

// subscribeTwinUpdates creates subscription to monitor twin update (desired property) requests for a given device
func (d *centralDevice) subscribeTwinUpdates() bool {
	var err error
	timeoutCtx, cancel := context.WithTimeout(d.context, time.Millisecond*time.Duration(10000))
	defer cancel()
	d.twinSub, err = d.iotHubClient.SubscribeTwinUpdates(timeoutCtx)
	if err != nil {
		// TODO: add retry
		log.Err(err).Str("deviceID", d.deviceID).Msg("twin update subscription failed")
		return false
	}

	go func() {
		for {
			select {
			case <-d.context.Done():
				log.Trace().Str("deviceID", d.deviceID).Msg("device twin subscription stopped")
				return
			case desiredTwin := <-d.twinSub.C():
				dt, _ := json.Marshal(desiredTwin)
				log.Trace().Str("deviceID", d.deviceID).
					Str("desiredTwin", fmt.Sprintf("%s", dt)).
					Msg("got twin update")

				// acknowledge twin update by echoing reported properties
				d.applyTwinUpdate(desiredTwin, false)
			}
		}
	}()

	return true
}

// unsubscribeTwinUpdates unsubscribe from twin updates for a given device
func (d *centralDevice) unsubscribeTwinUpdates() bool {
	if d.twinSub != nil {
		d.iotHubClient.UnsubscribeTwinUpdates(d.twinSub)
	}
	return true
}

// subscribeCommands subscribe for c2d command requests from IoT Central to the device
func (d *centralDevice) subscribeCommands() bool {
	// register for (Sync) Direct Methods

	return true
}

// unsubscribeCommands unsubscribe from c2d command requests for a given device
func (d *centralDevice) unsubscribeCommands() bool {

	return true
}

// provisionDevice provision the device in Central
func (d *centralDevice) provisionDevice() bool {
	// provision the device for the first time
	req := &ProvisioningRequest{
		DeviceID:    d.deviceID,
		Context:     d.context,
		Application: d.app,
		ModelID:     d.app.BoltMachineModelID,
	}
	result := d.provisioner.Provision(req)
	if result == nil {
		return false
	}
	d.connectionString = result.ConnectionString

	log.Trace().Str("deviceId", d.deviceID).Str("connectionString", result.ConnectionString).Msg("provisioned device")

	return true
}

func (d *centralDevice) applyTwinUpdate(desiredTwin iotdevice.TwinState, forceUpdate bool) bool {
	reportedTwin := make(iotdevice.TwinState)
	desiredVersion := desiredTwin.Version()
	deviceChanged := false
	for key, value := range desiredTwin {
		switch key {
		case "telemetryFrequency":
			val, responseTwin, ok := d.getIntTwinValue(key, value, desiredVersion)
			if ok {
				reportedTwin[key] = responseTwin
				d.telemetryFrequency = val
				deviceChanged = true
			}
		case "shiftDurationHours":
			val, responseTwin, ok := d.getIntTwinValue(key, value, desiredVersion)
			if ok {
				reportedTwin[key] = responseTwin
				d.shiftDurationHours = val
				deviceChanged = true
			}
		case "batchDurationHours":
			val, responseTwin, ok := d.getIntTwinValue(key, value, desiredVersion)
			if ok {
				reportedTwin[key] = responseTwin
				d.batchDurationHours = val
				deviceChanged = true
			}
		case "isMachineOn":
			val, responseTwin, ok := d.getBoolTwinValue(key, value, desiredVersion)
			if ok {
				reportedTwin[key] = responseTwin
				d.isMachineOn = val
				deviceChanged = true
			}
		}
	}

	timeoutCtx, cancel := context.WithTimeout(d.context, time.Millisecond*time.Duration(10000))
	defer cancel()
	_, err := d.iotHubClient.UpdateTwinState(timeoutCtx, reportedTwin)
	if err != nil {
		log.Err(err).Str("deviceID", d.deviceID).Msg("twin update failed")
		return false
	} else {
		rt, _ := json.Marshal(reportedTwin)
		log.Debug().Str("deviceID", d.deviceID).
			Str("reportedProperties", fmt.Sprintf("%s", rt)).
			Msg("acknowledged twin update")
	}

	// apply device changes
	if deviceChanged || forceUpdate {
		// reset wait loops
		d.telemetryWaitCancel()
		d.telemetryWaitContext, d.telemetryWaitCancel = context.WithCancel(d.context)
		d.reportedWaitCancel()
		d.reportedWaitContext, d.reportedWaitCancel = context.WithCancel(d.context)

		// restart pumps
		go d.startTelemetryPump()
		go d.startReportedPropsPump()
	}

	return true
}

func (d *centralDevice) applyInitialTwinState() bool {
	desired, _, err := d.iotHubClient.RetrieveTwinState(d.context)
	if err != nil {
		return false
	}
	d.applyTwinUpdate(desired, true)

	body, _ := json.Marshal(desired)
	log.Debug().Str("desired", string(body)).Msg("applied desired twin state")
	return true
}

func (d *centralDevice) getStringTwinValue(name string, value interface{}, desiredVersion int) (string, map[string]interface{}, bool) {
	stringVal, ok := value.(string)
	if !ok {
		log.Error().Str(name, fmt.Sprintf("%v", value)).Msg("got illegal twin data")
		return "", nil, false
	}
	responseTwin := map[string]interface{}{
		"value": value,
		"ac":    200,
		"ad":    "completed",
		"av":    desiredVersion,
	}
	return stringVal, responseTwin, true
}

func (d *centralDevice) getIntTwinValue(name string, value interface{}, desiredVersion int) (int, map[string]interface{}, bool) {
	floatVal, ok := value.(float64)
	if !ok {
		log.Error().Str(name, fmt.Sprintf("%v", value)).Msg("got illegal twin data")
		return 0, nil, false
	}
	responseTwin := map[string]interface{}{
		"value": value,
		"ac":    200,
		"ad":    "completed",
		"av":    desiredVersion,
	}
	return int(floatVal), responseTwin, true
}

func (d *centralDevice) getBoolTwinValue(name string, value interface{}, desiredVersion int) (bool, map[string]interface{}, bool) {
	boolVal, ok := value.(bool)
	if !ok {
		log.Error().Str(name, fmt.Sprintf("%v", value)).Msg("got illegal twin data")
		return false, nil, false
	}
	responseTwin := map[string]interface{}{
		"value": value,
		"ac":    200,
		"ad":    "completed",
		"av":    desiredVersion,
	}
	return boolVal, responseTwin, true
}

func (d *centralDevice) getReportedProperties() (*models.ReportedProperties, error) {
	reportedProps := models.ReportedProperties{
		HostName:  d.getHostName(),
		IPAddress: d.getLocalIP(),
		HostTime:  time.Now(),
	}
	return &reportedProps, nil
}

// sleep sleeps for the given duration with cancellation context
func (d *centralDevice) sleep(ctx context.Context, duration time.Duration) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(duration):
	}
}

// randSleep sleeps for random time within the given min/max range with cancellation context
func (d *centralDevice) randSleep(ctx context.Context, minMs int, maxMs int) {
	d.sleep(ctx, time.Millisecond*time.Duration(minMs+rand.Intn(maxMs)))
}

func (d *centralDevice) getHostName() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	return hostname
}

// getLocalIP returns the non loop back local IP of the host
func (d *centralDevice) getLocalIP() string {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addresses {
		// check the address type and if it is not a loop back the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func (d *centralDevice) getBoltTelemetryPayload(tm *models.BoltMachineTelemetryMessage, opcua bool) ([]byte, error) {
	if opcua {

		// OPCUA device sending JSON payload
		msgGuid, _ := uuid.GenerateUUID()
		payload := make(map[string]interface{})
		msgList := make([]map[string]interface{}, 1)
		d.telemetrySequenceNumber++
		msgList[0] = map[string]interface{}{
			"DataSetWriterId": fmt.Sprintf("%s-%s", d.deviceID, msgGuid),
			"MetaDataVersion": map[string]interface{}{
				"MajorVersion": 1,
				"MinorVersion": 0,
			},
			"SequenceNumber": d.telemetrySequenceNumber, //  rand.Intn(100000),
			"Status":         nil,
			"Timestamp":      d.getDateTime(),
			"Payload":        payload,
		}

		eventId, _ := uuid.GenerateUUID()
		telemetryValues := map[string]interface{}{
			"DataSetClassId":     nil,
			"DataSetWriterGroup": d.deviceID,
			"EventId":            eventId,
			"MessageId":          d.getString(5),
			"MessageType":        "ua-data",
			"PublisherId":        "Standalone_IIOTEdgeServer_opcpublisher",
			"Messages":           msgList,
		}

		now := time.Now().UTC()
		tvList := map[string]interface{}{
			"plantName":          tm.PlantName,
			"productionLine":     tm.ProductionLine,
			"shiftNumber":        tm.ShiftNumber,
			"batchNumber":        tm.BatchNumber,
			"messageTimestamp":   tm.MessageTimestamp,
			"totalPartsMade":     tm.TotalPartsMade,
			"defectivePartsMade": tm.DefectivePartsMade,
			"machineHealth":      tm.MachineHealth,
			"oilLevel":           tm.OilLevel,
			"temperature":        tm.Temperature,
			"Kwh":        		  tm.Kwh,
		}

		for name, value := range tvList {
			opcuaNodeId := fmt.Sprintf("nsu=%s;s=%s", d.getString(20), d.getString(20))
			payload[opcuaNodeId] = map[string]interface{}{
				"ServerTimestamp": now,
				"SourceTimestamp": now,
				"StatusCode":      nil,
				"Name":            name,
				"Value":           value,
			}
			telemetryValues[name] = value
		}

		body, err := json.Marshal(telemetryValues)
		if err != nil {
			return nil, err
		}
		return body, err

	} else {
		body, err := json.Marshal(tm)
		if err != nil {
			return nil, err
		}
		return body, err
	}
}

// getDateTime gets current date time as a string.
func (d *centralDevice) getDateTime() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// getString gets a random string.
func (d *centralDevice) getString(length int) string {
	var charSet string = "abcdefghijklmnopqrstuvwxyzACBDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var val strings.Builder
	for i := 0; i < length; i++ {
		val.WriteString(string(charSet[rand.Intn(len(charSet))]))
	}

	return val.String()
}

//getTime gets the current time as string.
func (d *centralDevice) getTime() string {
	return time.Now().UTC().Format(time.RFC3339)
}
