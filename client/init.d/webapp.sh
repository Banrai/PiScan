#! /bin/sh
case "$1" in
  start)
    echo "Starting WebApp"
    /home/pi/WebApp -templates /home/pi/ui/templates 
    ;;
  stop)
    echo "Stopping WebApp"
    killall WebApp
    ;;
  *)
    echo "Usage: /etc/init.d/webapp {start|stop}"
    exit 1
    ;;
esac

exit 0
