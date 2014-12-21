The client components can run on *any* computer, though this project was designed with the [Raspberry Pi](http://www.raspberrypi.org/) in mind, since it is compact, and efficient for a device mean to be mostly on all the time.

The two binaries in this folder should be built and run, after connecting the barcode scanner usb device.

The client datastore is a simple [SQLite](http://sqlite.org/) database file, consisting of the [basic tables](database) needed to keep track of individual user scans, product data contributions, and favorited items.

## Initial Raspberry Pi setup

1. Create a bootable SD card using the [Raspbian download](http://www.raspberrypi.org/downloads/), following the [image installation guide](http://www.raspberrypi.org/documentation/installation/installing-images/README.md) for your OS

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

  Use <tt>fdisk -l</tt> to confirm where the usb device appears. If this is the first and only external usb storage device, it *should* be <tt>/dev/sda1</tt> but double-check before updating the <tt>/etc/fstab</tt> file.

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
< # make sure our external usb drive (BDB data) is mounted on boot
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
		
5. Install [Go on the Raspberry Pi](http://dave.cheney.net/2012/09/25/installing-go-on-the-raspberry-pi)

## Installation

1. The client code uses these Go packages:

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

2. Build the client binaries:

  ```sh
cd $GOPATH/src/github.com/Banrai/PiScan
make clients
   ```

  This results in two binary files, <tt>PiScanner</tt> and <tt>WebApp</tt> in the client folder which you can leave there, or move to <tt>/home/pi</tt> and run from there.

  This guide will run them from where they are built, i.e., <tt>$GOPATH/src/github.com/Banrai/PiScan/client</tt>.

3. Initialize the local client database

  Before either the <tt>PiScanner</tt> or <tt>WebApp</tt> binaries can be run in normal mode, the local client sqlite database should be created.

  Run this command just once:

  ```sh
$GOPATH/src/github.com/Banrai/PiScan/client/PiScanner -sqliteTables $GOPATH/src/github.com/Banrai/PiScan/client/database
  ```

  The default target folder is <tt>/data</tt> on the Raspberry Pi, but that can be changed using the <tt>-sqlitePath</tt> command line argument when running the command above.

  If successful, the system will respond with a message like this:

  ```sh
2014/12/21 00:23:48 Client database 'PiScanDB.sqlite' created in '/data'
  ```

## Running the client code

  *[coming soon: need to add init.d scripts]*