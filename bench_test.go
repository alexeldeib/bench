package bench_test

import (
	"crypto/sha256"
	"encoding/json"
	"sync"
	"testing"

	fuzz "github.com/google/gofuzz"
)

type cache struct {
	mu   sync.RWMutex
	data map[string]*Metadata
}

func newCache() *cache {
	return &cache{
		data: make(map[string]*Metadata),
	}
}

func (c *cache) iterate(fn func(*Metadata)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, datum := range c.data {
		fn(datum)
	}
}

func (c *cache) fan(fn func(Metadata), concurrency int) {
	ch := make(chan Metadata)

	var wg sync.WaitGroup

	wg.Add(concurrency)
	for n := concurrency; n > 0; n-- {
		go drain(ch, &wg, fn)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, datum := range c.data {
		ch <- *datum
	}

	close(ch)
	wg.Wait()
}

func drain(ch chan Metadata, wg *sync.WaitGroup, fn func(Metadata)) {
	for datum := range ch {
		fn(datum)
	}
	wg.Done()
}

func (c *cache) load(count int) error {
	f := fuzz.New()
	for i := count; i > 0; i-- {
		datum := &Metadata{}
		f.Fuzz(datum)
		h := sha256.New()
		raw, err := json.Marshal(datum)
		if err != nil {
			return err
		}
		h.Write(raw)
		c.data[string(h.Sum(nil))] = datum
	}
	return nil
}

func BenchmarkSharedMap(b *testing.B) {
	c := newCache()
	c.load(10000)
	for i := 0; i < b.N; i++ {
		c.iterate(func(datum *Metadata) {
			raw, err := json.Marshal(datum)
			if err != nil {
				b.Errorf("failed to marshal customer resource: %v", err)
			}
			dst := &Metadata{}
			if err := json.Unmarshal(raw, dst); err != nil {
				b.Errorf("failed to marshal customer resource: %v", err)
			}
		})
	}
}

func BenchmarkChannel1(b *testing.B) {
	c := newCache()
	c.load(10000)
	for i := 0; i < b.N; i++ {
		c.fan(func(datum Metadata) {
			raw, err := json.Marshal(datum)
			if err != nil {
				b.Errorf("failed to marshal customer resource: %v", err)
			}
			dst := &Metadata{}
			if err := json.Unmarshal(raw, dst); err != nil {
				b.Errorf("failed to marshal customer resource: %v", err)
			}
		}, 1)
	}
}

func BenchmarkChannel5(b *testing.B) {
	c := newCache()
	c.load(10000)
	for i := 0; i < b.N; i++ {
		c.fan(func(datum Metadata) {
			raw, err := json.Marshal(datum)
			if err != nil {
				b.Errorf("failed to marshal customer resource: %v", err)
			}
			dst := Metadata{}
			if err := json.Unmarshal(raw, &dst); err != nil {
				b.Errorf("failed to marshal customer resource: %v", err)
			}
		}, 5)
	}
}

// Below this line is only type definitions to describe Azure IMDS 2019-08-15
// It's a fairly large struct for fuzzing purposes

type Metadata struct {
	Compute    Compute           `json:"compute"`
	Network    Network           `json:"network"`
	ParsedTags map[string]string `json:"parsedTags"`
}

type Plan struct {
	Name      string `json:"name"`
	Product   string `json:"product"`
	Publisher string `json:"publisher"`
}

type PublicKeys struct {
	KeyData string `json:"keyData"`
	Path    string `json:"path"`
}

type Image struct {
	URI string `json:"uri"`
}

type ManagedDisk struct {
	ID                 string `json:"id"`
	StorageAccountType string `json:"storageAccountType"`
}

type VHD struct {
	URI string `json:"uri"`
}

type DataDisks struct {
	Caching                 string      `json:"caching"`
	CreateOption            string      `json:"createOption"`
	DiskSizeGB              string      `json:"diskSizeGB"`
	Image                   Image       `json:"image"`
	LUN                     string      `json:"lun"`
	ManagedDisk             ManagedDisk `json:"managedDisk"`
	Name                    string      `json:"name"`
	VHD                     VHD         `json:"vhd"`
	WriteAcceleratorEnabled string      `json:"writeAcceleratorEnabled"`
}

type ImageReference struct {
	ID        string `json:"id"`
	Offer     string `json:"offer"`
	Publisher string `json:"publisher"`
	Sku       string `json:"sku"`
	Version   string `json:"version"`
}

type DiffDiskSettings struct {
	Option string `json:"option"`
}

type EncryptionSettings struct {
	Enabled string `json:"enabled"`
}

type OsDisk struct {
	Caching                 string             `json:"caching"`
	CreateOption            string             `json:"createOption"`
	DiffDiskSettings        DiffDiskSettings   `json:"diffDiskSettings"`
	DiskSizeGB              string             `json:"diskSizeGB"`
	EncryptionSettings      EncryptionSettings `json:"encryptionSettings"`
	Image                   Image              `json:"image"`
	ManagedDisk             ManagedDisk        `json:"managedDisk"`
	Name                    string             `json:"name"`
	OsType                  string             `json:"osType"`
	VHD                     VHD                `json:"vhd"`
	WriteAcceleratorEnabled string             `json:"writeAcceleratorEnabled"`
}

type StorageProfile struct {
	DataDisks      []DataDisks    `json:"dataDisks"`
	ImageReference ImageReference `json:"imageReference"`
	OsDisk         OsDisk         `json:"osDisk"`
}

type Compute struct {
	AzEnvironment        string         `json:"azEnvironment"`
	CustomData           string         `json:"customData"`
	Location             string         `json:"location"`
	Name                 string         `json:"name"`
	Offer                string         `json:"offer"`
	OsType               string         `json:"osType"`
	PlacementGroupID     string         `json:"placementGroupId"`
	Plan                 Plan           `json:"plan"`
	PlatformFaultDomain  string         `json:"platformFaultDomain"`
	PlatformUpdateDomain string         `json:"platformUpdateDomain"`
	Provider             string         `json:"provider"`
	PublicKeys           []PublicKeys   `json:"publicKeys"`
	Publisher            string         `json:"publisher"`
	ResourceGroupName    string         `json:"resourceGroupName"`
	ResourceID           string         `json:"resourceId"`
	Sku                  string         `json:"sku"`
	StorageProfile       StorageProfile `json:"storageProfile"`
	SubscriptionID       string         `json:"subscriptionId"`
	Tags                 string         `json:"tags"`
	Version              string         `json:"version"`
	VMID                 string         `json:"vmId"`
	VMScaleSetName       string         `json:"vmScaleSetName"`
	VMSize               string         `json:"vmSize"`
	Zone                 string         `json:"zone"`
}

type IPAddress struct {
	PrivateIPAddress string `json:"privateIpAddress"`
	PublicIPAddress  string `json:"publicIpAddress"`
}

type Subnet struct {
	Address string `json:"address"`
	Prefix  string `json:"prefix"`
}

type IPv4 struct {
	IPAddress []IPAddress `json:"ipAddress"`
	Subnet    []Subnet    `json:"subnet"`
}

type IPv6 struct {
	IPAddress []IPAddress `json:"ipAddress"`
}

type Interface struct {
	IPv4       IPv4   `json:"ipv4"`
	IPv6       IPv6   `json:"ipv6"`
	MacAddress string `json:"macAddress"`
}

type Network struct {
	Interface []Interface `json:"interface"`
}
