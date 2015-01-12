#! /bin/sh
### BEGIN INIT INFO
# Provides:          api-server.sh
# Required-Start:    $remote_fs $syslog
# Required-Stop:     $remote_fs $syslog
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: Runs the APIServer binary
# Description:       Makes sure the APIServer binary starts on boot
### END INIT INFO

case "$1" in
  start)
    echo "Starting APIServer"
    /home/pod/server/APIServer >> /home/pod/server/APIServer.log 2>&1
    ;;
  stop)
    echo "Stopping APIServer"
    killall APIServer
    ;;
  *)
    echo "Usage: /etc/init.d/api-server.sh {start|stop}"
    exit 1
    ;;
esac

exit 0
