// +build linux

package devmapper

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"

	"github.com/nholuongut/storage/drivers"
	"github.com/nholuongut/storage/pkg/devicemapper"
	"github.com/nholuongut/storage/pkg/idtools"
	"github.com/nholuongut/storage/pkg/locker"
	"github.com/nholuongut/storage/pkg/mount"
	"github.com/nholuongut/storage/pkg/system"
	units "github.com/docker/go-units"
	"github.com/sirupsen/logrus"
)

func init() {
	graphdriver.Register("devicemapper", Init)
}

// Driver contains the device set mounted and the home directory
type Driver struct {
	*DeviceSet
	home    string
	uidMaps []idtools.IDMap
	gidMaps []idtools.IDMap
	ctr     *graphdriver.RefCounter
	locker  *locker.Locker
}

// Init creates a driver with the given home and the set of options.
func Init(home string, options graphdriver.Options) (graphdriver.Driver, error) {
	deviceSet, err := NewDeviceSet(home, true, options.DriverOptions, options.UIDMaps, options.GIDMaps)
	if err != nil {
		return nil, err
	}

	if err := mount.MakePrivate(home); err != nil {
		return nil, err
	}

	d := &Driver{
		DeviceSet: deviceSet,
		home:      home,
		uidMaps:   options.UIDMaps,
		gidMaps:   options.GIDMaps,
		ctr:       graphdriver.NewRefCounter(graphdriver.NewDefaultChecker()),
		locker:    locker.New(),
	}

	return graphdriver.NewNaiveDiffDriver(d, graphdriver.NewNaiveLayerIDMapUpdater(d)), nil
}

func (d *Driver) String() string {
	return "devicemapper"
}

// Status returns the status about the driver in a printable format.
// Information returned contains Pool Name, Data File, Metadata file, disk usage by
// the data and metadata, etc.
func (d *Driver) Status() [][2]string {
	s := d.DeviceSet.Status()

	status := [][2]string{
		{"Pool Name", s.PoolName},
		{"Pool Blocksize", units.HumanSize(float64(s.SectorSize))},
		{"Base Device Size", units.HumanSize(float64(s.BaseDeviceSize))},
		{"Backing Filesystem", s.BaseDeviceFS},
		{"Data file", s.DataFile},
		{"Metadata file", s.MetadataFile},
		{"Data Space Used", units.HumanSize(float64(s.Data.Used))},
		{"Data Space Total", units.HumanSize(float64(s.Data.Total))},
		{"Data Space Available", units.HumanSize(float64(s.Data.Available))},
		{"Metadata Space Used", units.HumanSize(float64(s.Metadata.Used))},
		{"Metadata Space Total", units.HumanSize(float64(s.Metadata.Total))},
		{"Metadata Space Available", units.HumanSize(float64(s.Metadata.Available))},
		{"Thin Pool Minimum Free Space", units.HumanSize(float64(s.MinFreeSpace))},
		{"Udev Sync Supported", fmt.Sprintf("%v", s.UdevSyncSupported)},
		{"Deferred Removal Enabled", fmt.Sprintf("%v", s.DeferredRemoveEnabled)},
		{"Deferred Deletion Enabled", fmt.Sprintf("%v", s.DeferredDeleteEnabled)},
		{"Deferred Deleted Device Count", fmt.Sprintf("%v", s.DeferredDeletedDeviceCount)},
	}
	if len(s.DataLoopback) > 0 {
		status = append(status, [2]string{"Data loop file", s.DataLoopback})
	}
	if len(s.MetadataLoopback) > 0 {
		status = append(status, [2]string{"Metadata loop file", s.MetadataLoopback})
	}
	if vStr, err := devicemapper.GetLibraryVersion(); err == nil {
		status = append(status, [2]string{"Library Version", vStr})
	}
	return status
}

// Metadata returns a map of information about the device.
func (d *Driver) Metadata(id string) (map[string]string, error) {
	m, err := d.DeviceSet.exportDeviceMetadata(id)

	if err != nil {
		return nil, err
	}

	metadata := make(map[string]string)
	metadata["DeviceId"] = strconv.Itoa(m.deviceID)
	metadata["DeviceSize"] = strconv.FormatUint(m.deviceSize, 10)
	metadata["DeviceName"] = m.deviceName
	return metadata, nil
}

// Cleanup unmounts a device.
func (d *Driver) Cleanup() error {
	err := d.DeviceSet.Shutdown(d.home)

	if err2 := mount.Unmount(d.home); err == nil {
		err = err2
	}

	return err
}

// CreateFromTemplate creates a layer with the same contents and parent as another layer.
func (d *Driver) CreateFromTemplate(id, template string, templateIDMappings *idtools.IDMappings, parent string, parentIDMappings *idtools.IDMappings, opts *graphdriver.CreateOpts, readWrite bool) error {
	return d.Create(id, template, opts)
}

// CreateReadWrite creates a layer that is writable for use as a container
// file system.
func (d *Driver) CreateReadWrite(id, parent string, opts *graphdriver.CreateOpts) error {
	return d.Create(id, parent, opts)
}

// Create adds a device with a given id and the parent.
func (d *Driver) Create(id, parent string, opts *graphdriver.CreateOpts) error {
	var storageOpt map[string]string
	if opts != nil {
		storageOpt = opts.StorageOpt
	}

	if err := d.DeviceSet.AddDevice(id, parent, storageOpt); err != nil {
		return err
	}

	return nil
}

// Remove removes a device with a given id, unmounts the filesystem.
func (d *Driver) Remove(id string) error {
	d.locker.Lock(id)
	defer d.locker.Unlock(id)
	if !d.DeviceSet.HasDevice(id) {
		// Consider removing a non-existing device a no-op
		// This is useful to be able to progress on container removal
		// if the underlying device has gone away due to earlier errors
		return nil
	}

	// This assumes the device has been properly Get/Put:ed and thus is unmounted
	if err := d.DeviceSet.DeleteDevice(id, false); err != nil {
		return fmt.Errorf("failed to remove device %s: %v", id, err)
	}
	return system.EnsureRemoveAll(path.Join(d.home, "mnt", id))
}

// Get mounts a device with given id into the root filesystem
func (d *Driver) Get(id string, options graphdriver.MountOpts) (string, error) {
	d.locker.Lock(id)
	defer d.locker.Unlock(id)
	mp := path.Join(d.home, "mnt", id)
	rootFs := path.Join(mp, "rootfs")
	if count := d.ctr.Increment(mp); count > 1 {
		return rootFs, nil
	}

	uid, gid, err := idtools.GetRootUIDGID(d.uidMaps, d.gidMaps)
	if err != nil {
		d.ctr.Decrement(mp)
		return "", err
	}

	// Create the target directories if they don't exist
	if err := idtools.MkdirAllAs(path.Join(d.home, "mnt"), 0755, uid, gid); err != nil && !os.IsExist(err) {
		d.ctr.Decrement(mp)
		return "", err
	}
	if err := idtools.MkdirAs(mp, 0755, uid, gid); err != nil && !os.IsExist(err) {
		d.ctr.Decrement(mp)
		return "", err
	}

	// Mount the device
	if err := d.DeviceSet.MountDevice(id, mp, options); err != nil {
		d.ctr.Decrement(mp)
		return "", err
	}

	if err := idtools.MkdirAllAs(rootFs, 0755, uid, gid); err != nil && !os.IsExist(err) {
		d.ctr.Decrement(mp)
		d.DeviceSet.UnmountDevice(id, mp)
		return "", err
	}

	idFile := path.Join(mp, "id")
	if _, err := os.Stat(idFile); err != nil && os.IsNotExist(err) {
		// Create an "id" file with the container/image id in it to help reconstruct this in case
		// of later problems
		if err := ioutil.WriteFile(idFile, []byte(id), 0600); err != nil {
			d.ctr.Decrement(mp)
			d.DeviceSet.UnmountDevice(id, mp)
			return "", err
		}
	}

	return rootFs, nil
}

// Put unmounts a device and removes it.
func (d *Driver) Put(id string) error {
	d.locker.Lock(id)
	defer d.locker.Unlock(id)
	mp := path.Join(d.home, "mnt", id)
	if count := d.ctr.Decrement(mp); count > 0 {
		return nil
	}
	err := d.DeviceSet.UnmountDevice(id, mp)
	if err != nil {
		logrus.Errorf("devmapper: Error unmounting device %s: %s", id, err)
	}
	return err
}

// Exists checks to see if the device exists.
func (d *Driver) Exists(id string) bool {
	return d.DeviceSet.HasDevice(id)
}

// AdditionalImageStores returns additional image stores supported by the driver
func (d *Driver) AdditionalImageStores() []string {
	return nil
}
