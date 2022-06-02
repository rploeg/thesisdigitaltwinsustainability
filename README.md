# Thesis  Digital Twin and sustainability in manufacturing
Welcome. This repo describes the installation and configuration of the sustainable digital twin used for my thesis. Please contact me if you have any questions!

All the steps that I have taken are noted [here](https://github.com/rploeg/thesisdigitaltwinsustainability/blob/main/documentation/logbook.md) in my logbook.

# Architecture
Below you can see the high level architecture of the sustainable digital twin<br>

![image](https://user-images.githubusercontent.com/49752333/171601775-bb071c30-ff8b-4bf7-bbcf-8ca9f909cb27.png)
<br>
1. Machines - are simulators of machine data
2. Azure IoT Central is used to receive the raw machine data
3. Azure Data Explorer is used for long term storage and machine learning (forecast energy and anomaly)
4. Azure Digital Twins for latest data points of machine and placing it in the context of the production line
5. Azure Functions sends data from Azure IoT Central to Azure Digital Twins
6. Azure Digital Twins 3D scenes is used to create the 3D view of the sustainable digital twin
7. OEEE dashboard is created on Azure Data Explorer to calculate the OEEE and showcase the forecasts of energy and problems in the production line

# Installation of base infrastructuur
On the following page you can find the installation of the Azure components used in the sustainable digital twin
https://github.com/rploeg/thesisdigitaltwinsustainability/blob/main/documentation/install.md


# Configuration Digital Twin
On the following page you can find the configuration of the Digital Twin that represent sustainable digital twin <br>
https://github.com/rploeg/thesisdigitaltwinsustainability/blob/main/documentation/configdigitaltwin.md

# Configuration of Simulator

On the following page you can find the configuration of the simulator that send simulated data to the sustainable digital twin <br>
[https://github.com/rploeg/thesisdigitaltwinsustainability/blob/main/documentation/datapoint.md](https://github.com/rploeg/thesisdigitaltwinsustainability/blob/main/documentation/simulator.md)

# Use Machine Learning components
In the research also two algorithms are used.<br>
https://github.com/rploeg/thesisdigitaltwinsustainability/blob/main/documentation/ml.md

# Result

If everything is configured correctly you should have two dahsboard:

1. OEEE dashboard<br>
![image](https://user-images.githubusercontent.com/49752333/171603282-bf3c6730-a6dc-4656-bd4c-7a6a7fcebe1b.png)
<br>
2. 3D view of the sustainable Digital Twin

![image](https://user-images.githubusercontent.com/49752333/171603832-bbdc3249-0173-40dc-b240-646832cc0730.png)



# Used libraries
The following libraries are used to build up the DTDL tree in Azure Digital Twin. Because placing comments is not allowed in the DTDL structure the credits are placed here. Also the other libraries that are used are here:<br>
DTDL: https://github.com/Azure-Samples/digital-twins-samples/tree/master/HandsOnLab <br>
OEE: https://github.com/Azure/iot-central-industrial-OEE - used for simulator and baseline for machine template

I want to thank these authors for their work!
