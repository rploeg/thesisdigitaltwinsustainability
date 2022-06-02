# Machine Learning components used

In this research two algorithms are used. 

# Anomaly detection
Based on the kWh usage of the machines anomalies are checked and displayed in Azure Data Explorer dashboards

<code>
  let min_t = datetime(2022-05-24 09:11);
let max_t = datetime(2022-05-24 15:00);
let dt = 2h;
boltmaker
| make-series num=avg(kwh) on messageTimestamp from min_t to max_t step dt by deviceId 
| extend (anomalies, score, baseline) = series_decompose_anomalies(num, 1.5, -1, 'linefit')
| render anomalychart with(anomalycolumns=anomalies, title='Anomalies kWh usage per machine')
  </code>

# Forecast energy usage
Based on the historical energy usage the 1 day energy forecast usage is created

<code>
  let min_t = datetime(2022-05-24 09:11);
let max_t = datetime(2022-05-24 15:00);
let dt = 2h;
let horizon=16h;
boltmaker
| make-series num=avg(kwh) on messageTimestamp from min_t to max_t+horizon step dt by deviceId 
| extend forecast = series_decompose_forecast(num, toint(horizon/dt))
| render timechart with(title='kwH forecast')
  </code>
