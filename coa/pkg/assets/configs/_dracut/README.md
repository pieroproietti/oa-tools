# dracut on Fedora

## Install

### fedora 
```
sudo dnf -y install \
	dracut \
	dracut-live
```

### opensuse
```
sudo zypper -y install \
	dracut \
	dracut-tools
```

### pacchetti da installare/rimuovere per opensuse
sudo zypper rm multipath-tools
sudo zypper install libxkbcommon-tools util-linux kf6-qqc2-desktop-styl


# procedura per grub su btrfs
sudo mount -o subvol=@ /dev/sda2 /mnt
sudo mount -o subvol=@home /dev/sda2 /mnt/home
sudo mount /dev/sda1 /mnt/boot/efi
for i in /dev /proc /sys; do sudo mount --bind $i /mnt$i; done
sudo mount --bind /sys/firmware/efi/efivars /mnt/sys/firmware/efi/efivars
sudo chroot /mnt
grub2-install
grub2-mkconfig -o /boot/grub2/grub.cfg
