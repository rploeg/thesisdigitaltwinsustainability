using Azure;
using Azure.Core.Pipeline;
using Azure.DigitalTwins.Core;
using Azure.Identity;
using System;
using System.IO;
using System.Threading.Tasks;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Azure.WebJobs;
using Microsoft.Azure.WebJobs.Extensions.Http;
using Microsoft.AspNetCore.Http;
using Microsoft.Extensions.Logging;
using Newtonsoft.Json;
using Newtonsoft.Json.Linq;
using System.Net.Http;
using Company.Models;

namespace Company.Function
{
    public static class HttpTriggerToADT
    {

        private static readonly string adtInstanceUrl = Environment.GetEnvironmentVariable("ADT_SERVICE_URL");
        private static readonly HttpClient httpClient = new HttpClient();

        [FunctionName("HttpTriggerToADT")]
        public static async Task<IActionResult> Run(
            [HttpTrigger(AuthorizationLevel.Anonymous, "get", "post", Route = null)] HttpRequest req,
            ILogger log)
        {
            log.LogInformation("C# HTTP trigger function processed a request.");

            string name = req.Query["name"];

            string requestBody = await new StreamReader(req.Body).ReadToEndAsync();
            dynamic data = JsonConvert.DeserializeObject(requestBody);
            name = name ?? data?.name;

            log.LogInformation(requestBody.ToString());
            if (adtInstanceUrl == null) log.LogError("Application setting \"ADT_SERVICE_URL\" not set");
            try
            {
                //Authenticate with Digital Twins
                ManagedIdentityCredential cred = new ManagedIdentityCredential("https://digitaltwins.azure.net");
                DigitalTwinsClient client = new DigitalTwinsClient(new Uri(adtInstanceUrl), cred, new DigitalTwinsClientOptions { Transport = new HttpClientTransport(httpClient) });
                log.LogInformation($"ADT service client connection created.");
                if (requestBody != null && requestBody.ToString() != null)
                {
                    log.LogInformation(requestBody.ToString());

                    // Reading deviceId and temperature from http request
                    var deviceMessage = JsonConvert.DeserializeObject<Message>(requestBody.ToString());
                    string deviceId = (string)deviceMessage.deviceId;
                    log.LogInformation(deviceId);
                    string deviceType = "test";
                     var updateTwinData = new JsonPatchDocument();
                     var updateTwinData2 = new JsonPatchDocument();
                    switch (deviceType){
                        case "test":
                            updateTwinData.AppendAdd("/MotorStatus", deviceMessage.properties[0].value);
                            //updateTwinData.AppendAdd("/MotorStatus", ((JObject)deviceMessage["data.properties"][0]).Value<Boolean>());
                            log.LogInformation("update ADT device");
                            await client.UpdateDigitalTwinAsync(deviceId, updateTwinData);
                            if ((bool)deviceMessage.properties[0].value)
                                {
                                updateTwinData2.AppendAdd("/double01", 30);
                                await client.UpdateDigitalTwinAsync("GenericSensor04", updateTwinData2);
                                }
                                else 
                                {
                                updateTwinData2.AppendAdd("/double01", 0);
                                await client.UpdateDigitalTwinAsync("GenericSensor04", updateTwinData2); 
                                }

                        break;
                    }

                }
            }
            catch (Exception e)
            {
                log.LogInformation("In Expection");
                log.LogError(e.Message);
            }
                log.LogInformation("return message");
               return new OkObjectResult("responseMessage");

        }


        }
    }
