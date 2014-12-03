#/bin/sh
sed -i '/echo 2 > \/proc\/sys\/vm\/overcommit_memory/c\echo 1 > \/proc\/sys\/vm\/overcommit_memory' /etc/init.d/rcS
echo "sleep 10" >> /etc/init.d/rcS
echo "/data/video/ardrone_commander&" >> /etc/init.d/rcS

echo "1781 digispark" >> /etc/usb.ids
echo "  t0c9f digispark" >> /etc/usb.ids
echo "16d0 digispark" >> /etc/usb.ids
echo "  t0753 digispark" >> /etc/usb.ids

chmod +x /data/video/ardrone_commander

reboot
