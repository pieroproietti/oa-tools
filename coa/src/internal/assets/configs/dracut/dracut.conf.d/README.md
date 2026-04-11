# dracut.conf.d/50-live.conf
mdadmconf="no"
lvmconf="no"
use_fstab="no"
squash_compress="zstd"
compress="zstd"
add_dracutmodules+=" crypt livenet dmsquash-live dmsquash-live-autooverlay convertfs pollcdrom qemu qemu-net "
hostonly="no"
early_microcode="no"

