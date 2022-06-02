# Configuration of Azure Digital Twins


# Importing DTDL

1. Import [these files](https://github.com/rploeg/thesisdigitaltwinsustainability/tree/main/DTDL) in the Azure Digital Twin Explorer.
2. Than craete the following ontology:

![image](https://user-images.githubusercontent.com/49752333/171480694-5e2c9b4c-d8dc-4648-9efd-41e5a507e7c8.png)

3. Enalbe managed idenity of Azure IoT Central in your ADT configuration

# Create Azure Function

The Azure Function sends data from IoTC to Azure Digital Twins (ADT)

1. Publish the [Azure Function](https://github.com/rploeg/thesisdigitaltwinsustainability/tree/main/FunctionIoTCtoADT) to your Azure subscription.
2. Enalbe the managed identity option
3. Create export to the newly created http endpoint of the Azure Function in Auzre IoT Central

# 3D Scene Studio

The following steps are needed to create the 3D view of the digital twin

1. Upload the GBL file to your storage container [link](https://www.turbosquid.com/3d-models/max-line-packaging/767476)
2. Create a new element and connect in to the Twin in [Azure Digital Twins 3D scenes](https://explorer.digitaltwins.azure.net/3DScenes/):
![image](https://user-images.githubusercontent.com/49752333/171597221-5dcf7900-6439-4a99-b340-4f8b50d502bb.png)
Do this for every element in your drawing
3. Create an behaviour to show when a machine is consuming too much energy. When energy level is higher than 90 kwH the machine should go on red
![image](https://user-images.githubusercontent.com/49752333/171597452-2af46464-37a4-42c9-9e51-389c373a2e24.png)
4. Create a Widget to show the current status of the machine 
