using System;
using System.Text.Json.Serialization;

namespace Company.Models
{ 

public class MessageProperties
{
    [JsonPropertyName("iothub-connection-device-id")]
    public string IothubConnectionDeviceId { get; set; }

    [JsonPropertyName("iothub-creation-time-utc")]
    public DateTime IothubCreationTimeUtc { get; set; }

    [JsonPropertyName("iothub-interface-id")]
    public string IothubInterfaceId { get; set; }
}

public class IoTCentralMessage
{
    public string applicationId { get; set; }
    public string deviceId { get; set; }
    public DateTime enqueuedTime { get; set; }
    public Enrichments enrichments { get; set; }
    public MessageProperties messageProperties { get; set; }
    public string messageSource { get; set; }
    public string schema { get; set; }
    public dynamic telemetry { get; set; }
    public string templateId { get; set; }
}

}