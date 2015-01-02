#! /bin/sh
case "$1" in
  start)
    echo "Starting PiScanner"
    /home/pi/PiScanner
    ;;
  stop)
    echo "Stopping PiScanner"
    killall PiScanner
    ;;
  *)
    echo "Usage: /etc/init.d/scanner {start|stop}"
    exit 1
    ;;
esac

exit 0
