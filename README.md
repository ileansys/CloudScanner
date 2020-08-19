# Cloudiff
============================================================================================================

A cloud security and compliance change tracker.

### Current Features
- Retrieve IPs from AWS, GCP and DigitalOcean Cloud SDKs for compute resources
- Detect new public IPs (outliers) from a previous IP baseline across AWS, GCP and Digital Ocean 
- Scan those new IPs to detect services
- Store the new+old IPs as the new baseline

### Integrations
- AWS Go SDKs
- GCP Go SDKs
- DigitalOcean Go SDKs

### Data Stores
- memcached 

### Wrapped Tools
- nmap

### Roadmap Features
- Archive IP Baseline Data in Persistent Database
- Store, Track, Archive Host Service Baselines
- Customized Schedules and Scan Intervals
- Integrate with Vulnerability Scanners e.g. Nessus, OpenVAS, 
- CIS Host and Container/k8s Hardening Inspection with Chef Inspec 
- Ansible CIS Remediation Workflows
- Integrate with more cloud providers e.g. Azure, Oracle, Linode...
- Change Notifications and Alerts
- Gather Threat Intelligence about Public IPs on Platforms like Shodan, PassiveTotal, VirusTotal e.t.c
