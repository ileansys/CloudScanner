# Cloudiff


![alt text](https://github.com/ileansys/cloudiff/blob/master/cloudiff.png?raw=true)


A cloud security and compliance change tracker. See beyond the clouds...

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

### Data Stores
- memcached 

### Wrapped Tools
- nmap

### Roadmap Features
- Configurability using yaml
- Archive IP Baseline Data in Persistent Database
- Track and Archive Service Scan Baselines
- Integrate with Vulnerability Scanners e.g. Nessus, OpenVAS, 
- CIS Host and Container/k8s Hardening Inspection with Chef Inspec 
- Ansible CIS Remediation Workflows
- Integrate with more cloud providers e.g. Azure, Oracle, Linode...
- Containerization
- Interactive Console
- Custom Provider Modules
- Gather Threat Intelligence about Public IPs on Platforms like Shodan, PassiveTotal, VirusTotal e.t.c
