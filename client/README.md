The client components can run on *any* computer, though this project was designed with the [Raspberry Pi](http://www.raspberrypi.org/) in mind, since it is compact, and efficient for a device mean to be mostly on all the time.

The two binaries in this folder should be built and run, after connecting the barcode scanner usb device.

The client datastore is a simple [SQLite](http://sqlite.org/) database file, consisting of the [basic tables](database) needed to keep track of individual user scans, product data contributions, and favorited items.

## Initial Raspberry Pi setup

1. Create a bootable SD card using the [Raspbian download](http://www.raspberrypi.org/downloads/), following the [image installation guide](http://www.raspberrypi.org/documentation/installation/installing-images/README.md) for your OS

  Under linux, this involves using the [dd](http://linux.die.net/man/1/dd) command to copy the unzipped Raspbian img file onto a newly unmounted micro SD card:

  ```sh
# umount /dev/sdb1 # use df or fdisk -l to determine SD card location
# dd bs=4M if=/opt/downloads/2014-12-24-wheezy-raspbian.img of=/dev/sdb
# sync
  ```

  Eject the SD card, remove it from its adapter, and install the micro SD card in the Pi (push click to confirm). The Pi is now ready to be booted for the first time.

2. Boot the Pi for the first time

  Connect with either an HDMI monitor and keyboard, or ssh terminal, if also using an ethernet cable to connect to a local network.

  The [config screen](http://www.raspberrypi.org/documentation/configuration/raspi-config.md) will appear. 

  Choose the first option, [Expand Filesystem](http://www.raspberrypi.org/documentation/configuration/raspi-config.md#expand-filesystem), and hit enter. If successful, you will see this prompt:

  ![Imgur](http://i.imgur.com/GORxuPk.jpg "Expand File Reboot")
	
  Hit enter again, to say <tt>Ok</tt>, then select <tt>Finish</tt> to exit [raspi-config](http://www.raspberrypi.org/documentation/configuration/raspi-config.md) but do not reboot just yet.

3. Configure the SD card for long-term use 

  Use [tmpfs](https://en.wikipedia.org/wiki/Tmpfs) to have folders with frequent writes, such as <tt>/var/log</tt>, to write to RAM instead of the local disk (in this case, the SD card), which will prolong its life (we hope)

  Backup the original <tt>/etc/fstab</tt> file:
	
  ```sh
pi@raspberrypi ~ $ sudo cp -ip /etc/fstab /etc/fstab.bak
  ```
      
  Then edit it to include these lines:
	
  ```
tmpfs    /tmp    tmpfs    defaults,noatime,nosuid,size=100m    0 0
tmpfs    /var/tmp    tmpfs    defaults,noatime,nosuid,size=30m    0 0
tmpfs    /var/log    tmpfs    defaults,noatime,nosuid,mode=0755,size=100m    0 0
tmpfs    /var/run    tmpfs    defaults,noatime,nosuid,mode=0755,size=2m    0 0
tmpfs    /var/spool/mqueue    tmpfs    defaults,noatime,nosuid,mode=0700,gid=12,size=30m    0 0
  ```
     
  Do a [diff](http://linux.die.net/man/1/diff) and comfirm the changes:
	
  ```sh
pi@raspberrypi ~ $ sudo diff /etc/fstab /etc/fstab.bak
5,10d4
< # tmpfs settings to minimize writes to the SD card
< tmpfs    /tmp    tmpfs    defaults,noatime,nosuid,size=100m    0 0
< tmpfs    /var/tmp    tmpfs    defaults,noatime,nosuid,size=30m    0 0
< tmpfs    /var/log    tmpfs    defaults,noatime,nosuid,mode=0755,size=100m    0 0
< tmpfs    /var/run    tmpfs    defaults,noatime,nosuid,mode=0755,size=2m    0 0
< tmpfs    /var/spool/mqueue    tmpfs    defaults,noatime,nosuid,mode=0700,gid=12,size=30m    0 0
  ```

  Disable [swap](https://en.wikipedia.org/wiki/Linux_swap#LINUX), which uses part of the SD card as volatile memory.

  While swap does increase the amount of RAM available, on this hardware it is unlikely to increase performance significantly, and it results in a high number of read and writes which has a negative effect on the SD card's long-term viability.

  ```sh
pi@raspberrypi ~ $ sudo swapoff --all
pi@raspberrypi ~ $ sudo dphys-swapfile swapoff --all
  ```

  To prevent swap from coming back after rebooting, remove the [dphys-swapfile](http://neil.franklin.ch/Projects/dphys-swapfile/) package entirely:
	
  ```sh
pi@raspberrypi ~ $ sudo apt-get remove dphys-swapfile
Reading package lists... Done
Building dependency tree       
Reading state information... Done
The following packages will be REMOVED:
  dphys-swapfile
0 upgraded, 0 newly installed, 1 to remove and 0 not upgraded.
After this operation, 69.6 kB disk space will be freed.
Do you want to continue [Y/n]? y
(Reading database ... 69052 files and directories currently installed.)
Removing dphys-swapfile ...
Stopping dphys-swapfile swapfile setup ..., done.
Processing triggers for man-db ...
  ```

  Finally, reboot:

  ```sh
pi@raspberrypi ~ $ sudo reboot
  ```

  ```sh
Broadcast message from root@raspberrypi (pts/0) (Wed Aug 27 19:08:12 2014):
The system is going down for reboot NOW!
  ```

  **Configuration References**
  * [Raspberry Pi: Extending the life of the SD card](http://www.zdnet.com/raspberry-pi-extending-the-life-of-the-sd-card-7000025556/)
  * [How can I extend the life of my SD card?](http://raspberrypi.stackexchange.com/a/186)

4. Add External Storage

  Create a usb drive that will act as the container or target folder for the embedded database, so that it does *not* reside on the sd card running Raspbian.

  This will enable us to remove and replace external storage without affecting how the core services and applications run.

  Start by formatting the usb drive as an [ext4](https://en.wikipedia.org/wiki/Ext4) filesystem.

  Do this on any computer *outside* of the Raspberry Pi.

  On linux, plug in the usb stick and use <tt>fdisk -l</tt> to determine the device name, and partition number, in the form `/dev/[DEVICEn]`.

  If there is no partition defined, a message similar to this one will appear:

  ```
Disk /dev/[DEVICE] doesn't contain a valid partition table
```

  If this does occur, use <tt>fdisk</tt> to create one:

  ```
fdisk /dev/[DEVICE]
```

  Use <tt>n</tt> to create a new partition, make it the <tt>p</tt> primary, say yes to all the defaults, and finally choose <tt>w</tt> to save the changes.

  To format the usb stick, unmount the stick with `sudo umount /dev/[DEVICEn]` where `n` represents the partition number.

  Use the <tt>sudo mkfs.ext4 /dev/[DEVICEn]</tt> command to do the actual formatting, where <tt>[DEVICEn]</tt> points to the usb drive, for example <tt>/dev/sdb1</tt>.

  Back on the Raspberry Pi, create a permanent mount point, as <tt>/data</tt>

  ```sh
pi@raspberrypi ~ $ sudo mkdir /data
  ```

  Use <tt>fdisk -l</tt> to confirm where the usb device appears. If this is the first and only external usb storage device, it *should* be <tt>/dev/sda1</tt> but double-check before mounting and updating the <tt>/etc/fstab</tt> file.

  Mount the external drive to <tt>/data</tt> with this command:

  ```sh
pi@raspberrypi ~ $ sudo mount -t ext4 /dev/sda1 /data
   ```

  Next, update <tt>/etc/fstab</tt> so the usb drive is mounted to <tt>/data</tt> on boot.

  Add this line at the end of the file:

  ```sh
/dev/sda1       /data        ext4   rw,exec,auto,users    0  1
  ```

  The difference between the original <tt>/etc/fstab</tt> file and the current one should be:

  ```sh
pi@raspberrypi ~ $ sudo diff /etc/fstab /etc/fstab.bak
5,12d4
< # tmpfs settings to minimize writes to the SD card
< tmpfs    /tmp    tmpfs    defaults,noatime,nosuid,size=100m    0 0
< tmpfs    /var/tmp    tmpfs    defaults,noatime,nosuid,size=30m    0 0
< tmpfs    /var/log    tmpfs    defaults,noatime,nosuid,mode=0755,size=100m    0 0
< tmpfs    /var/run    tmpfs    defaults,noatime,nosuid,mode=0755,size=2m    0 0
< tmpfs    /var/spool/mqueue    tmpfs    defaults,noatime,nosuid,mode=0700,gid=12,size=30m    0 0
< # make sure our external usb drive is mounted on boot
< /dev/sda1       /data        ext4   rw,exec,auto,users    0  1
  ```

  Reboot the Raspberry Pi, and confirm that the mount succeeded:

  ```sh
pi@raspberrypi ~ $ ls -ltr /data
total 16
drwx------ 2 root root 16384 Aug 27 19:49 lost+found
  ```

  Change permission so that <tt>/data</tt> is accessible by the <tt>pi</tt> user, and confirm with a simple read and delete:

  ```sh
pi@raspberrypi ~ $ sudo chown -R pi:pi /data
pi@raspberrypi ~ $ ls -ltr /data
total 16
drwx------ 2 pi pi 16384 Aug 27 19:49 lost+found
pi@raspberrypi ~ $ touch /data/x
pi@raspberrypi ~ $ ls -ltr /data
total 16
drwx------ 2 pi pi 16384 Aug 27 19:49 lost+found
-rw-r--r-- 1 pi pi     0 Aug 27 20:20 x
pi@raspberrypi ~ $ rm /data/x
pi@raspberrypi ~ $ ls -ltr /data
total 16
drwx------ 2 pi pi 16384 Aug 27 19:49 lost+found
  ```
		
5. Optional: wifi setup

  If you purchased a usb wifi dongle, it is a good idea to configure it so that it will attach to the desired wifi access point on boot.

  Set your access point configuration using either the [built-in graphical interface](https://learn.adafruit.com/adafruits-raspberry-pi-lesson-3-network-setup/setting-up-wifi-with-raspbian), [edit /etc/network/interfaces manually](https://gist.github.com/dpapathanasiou/4599fbae0640aceb598b#editing-etcnetworkinterfaces), or install [wicd-curses](http://sourceforge.net/projects/wicd/):

  ```sh
pi@raspberrypi ~ $ sudo apt-get install wicd-curses
pi@raspberrypi ~ $ sudo wicd-curses
  ```
	
	The <tt>wicd-curses</tt> tool presents an [easy to use menu](http://windowslinuxcommands.blogspot.com/2013/06/raspberry-pis-new-wifi-manager-friend.html) to set the access point and password.

## Installation

Use either (a) the included ARM binaries; or (b) build them from source directly on the Pi.

### (a) Install the client binaries 

  Copy the PiScanner and WebApp files and put them anywhere under the <tt>/home/pi</tt> folder.

  The simplest way is to use the [scp command](http://linux.die.net/man/1/scp) like this (replace <tt>192.168.1.108</tt> with the actual IP address of your Pi on your network):

  ```sh
  cd binaries/linux/arm
  scp PiScanner pi@192.168.1.108:/home/pi
  scp WebApp pi@192.168.1.108:/home/pi
  ```

### (b) Install from source

1.  Install [Go on the Raspberry Pi](http://dave.cheney.net/2012/09/25/installing-go-on-the-raspberry-pi)

2. The client code uses these Go packages:

  ```sh
go get github.com/mxk/go-sqlite/sqlite3
go get github.com/go-sql-driver/mysql
go get github.com/Banrai/PiScan
  ```

  The last package fetch (this repo) results in this warning, which can be ignored:

  ```sh
package github.com/Banrai/PiScan
imports github.com/Banrai/PiScan
imports github.com/Banrai/PiScan: no buildable Go source files in /home/pi/go-workspace/src/github.com/Banrai/PiScan
  ```

3. Build the client binaries:

  ```sh
cd $GOPATH/src/github.com/Banrai/PiScan
make clients
   ```

  This results in two binary files, <tt>PiScanner</tt> and <tt>WebApp</tt> in the client folder which you can leave there, or move to <tt>/home/pi</tt> and run from there.

  This guide will run them from where they are built, i.e., <tt>$GOPATH/src/github.com/Banrai/PiScan/client</tt>.

### Post Install Configuration

1. Initialize the local client database

  Copy the [PiScanDB.sqlite](PiScanDB.sqlite) file from this repo into the <tt>/data</tt> folder on your Pi:

  ```sh
  scp PiScanDB.sqlite pi@192.168.1.108:/data
  ```

2. Copy the client template folders under the [ui](ui) folder onto the Pi (these are required for the [WebApp](../binaries/linux/arm/WebApp) to run).

  The simplest way is to create a single [tar](http://linux.die.net/man/1/tar) archive, use scp to copy it, and then unpack it on the Pi:

  ```sh
  cd client/ui; tar cf /tmp/webapp_templates.tar fonts images js css templates
  scp /tmp/webapp_templates.tar pi@192.168.1.108:/home/pi
  ```

  On the Pi itself, unpack the file under its own folder, <tt>/home/pi/ui</tt>, or elsewhere:

  ```sh
pi@raspberrypi ~ $ mkdir ui
pi@raspberrypi ~ $ mv webapp_templates.tar ui
pi@raspberrypi ~ $ cd ui
pi@raspberrypi ~/ui $ tar xf webapp_templates.tar
pi@raspberrypi ~/ui $ rm webapp_templates.tar
  ```

3. PiScanner and WebApp startup scripts

  Copy the [init.d scripts](init.d) to the Pi:

  ```sh
cd client/init.d
scp webapp.sh pi@192.168.1.108:/tmp
scp scanner.sh pi@192.168.1.108:/tmp
  ```

  Then, on the Pi, install them under <tt>/etc/init.d</tt> with the correct permissions.

  First, for the PiScanner application:

  ```sh
pi@raspberrypi ~ $ sudo mv /tmp/scanner.sh /etc/init.d
pi@raspberrypi ~ $ sudo chmod 755 /etc/init.d/scanner.sh
pi@raspberrypi ~ $ sudo update-rc.d scanner.sh defaults
  ```

  Then the WebApp (client web server) application:

  ```sh
pi@raspberrypi ~ $ sudo mv /tmp/webapp.sh /etc/init.d
pi@raspberrypi ~ $ sudo chmod 755 /etc/init.d/webapp.sh
pi@raspberrypi ~ $ sudo update-rc.d webapp.sh defaults
  ```

