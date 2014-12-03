#/bin/sh

cp /data/video/rcS.dronedrop /etc/init.d/rcS
cp /data/video/usb.ids.dronedrop /etc/usb.ids
chmod +x /data/video/ardrone_commander

reboot
