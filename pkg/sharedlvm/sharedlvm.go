package sharedlvm

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	mnt "github.com/phillipleblanc/sharedlvm/pkg/mount"
	"k8s.io/klog"
	utilexec "k8s.io/utils/exec"
	"k8s.io/utils/mount"
)

func CreateVolumeIfNotExists(name string, volumeGroup string, capacityBytes int64) error {
	err := exec.Command("lvdisplay", "/dev/"+volumeGroup+"/"+name).Run()
	if err == nil {
		return nil
	}

	if _, ok := err.(*exec.ExitError); !ok {
		klog.Errorf("Error checking if volume exists: %s", err.Error())
		return err
	}

	return exec.Command("lvcreate", "-L", fmt.Sprintf("%sb", strconv.FormatInt(capacityBytes, 10)), "-n", name, volumeGroup).Run()
}

func ActivateVolumeGroupLock(volumeGroup string) error {
	return exec.Command("vgchange", "--lockstart", volumeGroup).Run()
}

func ActivateVolume(name string, volumeGroup string) error {
	return exec.Command("lvchange", "-ay", volumeGroup+"/"+name).Run()
}

func DeactivateVolume(name string, volumeGroup string) error {
	return exec.Command("lvchange", "-an", volumeGroup+"/"+name).Run()
}

func UnmountFilesystem(targetPath string) error {
	mounter := &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: utilexec.New()}

	dev, ref, err := mount.GetDeviceNameFromMount(mounter, targetPath)
	if err != nil {
		return fmt.Errorf("failed to get device from mnt: %s\nError: %v", targetPath, err)
	}

	// device has already been un-mounted, return successful
	if len(dev) == 0 || ref == 0 {
		klog.Warningf(
			"Warning: Unmount skipped because volume not mounted: %v",
			targetPath,
		)
		return nil
	}

	if pathExists, pathErr := mount.PathExists(targetPath); pathErr != nil {
		return fmt.Errorf("error checking if path exists: %v", pathErr)
	} else if !pathExists {
		klog.Warningf(
			"Warning: Unmount skipped because path does not exist: %v",
			targetPath,
		)
		return nil
	}

	if err = mounter.Unmount(targetPath); err != nil {
		return fmt.Errorf("failed to unmount: %v", err)
	}

	if err := os.Remove(targetPath); err != nil {
		klog.Errorf("lvm: failed to remove mount path vol err: %v", err)
	}

	klog.Infof("umount done path %v", targetPath)

	return nil
}

func MountFilesystem(name string, volumeGroup string, targetPath string, fsType string, mountOptions []string) error {
	if len(targetPath) == 0 {
		return fmt.Errorf("target path is empty")
	}

	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return fmt.Errorf("could not create dir {%q}, err: %v", targetPath, err)
	}

	devPath := GetVolumeDevPath(name, volumeGroup)

	currentMounts, err := mnt.GetMounts(devPath)
	if err != nil {
		klog.Errorf("can not get mounts for volume:%s dev %s err: %v",
			name, devPath, err.Error())
		return fmt.Errorf("GetMounts failed %s", err.Error())
	} else if len(currentMounts) >= 1 {
		// if device is already mounted at the mount point, return successful
		for _, mp := range currentMounts {
			if mp == targetPath {
				return nil
			}
		}
	}

	mounter := &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: utilexec.New()}
	return mounter.FormatAndMount(devPath, targetPath, fsType, mountOptions)
}

func GetVolumeDevPath(name, volumeGroup string) string {
	// LVM doubles the hiphen for the mapper device name
	// and uses single hiphen to separate volume group from volume
	vg := strings.Replace(volumeGroup, "-", "--", -1)

	lv := strings.Replace(name, "-", "--", -1)
	dev := "/dev/mapper/" + vg + "-" + lv

	return dev
}

func GetVolumeId(name, volumeGroup string) string {
	return volumeGroup + "/" + name
}

func GetVolumeNameAndGroup(volumeId string) (string, string) {
	return strings.Split(volumeId, "/")[1], strings.Split(volumeId, "/")[0]
}

func ValidateName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("name is empty")
	}

	if strings.Contains(name, "/") {
		return fmt.Errorf("name can not contain '/'")
	}

	return nil
}
