#!/bin/bash
# Running this script as a cronjob to check if cloudiff is working
# Example 
# # min   hour    day month   dow cmd
# */1 *   *   *   *   /path/to/cloudiff-service-check.sh
#
STATUS=$(/etc/init.d/cloudiff status)
# Most services will return something like "OK" if they are in fact "OK"
test "$STATUS" = "Running" || /etc/init.d/cloudiff restart