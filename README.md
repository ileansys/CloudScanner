# Cloudiff

A cloud security compliance change tracker. 

![alt text](https://github.com/ileansys/cloudiff/blob/master/cloudiff.png?raw=true)

See beyond the clouds...

### Current Features
- Retrieve IPs from AWS, GCP and DigitalOcean Cloud SDKs for compute resources
- Detect new public IPs (outliers) from a previous IP baseline across AWS, GCP and Digital Ocean 
- Scan those new IPs to detect services
- Store the new+old IPs as the new baseline
- Customized Schedules and Scan Intervals (cron)
- Storing Service Scan Baselines
- Change Notifications and Alerts

### Integrations
- AWS Go SDKs
- GCP Go SDKs
- DigitalOcean Go SDKs

### OS Tools/Dependicies
- go - Writing and building code
- memcached - Storing IP and service scan data
- xalan - Converting xml to html reports
- nmap & nse scripts - Service scans

### Roadmap Features
- Configurability of scan workflows using yaml
- Integrate with Nuclei or other suitable DAST scanner
- Other Cloud Provider Modules
- Replace memcached with PostgreSQL

### Disclaimer!!!
- Cloudiff is still in active developement
