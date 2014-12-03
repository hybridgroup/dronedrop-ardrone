#/bin/sh

cp /data/video/rcS.stock /etc/init.d/rcS
cp /data/video/usb.ids.stock /etc/usb.ids
kill $(pidof ardrone_commander)
