package main

import (
	"os"
	"strings"

	"github.com/jaxxstorm/flexvolume"
	"github.com/kolyshkin/goploop-cli"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "ploop flexvolume"
	app.Usage = "Mount ploop volumes in kubernetes using the flexvolume driver"
	app.Commands = flexvolume.Commands(Ploop{})
	app.CommandNotFound = flexvolume.CommandNotFound
	app.Authors = []cli.Author{
		cli.Author{
			Name: "Lee Briggs",
		},
	}
	app.Version = "0.1a"
	app.Run(os.Args)
}

type Ploop struct{}

func (p Ploop) Init() flexvolume.Response {
	return flexvolume.Response{
		Status:  flexvolume.StatusSuccess,
		Message: "Ploop is available",
	}
}

func (p Ploop) GetVolumeName(options map[string]string) flexvolume.Response {
	if options["volumePath"] == "" {
		return flexvolume.Response{
			Status:     flexvolume.StatusFailure,
			Message:    "Must specify a volume path",
			VolumeName: "unknown",
		}
	}

	if options["volumeId"] == "" {
		return flexvolume.Response{
			Status:  flexvolume.StatusFailure,
			Message: "Must specify a volume id",
		}
	}

	return flexvolume.Response{
		Status:     flexvolume.StatusSuccess,
		VolumeName: options["volumePath"] + "/" + options["volumeId"],
	}
}

func (p Ploop) Attach(options map[string]string) flexvolume.Response {

	if options["volumePath"] == "" {
		return flexvolume.Response{
			Status:  flexvolume.StatusFailure,
			Message: "Must specify a volume path",
		}
	}

	if options["volumeId"] == "" {
		return flexvolume.Response{
			Status:  flexvolume.StatusFailure,
			Message: "Must specify a volume id",
		}
	}

	if _, err := os.Stat(options["volumePath"] + "/" + options["volumeId"] + "/" + "DiskDescriptor.xml"); err == nil {
		return flexvolume.Response{
			Status:  flexvolume.StatusSuccess,
			Message: "Volume is found",
			Device:  options["volumePath"] + "/" + options["volumeId"],
		}
	}

	return flexvolume.Response{
		Status:  flexvolume.StatusFailure,
		Message: "Volume does not exist",
	}
}

func (p Ploop) Detach(volumeName string) flexvolume.Response {

	device := strings.Replace(volumeName, "~", "/", -1)

	// open the disk descriptor first
	volume, err := ploop.Open(device + "/" + "DiskDescriptor.xml")
	if err != nil {
		return flexvolume.Response{
			Status:  flexvolume.StatusFailure,
			Message: err.Error(),
		}
	}
	defer volume.Close()

	if m, _ := volume.IsMounted(); m {
		if err := volume.Umount(); err != nil {
			return flexvolume.Response{
				Status:  flexvolume.StatusFailure,
				Message: "Unable to detach ploop volume",
				Device:  device,
			}
		}
	}
	return flexvolume.Response{
		Status:  flexvolume.StatusSuccess,
		Message: "Successfully detached the ploop volume",
		Device:  device,
	}
}

func (p Ploop) MountDevice(target string, options map[string]string) flexvolume.Response {
	// make the target directory we're going to mount to
	err := os.MkdirAll(target, 0755)
	if err != nil {
		return flexvolume.Response{
			Status:  flexvolume.StatusFailure,
			Message: err.Error(),
		}
	}

	// open the disk descriptor first
	volume, err := ploop.Open(options["volumePath"] + "/" + options["volumeId"] + "/" + "DiskDescriptor.xml")
	if err != nil {
		return flexvolume.Response{
			Status:  flexvolume.StatusFailure,
			Message: err.Error(),
		}
	}
	defer volume.Close()

	if m, _ := volume.IsMounted(); !m {
		// If it's mounted, let's mount it!

		readonly := false
		if options["kubernetes.io/readwrite"] == "ro" {
			readonly = true
		}

		mp := ploop.MountParam{Target: target, Readonly: readonly}

		dev, err := volume.Mount(&mp)
		if err != nil {
			return flexvolume.Response{
				Status:  flexvolume.StatusFailure,
				Message: err.Error(),
				Device:  dev,
			}
		}

		return flexvolume.Response{
			Status:  flexvolume.StatusSuccess,
			Message: "Successfully mounted the ploop volume",
		}
	} else {

		return flexvolume.Response{
			Status:  flexvolume.StatusSuccess,
			Message: "Ploop volume already mounted",
		}

	}
}

func (p Ploop) UnmountDevice(mount string) flexvolume.Response {
	if err := ploop.UmountByMount(mount); err != nil {
		return flexvolume.Response{
			Status:  flexvolume.StatusFailure,
			Message: err.Error(),
		}
	}

	return flexvolume.Response{
		Status:  flexvolume.StatusSuccess,
		Message: "Successfully unmounted the ploop volume",
	}
}
