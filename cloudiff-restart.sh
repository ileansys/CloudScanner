#!/bin/bash
# Running this script as a cronjob to check if cloudiff is working
# * * * * * /home/cloudiff/go/src/ileansys.com/cloudiff/cloudiff-service-check.sh cloudiff > /dev/null
service=$@
/bin/systemctl restart "$service.service"
