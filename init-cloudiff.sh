#!/bin/bash
#
# chkconfig: 35 95 05
# description: Cloudiff application.

# Run at startup: sudo chkconfig cloudiff on

# source init-functions
source /lib/lsb/init-functions 

# Name of the application
app="cloudiff"

pid_file="/var/run/$app.pid"
get_pid() {
    cat "$pid_file"
}

is_running() {
    [ -f "$pid_file" ] && ps -p `get_pid` > /dev/null 2>&1
}

# Start the service
run() {
  echo -n $"Starting $app:"
  cd /home/cloudiff/go/src/ileansys.com/cloudiff  #set path to your cloudiff binary
  su - cloudiff ./$app > /home/cloudiff/$app.log 2> /home/cloudiff/$app.err < /dev/null &
  
  sleep 1
  
  status $app > /dev/null
  # If application is running
  if [[ $? -eq 0 ]]; then
    # Store PID in lock file
    echo $! > /var/lock/subsys/$app
    success
    echo
  else
    failure
    echo
  fi
}

# Start the service
start() {
  status $app > /dev/null
  # If application is running
  if [[ $? -eq 0 ]]; then
    status $app
  else
    run
  fi
}

# Restart the service
stop() {
  echo -n "Stopping $app: "
  killproc $app
  rm -f /var/lock/subsys/$app
  echo
}

# Reload the service
reload() {
  status $app > /dev/null
  # If application is running
  if [[ $? -eq 0 ]]; then
    echo -n $"Reloading $app:"
    kill -HUP `pidof $app`
    sleep 1
    status $app > /dev/null
    # If application is running
    if [[ $? -eq 0 ]]; then
      success
      echo
    else
      failure
      echo
    fi
  else
    run
  fi
}

# Main logic
case "$1" in
  start)
    start
    ;;
  stop)
    stop
    ;;
  status)
    if is_running; then
        echo "Running"
    else
        echo "Stopped"
        exit 1
    fi
    ;;
  restart)
    stop
    sleep 1
    start
    ;;
  reload)
    reload
    ;;
  *)
    echo $"Usage: $0 {start|stop|restart|reload|status}"
    exit 1
esac
exit 0