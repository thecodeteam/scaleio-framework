package kvstore

import (
	"errors"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/libkv"
	store "github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"github.com/docker/libkv/store/consul"
	"github.com/docker/libkv/store/etcd"
	"github.com/docker/libkv/store/zookeeper"
	xplatform "github.com/dvonthenen/goxplatform"

	"github.com/codedellemc/scaleio-framework/scaleio-scheduler/config"
	"github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

const (
	rootKey           = "scaleio-framework"
	mesosServerPrefix = "scaleio-s-"
	mesosClientPrefix = "scaleio-c-"

	//SdsModeAll is both client and server
	SdsModeAll = 1

	//SdsModeClient is client only
	SdsModeClient = 2

	//SdsModeServer is server only
	SdsModeServer = 3
)

var (
	//ErrInvalidKeyValue The Key/Value returned is nil
	ErrInvalidKeyValue = errors.New("The Key/Value returned is nil")

	//ErrStoreType Invalid store type
	ErrStoreType = errors.New("Invalid store type")
)

//KvStore representation a KeyValue Store
type KvStore struct {
	Config  *config.Config
	Store   store.Store
	RootKey string
}

//Device representation
type Device struct {
	Name   string
	Delete bool
	Add    bool
}

//StoragePool representation
type StoragePool struct {
	Name    string
	Devices map[string]*Device
	Delete  bool
	Add     bool
}

//Sds representation
type Sds struct {
	Name   string
	Mode   int
	Delete bool
	Add    bool
}

//ProtectionDomain representation
type ProtectionDomain struct {
	Name   string
	Pools  map[string]*StoragePool
	Sdss   map[string]*Sds
	Delete bool
	Add    bool
}

//Metadata representation
type Metadata struct {
	ProtectionDomains map[string]*ProtectionDomain
}

//NewKvStore generates a new KvStore object
func NewKvStore(cfg *config.Config) (*KvStore, error) {
	storeCfg := store.Config{
		ConnectionTimeout: 10 * time.Second,
	}

	backend := store.Backend(cfg.Store)
	log.Debugln("backend:", backend)

	endpoints := []string{cfg.StoreURI}
	if len(cfg.StoreURI) == 0 {
		var err error
		endpoints, err = readZookeeperFile()
		if err != nil {
			log.Errorln("Invalid libkv store type.")
			return nil, err
		}
	}
	log.Debugln("endpoints:", endpoints)

	switch backend {
	case store.CONSUL:
		consul.Register()
	case store.ZK:
		zookeeper.Register()
	case store.ETCD:
		etcd.Register()
	case store.BOLTDB:
		boltdb.Register()
		storeCfg.Bucket = "/tmp/boltdb"
	default:
		log.Errorln("Invalid libkv store type.")
		return nil, ErrStoreType
	}

	myStore, err := libkv.NewStore(backend, endpoints, &storeCfg)
	if err != nil {
		log.Errorln("Unable to initialize the Store")
		return nil, err
	}

	//sets the root for this framework instance
	myRootKey := xplatform.GetInstance().Fs.AppendSlash(rootKey) + cfg.Role
	log.Debugln("myRootKey:", myRootKey)

	pair, err := myStore.Get(myRootKey + "/version")
	if pair != nil && err != nil {
		log.Infoln(pair.Key, "=", pair.Value)
		//TODO: if you need to some form of metadata update, do it here
	}

	//record the current version for the metadata
	myStore.Put(myRootKey, []byte(""), nil)
	err = myStore.Put(myRootKey+"/version", []byte(strconv.Itoa(config.VersionInt)), nil)
	if err != nil {
		log.Errorln("Failed to set version on store:", err)
		return nil, err
	} else {
		log.Debugln("Successfully set KV Store version to:", config.VersionInt)
	}

	myKvStore := &KvStore{
		Config:  cfg,
		Store:   myStore,
		RootKey: myRootKey,
	}

	return myKvStore, nil
}

func readZookeeperFile() ([]string, error) {
	dat, err := ioutil.ReadFile("/etc/mesos/zk")
	if err != nil {
		return nil, err
	}
	zkStr := strings.TrimSpace(string(dat))

	var tmpZkStr string
	lindex := strings.LastIndex(zkStr, "/")
	if lindex == -1 {
		tmpZkStr = zkStr
	} else {
		tmpZkStr = zkStr[:lindex]
	}

	lindex = strings.LastIndex(tmpZkStr, "/")
	if lindex != -1 {
		tmpZkStr = tmpZkStr[lindex+1:]
	}

	return strings.Split(tmpZkStr, ","), nil
}

func (kv *KvStore) deleteTree(dir string) error {
	items, err := kv.Store.List(dir)
	if err != nil || len(items) == 0 {
		//assume this is a key not a dir
		err = kv.Store.Delete(dir)
		if err != nil {
			log.Debugln("Delete(", dir, ") Failed. Err:", err)
			return err
		}
		log.Debugln("Delete Key:", dir)
		return nil
	}

	for _, item := range items {
		newDir := dir + "/" + item.Key

		err := kv.Store.DeleteTree(newDir)
		if err != nil && strings.Contains(err.Error(), "node has children") {
			log.Debugln("Recurse into Dir:", newDir)
			err = kv.deleteTree(newDir)
			if err != nil {
				return err
			}
		} else {
			err = kv.Store.Delete(newDir)
			if err != nil {
				log.Debugln("Delete(", newDir, ") Failed. Err:", err)
				return err
			}
			log.Debugln("Delete Key:", newDir)
		}
	}

	err = kv.Store.DeleteTree(dir)
	if err != nil {
		log.Debugln("DeleteTree(", dir, ") Failed. Err:", err)
		return err
	}
	log.Debugln("Deleted Dir:", dir)
	return nil
}

func (kv *KvStore) dumpTree(dir string) {
	items, err := kv.Store.List(dir)
	if err != nil || len(items) == 0 {
		item, err := kv.Store.Get(dir)
		if err != nil {
			log.Debugln("Get(", dir, ") Err:", err)
			return
		}
		if item != nil {
			log.Debugln("Key", dir, "= Value", string(item.Value))
		}
		return
	}

	for _, item := range items {
		newDir := dir + "/" + item.Key

		_, err := kv.Store.List(newDir)
		if err == nil {
			kv.dumpTree(newDir)
		} else {
			if err != nil {
				log.Debugln("List(", newDir, ") Err:", err)
			}
			log.Debugln("Key", newDir, "= Value", string(item.Value))
		}
	}
}

//DeleteStore deletes all ScaleIO Framework metadata
func (kv *KvStore) DeleteStore() {
	log.Debugln("Calling DeleteStore...")
	kv.deleteTree(kv.RootKey)
}

//DumpStore prints out the ScaleIO Framework metadata
func (kv *KvStore) DumpStore() {
	log.Debugln("Calling DumpStore...")
	kv.dumpTree(kv.RootKey)
}

//UserKeyValue returns debug tool for modifying keyvalue pairs
func (kv *KvStore) UserKeyValue(key string, value string) error {
	log.Infoln("Key:", key)
	log.Infoln("Value:", value)

	err := kv.Store.Put(key, []byte(value), nil)
	if err != nil {
		log.Errorln("Put err:", err)
		return err
	}

	log.Infoln("UserKeyValue Succeeded")
	return nil
}

//UserDeleteKey returns debug tool for modifying keyvalue pairs
func (kv *KvStore) UserDeleteKey(key string) error {
	log.Infoln("Key:", key)

	err := kv.deleteTree(key)
	if err != nil {
		log.Debugln("deleteTree err:", err)
		return err
	}

	log.Infoln("UserDeleteKey Succeeded")
	return nil
}

//GetKvStoreVersion returns the metadata version
func (kv *KvStore) GetKvStoreVersion() string {
	pair, err := kv.Store.Get(kv.RootKey + "/version")
	if pair != nil && err != nil {
		log.Debugln(pair.Key, "=", pair.Value)
	}
	return string(pair.Value)
}

//GetConfigured returns if the ScaleIO is configured
func (kv *KvStore) GetConfigured() bool {
	pair, err := kv.Store.Get(kv.RootKey + "/configuration/configured")
	if err != nil {
		log.Errorln("GetConfigured Err:", err)
		return false
	}
	if pair == nil {
		log.Errorln("pair == nil. Err:", ErrInvalidKeyValue)
		return false
	}

	log.Debugln("Value:", string(pair.Value))
	if string(pair.Value) == "true" {
		log.Debugln("GetConfigured = TRUE")
		return true
	}
	log.Debugln("GetConfigured = FALSE")
	return false
}

//SetConfigured set the ScaleIO node to configured
func (kv *KvStore) SetConfigured() error {
	rootConfig := kv.RootKey + "/configuration"
	kv.Store.Put(rootConfig, []byte(""), nil)

	err := kv.Store.Put(rootConfig+"/configured", []byte(string("true")), nil)
	if err != nil {
		log.Debugln("SetConfigured err:", err)
		return err
	}
	log.Debugln("SetConfigured Succeeded")
	return nil
}

//GetMdmNodes returns the pri, sec, tb mdms nodes in the Store
func (kv *KvStore) GetMdmNodes() (string, string, string) {
	var pri, sec, tb string

	pairPri, errPri := kv.Store.Get(kv.RootKey + "/configuration/primary")
	if errPri != nil {
		log.Debugln("store.Get(primary) err:", errPri)
	}
	if pairPri != nil {
		log.Debugln(pairPri.Key, "=", string(pairPri.Value))
		pri = string(pairPri.Value)
	} else {
		log.Debugln("pairPri is empty")
	}
	pairSec, errSec := kv.Store.Get(kv.RootKey + "/configuration/secondary")
	if errSec != nil {
		log.Debugln("store.Get(secondary) err:", errSec)
	}
	if pairSec != nil {
		log.Debugln(pairSec.Key, "=", string(pairSec.Value))
		sec = string(pairSec.Value)
	} else {
		log.Debugln("pairSec is empty")
	}
	pairTb, errTb := kv.Store.Get(kv.RootKey + "/configuration/tiebreaker")
	if errTb != nil {
		log.Debugln("store.Get(tiebreaker) err:", errTb)
	}
	if pairTb != nil {
		log.Debugln(pairTb.Key, "=", string(pairTb.Value))
		tb = string(pairTb.Value)
	} else {
		log.Debugln("pairTb is empty")
	}

	return pri, sec, tb
}

//GetNodeInfo returns all metadata for a give node
func (kv *KvStore) GetNodeInfo(nodeID string) (int, int, error) {
	log.Debugln("GetNodeInfo ENTER")
	log.Debugln("nodeID:", nodeID)

	if len(nodeID) == 0 {
		log.Errorln("nodeID is empty. Return error.")
		log.Debugln("GetNodeInfo LEAVE")
		return 0, 0, ErrInvalidKeyValue
	}

	var persona, state int

	pairPersona, errPersona := kv.Store.Get(kv.RootKey + "/configuration/" + nodeID + "/persona")
	if errPersona != nil {
		log.Errorln("Store.Get(persona) err:", errPersona)
		log.Debugln("GetNodeInfo LEAVE")
		return 0, 0, errPersona
	}
	if pairPersona == nil {
		log.Errorln("pairPersona = nil. Return error.")
		log.Debugln("GetNodeInfo LEAVE")
		return 0, 0, ErrInvalidKeyValue
	}
	log.Debugln(pairPersona.Key, "=", string(pairPersona.Value))
	persona, errPersona = strconv.Atoi(string(pairPersona.Value))
	if errPersona != nil {
		log.Errorln("Atoi err:", errPersona)
		log.Debugln("GetNodeInfo LEAVE")
		return 0, 0, errPersona
	}

	pairState, errState := kv.Store.Get(kv.RootKey + "/configuration/" + nodeID + "/state")
	if errState != nil {
		log.Errorln("Store.Get(state) err:", errState)
		log.Debugln("GetNodeInfo LEAVE")
		return 0, 0, errState
	}
	if pairState == nil {
		log.Errorln("pairPersona = nil. Return error.")
		log.Debugln("GetNodeInfo LEAVE")
		return 0, 0, ErrInvalidKeyValue
	}
	log.Debugln(pairState.Key, "=", string(pairState.Value))
	state, errState = strconv.Atoi(string(pairState.Value))
	if errState != nil {
		log.Errorln("Atoi err:", errState)
		log.Debugln("GetNodeInfo LEAVE")
		return 0, 0, errState
	}

	log.Debugln("GetNodeInfo Succeeded. persona:", persona, "state:", state)
	log.Debugln("GetNodeInfo LEAVE")
	return persona, state, nil
}

//SetNodeInfo sets all metadata for a given node
func (kv *KvStore) SetNodeInfo(nodeID string, persona int, state int) error {
	log.Debugln("SetNodeInfo ENTER")
	log.Debugln("persona:", persona)
	log.Debugln("state:", state)

	rootConfig := kv.RootKey + "/configuration"
	kv.Store.Put(rootConfig, []byte(""), nil)
	rootNode := kv.RootKey + "/configuration/" + nodeID
	kv.Store.Put(rootNode, []byte(""), nil)

	if persona != -1 {
		log.Debugln("Changing persona to", persona)
		err := kv.Store.Put(rootNode+"/persona", []byte(strconv.Itoa(persona)), nil)
		if err != nil {
			log.Errorln("Failed to set version on store:", err)
			log.Debugln("SetNodeInfo LEAVE")
			return err
		}
	} else {
		log.Debugln("Skip changing persona")
	}
	if state != -1 {
		log.Debugln("Changing state to", state)
		err := kv.Store.Put(rootNode+"/state", []byte(strconv.Itoa(state)), nil)
		if err != nil {
			log.Errorln("Failed to set version on store:", err)
			log.Debugln("SetNodeInfo LEAVE")
			return err
		}
	} else {
		log.Debugln("Skip changing state")
	}

	switch persona {
	case types.PersonaMdmPrimary:
		log.Debugln("Saving primary MDM node ID:", nodeID)
		err := kv.Store.Put(rootConfig+"/primary", []byte(nodeID), nil)
		if err != nil {
			log.Errorln("Failed to set primary on store:", err)
			log.Debugln("SetNodeInfo LEAVE")
			return err
		}

	case types.PersonaMdmSecondary:
		log.Debugln("Saving secondary MDM node ID:", nodeID)
		err := kv.Store.Put(rootConfig+"/secondary", []byte(nodeID), nil)
		if err != nil {
			log.Errorln("Failed to set secondary on store:", err)
			log.Debugln("SetNodeInfo LEAVE")
			return err
		}

	case types.PersonaTb:
		log.Debugln("Saving tiebreaker MDM node ID:", nodeID)
		err := kv.Store.Put(rootConfig+"/tiebreaker", []byte(nodeID), nil)
		if err != nil {
			log.Errorln("Failed to set tiebreaker on store:", err)
			log.Debugln("SetNodeInfo LEAVE")
			return err
		}
	}

	log.Debugln("SetNodeInfo Succeeded")
	log.Debugln("SetNodeInfo LEAVE")
	return nil
}

//GetMetadata gets all domains/pools for a given node
func (kv *KvStore) GetMetadata(nodeID string) (*Metadata, error) {
	log.Debugln("GetMetadata LEAVE")
	log.Debugln("nodeID:", nodeID)

	rootNode := kv.RootKey + "/configuration/" + nodeID + "/domains"

	domainNode := rootNode + "/domains"
	pairDomain, errDomain := kv.Store.Get(domainNode)
	if errDomain != nil {
		log.Errorln("Store.Get(domain) err:", errDomain)
		log.Debugln("GetMetadata LEAVE")
		return nil, errDomain
	}
	if pairDomain == nil {
		log.Errorln("pairDomain = nil. Return error.")
		log.Debugln("GetMetadata LEAVE")
		return nil, ErrInvalidKeyValue
	}
	log.Debugln(pairDomain.Key, "=", string(pairDomain.Value))

	md := new(Metadata)
	md.ProtectionDomains = make(map[string]*ProtectionDomain)

	//Domains
	domainList := strings.Split(string(pairDomain.Value), ",")
	for _, domain := range domainList {

		pd := new(ProtectionDomain)
		pd.Name = domain
		pd.Pools = make(map[string]*StoragePool)
		pd.Sdss = make(map[string]*Sds)
		md.ProtectionDomains[domain] = pd

		//Sds
		sdsNode := rootNode + "/" + domain + "/sdss"
		pairSds, errSds := kv.Store.Get(sdsNode)
		if errSds != nil {
			log.Errorln("Store.Get(sds) err:", errSds)
			log.Debugln("GetMetadata LEAVE")
			return nil, errSds
		}
		if pairSds == nil {
			log.Errorln("pairSds = nil. Return error.")
			log.Debugln("GetMetadata LEAVE")
			return nil, ErrInvalidKeyValue
		}
		log.Debugln(pairSds.Key, "=", string(pairSds.Value))

		sdsList := strings.Split(string(pairSds.Value), ",")
		for index, sds := range sdsList {
			s := new(Sds)
			s.Name = sds
			if len(sdsList) == 1 {
				s.Mode = SdsModeAll
			} else {
				s.Mode = index + 2
			}
			pd.Sdss[sds] = s
		}

		//Pools
		poolNode := rootNode + "/" + domain + "/pools"
		pairPool, errPool := kv.Store.Get(poolNode)
		if errPool != nil {
			log.Errorln("Store.Get(pool) err:", errPool)
			log.Debugln("GetMetadata LEAVE")
			return nil, errPool
		}
		if pairPool == nil {
			log.Errorln("pairPool = nil. Return error.")
			log.Debugln("GetMetadata LEAVE")
			return nil, ErrInvalidKeyValue
		}
		log.Debugln(pairPool.Key, "=", string(pairPool.Value))

		poolList := strings.Split(string(pairPool.Value), ",")
		for _, pool := range poolList {

			sp := new(StoragePool)
			sp.Name = pool
			sp.Devices = make(map[string]*Device)
			pd.Pools[pool] = sp

			//Devices
			deviceNode := rootNode + "/" + domain + "/" + pool
			pairDevice, errDevice := kv.Store.Get(deviceNode)
			if errDevice != nil {
				log.Errorln("Store.Get(device) err:", errDevice)
				log.Debugln("GetMetadata LEAVE")
				return nil, errDevice
			}
			if pairDevice == nil {
				log.Errorln("pairDevice = nil. Return error.")
				log.Debugln("GetMetadata LEAVE")
				return nil, ErrInvalidKeyValue
			}
			log.Debugln(pairDevice.Key, "=", string(pairDevice.Value))

			deviceList := strings.Split(string(pairDevice.Value), ",")
			for _, device := range deviceList {
				dev := new(Device)
				dev.Name = device
				sp.Devices[device] = dev
			}
		}
	}

	log.Debugln("GetMetadata Succeeded")
	log.Debugln("GetMetadata LEAVE")
	return md, nil
}

//SetMetadata sets all domains/pools for a given node
func (kv *KvStore) SetMetadata(nodeID string, metaData *Metadata) error {
	log.Debugln("SetMetadata ENTER")
	log.Debugln("nodeID:", nodeID)

	rootNode := kv.RootKey + "/configuration/" + nodeID + "/domains"
	errRoot := kv.Store.Put(rootNode, []byte(""), nil)
	if errRoot == nil {
		log.Debugln("Create dir", rootNode, "succeeded")
	} else {
		log.Errorln("Failed to create dir", rootNode, ". Err:", errRoot)
	}

	//Domain
	domainList := ""
	for _, domain := range metaData.ProtectionDomains {
		domainNode := rootNode + "/" + domain.Name
		log.Debugln("Enter", domainNode)

		if len(domain.Pools) == 0 {
			log.Debugln("PoolList is empty. Set Delete = true")
			domain.Delete = true
		}

		if domain.Delete {
			err := kv.deleteTree(domainNode)
			if err == nil {
				log.Debugln("DeleteTree", domainNode, "succeeded")
			} else {
				log.Warnln("DeleteTree", domainNode, ". Err:", err)
			}
			continue
		}

		if len(domainList) > 0 {
			domainList += ","
		}
		domainList += domain.Name

		errDomain := kv.Store.Put(domainNode, []byte(""), nil)
		if errDomain == nil {
			log.Debugln("Create dir", domainNode, "succeeded")
		} else {
			log.Errorln("Failed to create dir", domainNode, ". Err:", errDomain)
		}

		//Sds
		sdsList := ""
		for _, sds := range domain.Sdss {
			if len(sdsList) > 0 {
				sdsList += ","
			}
			sdsList += sds.Name
		}

		sdsNode := domainNode + "/sdss"
		errSds := kv.Store.Put(sdsNode, []byte(sdsList), nil)
		if errSds == nil {
			log.Debugln("Set", sdsNode, "=", sdsList, "succeeded")
		} else {
			log.Errorln("Failed to set", sdsNode, ". Err:", errSds)
		}

		//Pool
		poolList := ""
		for _, pool := range domain.Pools {
			poolNode := domainNode + "/" + pool.Name

			if len(pool.Devices) == 0 {
				log.Debugln("DeviceList is empty. Set Delete = true")
				pool.Delete = true
			}

			if pool.Delete {
				err := kv.Store.Delete(poolNode)
				if err == nil {
					log.Debugln("Delete", poolNode, "succeeded")
				} else {
					log.Warnln("Delete", poolNode, ". Err:", err)
				}
				continue
			}

			if len(poolList) > 0 {
				poolList += ","
			}
			poolList += pool.Name

			deviceList := ""
			for _, device := range pool.Devices {
				if device.Delete {
					continue
				}
				if len(deviceList) > 0 {
					deviceList += ","
				}
				deviceList += device.Name
			}

			errPool := kv.Store.Put(poolNode, []byte(deviceList), nil)
			if errPool == nil {
				log.Debugln("Set", poolNode, "=", deviceList, "succeeded")
			} else {
				log.Errorln("Failed to set", poolNode, ". Err:", errPool)
			}
		}

		if len(poolList) > 0 {
			poolsNode := domainNode + "/pools"
			errPools := kv.Store.Put(poolsNode, []byte(poolList), nil)
			if errPools == nil {
				log.Debugln("Set", poolsNode, "=", poolList, "succeeded")
			} else {
				log.Errorln("Failed to set", poolsNode, ". Err:", errPools)
			}
		}
	}

	if len(domainList) > 0 {
		domainsNode := rootNode + "/domains"
		errDomains := kv.Store.Put(domainsNode, []byte(domainList), nil)
		if errDomains == nil {
			log.Debugln("Set", domainsNode, "=", domainList, "succeeded")
		} else {
			log.Errorln("Failed to set", domainsNode, ". Err:", errDomains)
		}
	}

	log.Debugln("SetMetadata Succeeded")
	log.Debugln("SetMetadata LEAVE")
	return nil
}
