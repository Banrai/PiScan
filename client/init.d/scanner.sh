#! /bin/sh
### BEGIN INIT INFO
# Provides:          scanner.sh
# Required-Start:    $remote_fs $syslog
# Required-Stop:     $remote_fs $syslog
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: Runs the PiScanner binary
# Description:       Makes sure the PiScanner binary starts on boot
### END INIT INFO

case "$1" in
  start)
    echo "Starting PiScanner"
    /home/pi/PiScanner >> /home/pi/PiScanner.log 2>&1
    ;;
  stop)
    echo "Stopping PiScanner"
    killall PiScanner
    ;;
  *)
    echo "Usage: /etc/init.d/scanner.sh {start|stop}"
    exit 1
    ;;
esac

exit 0
