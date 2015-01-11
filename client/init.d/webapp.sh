#! /bin/sh
### BEGIN INIT INFO
# Provides:          webapp.sh
# Required-Start:    $remote_fs $syslog
# Required-Stop:     $remote_fs $syslog
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: Runs the client WebApp binary
# Description:       Makes sure the client WebApp binary starts on boot
### END INIT INFO

case "$1" in
  start)
    echo "Starting WebApp"
    /home/pi/WebApp -templates /home/pi/ui/templates >> /home/pi/WebApp.log 2>&1
    ;;
  stop)
    echo "Stopping WebApp"
    killall WebApp
    ;;
  *)
    echo "Usage: /etc/init.d/webapp.sh {start|stop}"
    exit 1
    ;;
esac

exit 0
