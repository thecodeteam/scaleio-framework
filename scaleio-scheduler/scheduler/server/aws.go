package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	goscaleio "github.com/codedellemc/goscaleio"

	"github.com/codedellemc/scaleio-framework/scaleio-scheduler/scheduler/common"
	kvstore "github.com/codedellemc/scaleio-framework/scaleio-scheduler/scheduler/kvstore"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

const (
	awsDefaultMaxRetries = 10
)

type pairDomainPool struct {
	Domain *goscaleio.ProtectionDomain
	Pools  []*goscaleio.StoragePool
}

type pairHost struct {
	Ec2Instance *ec2.Instance
	ScaleIONode *types.ScaleIONode
}

type instanceInfo struct {
	InstanceID       string `json:"instanceId,omitempty"`
	Region           string `json:"region,omitempty"`
	AvailabilityZone string `json:"availabilityZone,omitempty"`
}

func getInstanceInfo() (*instanceInfo, error) {
	log.Debugln("getInstanceInfo ENTER")
	url := "http://169.254.169.254/latest/dynamic/instance-identity/document"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorln("Error is HTTP NewRequest:", err)
		log.Debugln("getInstanceInfo LEAVE")
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorln("Error is HTTP Do:", err)
		log.Debugln("getInstanceInfo LEAVE")
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1048576))
	if err != nil {
		log.Errorln("Error is IO ReadAll:", err)
		log.Debugln("getInstanceInfo LEAVE")
		return nil, err
	}

	log.Debugln("Body: ", string(body))

	var state instanceInfo
	err = json.Unmarshal(body, &state)
	if err != nil {
		log.Errorln("Error is IO ReadAll:", err)
		log.Debugln("getInstanceInfo LEAVE")
		return nil, err
	}

	log.Debugln("getInstanceInfo Succeeded")
	log.Debugln("getInstanceInfo LEAVE")

	return &state, nil
}

func awsConnect(accessKey string, secretKey string) (*ec2.EC2, *instanceInfo, error) {
	//Get AWS region
	awsInfo, errInfo := getInstanceInfo()
	if errInfo != nil {
		log.Errorln("getInstanceInfo Failed. Err:", errInfo)
		log.Infoln("expandPools LEAVE")
		return nil, nil, errInfo
	}

	//Create AWS session
	sess, errAws := session.NewSession()
	if errAws != nil {
		log.Errorln("NewSession Failed. Err:", errAws)
		log.Infoln("expandPools LEAVE")
		return nil, nil, errAws
	}

	region := awsInfo.Region
	endpoint := fmt.Sprintf("ec2.%s.amazonaws.com", region)
	maxRetries := awsDefaultMaxRetries

	client := ec2.New(
		sess,
		&aws.Config{
			Region:     &region,
			Endpoint:   &endpoint,
			MaxRetries: &maxRetries,
			Credentials: credentials.NewChainCredentials(
				[]credentials.Provider{
					&credentials.StaticProvider{
						Value: credentials.Value{
							AccessKeyID:     accessKey,
							SecretAccessKey: secretKey,
						},
					},
					&credentials.EnvProvider{},
					&credentials.SharedCredentialsProvider{},
					&ec2rolecreds.EC2RoleProvider{
						Client: ec2metadata.New(sess),
					},
				},
			),
		},
	)

	return client, awsInfo, nil
}

func (s *RestServer) getInstanceList(client *ec2.EC2) ([]*pairHost, error) {
	log.Infoln("getInstanceList ENTER")

	filters := []*ec2.Filter{
		&ec2.Filter{
			Name:   aws.String("instance-state-name"),
			Values: []*string{aws.String("running")},
		},
	}
	request := ec2.DescribeInstancesInput{Filters: filters}
	result, errInst := client.DescribeInstances(&request)
	if errInst != nil {
		log.Errorln("DescribeInstances Failed. Err:", errInst)
		log.Infoln("getInstanceList LEAVE")
		return nil, errInst
	}

	var hostList []*pairHost
	hostList = make([]*pairHost, 0)

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			fmt.Println(*instance.PrivateDnsName)
			fmt.Println(*instance.PublicDnsName)
			node := s.findAwsInstanceInScaleIONode(instance)
			if node != nil {
				hostList = append(hostList, &pairHost{
					Ec2Instance: instance,
					ScaleIONode: node,
				})
			}
		}
	}

	log.Infoln("getInstanceList Succeeded")
	log.Infoln("getInstanceList LEAVE")
	return hostList, nil
}

func getVolume(client *ec2.EC2, awsInfo *instanceInfo, volumeID string) (*ec2.Volume, error) {
	log.Infoln("getVolume ENTER")

	// prepare filters
	filters := []*ec2.Filter{}

	filters = append(filters, &ec2.Filter{
		Name:   aws.String("availability-zone"),
		Values: []*string{aws.String(awsInfo.AvailabilityZone)},
	})
	filters = append(filters, &ec2.Filter{
		Name:   aws.String("volume-id"),
		Values: []*string{&volumeID},
	})

	// Prepare input
	dvInput := &ec2.DescribeVolumesInput{}

	dvInput.Filters = filters
	dvInput.VolumeIds = []*string{&volumeID}

	// Retrieve filtered volumes through EC2 API call
	resp, err := client.DescribeVolumes(dvInput)
	if err != nil {
		log.Infoln("DescribeVolumes Failed. Err:", err)
		log.Infoln("getVolume LEAVE")
		return &ec2.Volume{}, err
	}

	if len(resp.Volumes) != 1 {
		log.Infoln("DescribeVolumes Failed. Err:", err)
		log.Infoln("getVolume LEAVE")
		return &ec2.Volume{}, errors.New("Number of volumes is greater than 1")
	}

	log.Infoln("getVolume Succeeded")
	log.Infoln("getVolume LEAVE")
	return resp.Volumes[0], nil
}

func createVolume(client *ec2.EC2, awsInfo *instanceInfo, host *pairHost, growthSize int) error {
	log.Infoln("createVolume ENTER")

	encrypted := false
	size := int64(growthSize)
	volType := "gp2"

	createOptions := &ec2.CreateVolumeInput{
		Size:             &size,
		AvailabilityZone: &awsInfo.AvailabilityZone,
		Encrypted:        &encrypted,
		VolumeType:       &volType,
	}

	log.Infoln("Calling CreateVolume()...")
	resVolume, errCreate := client.CreateVolume(createOptions)
	if errCreate != nil {
		log.Errorln("CreateVolume Failed. Err:", errCreate)
		log.Infoln("createVolume LEAVE")
		return errCreate
	}
	log.Infoln("Volume:", resVolume)

	log.Infoln("Checking for CreateVolume status")
	for {
		time.Sleep(time.Duration(common.WaitForVolumeState) * time.Second)

		tmpVol, errVol := getVolume(client, awsInfo, *resVolume.VolumeId)
		if errVol != nil {
			log.Errorln("getVolume Failed. Err:", errVol)
			break
		}

		log.Infoln("CreateVolume - Volume:", *tmpVol.VolumeId, "State:", *tmpVol.State)

		if *tmpVol.State == ec2.VolumeStateAvailable ||
			*tmpVol.State == ec2.VolumeStateInUse {
			log.Infoln("Volume has been created!")
			break
		}
	}

	attachOptions := &ec2.AttachVolumeInput{
		Device:     aws.String("/dev/sdg"),
		InstanceId: host.Ec2Instance.InstanceId,
		VolumeId:   resVolume.VolumeId,
		DryRun:     aws.Bool(false),
	}

	log.Infoln("Calling AttachVolume()...")
	log.Infoln("InstanceId:", *host.Ec2Instance.InstanceId, "VolumeId:", *resVolume.VolumeId)
	resAttach, errAttach := client.AttachVolume(attachOptions)
	if errAttach != nil {
		log.Errorln("AttachVolume Failed. Err:", errAttach)
		log.Infoln("createVolume LEAVE")
		return errCreate
	}

	log.Infoln("Checking for AttachVolume status")
	for {
		time.Sleep(time.Duration(common.WaitForVolumeState) * time.Second)

		tmpVol, errVol := getVolume(client, awsInfo, *resAttach.VolumeId)
		if errVol != nil {
			log.Errorln("getVolume Failed. Err:", errVol)
			break
		}

		found := false
		for _, attach := range tmpVol.Attachments {
			log.Infoln("AttachVolume - InstanceId:", *attach.InstanceId, "State:", *attach.State)
			if strings.EqualFold(*attach.InstanceId, *host.Ec2Instance.InstanceId) &&
				*attach.State == ec2.VolumeAttachmentStateAttached {
				log.Infoln("Volume has been attached!")
				found = true
				break
			}
		}

		if found {
			break
		}
	}

	log.Infoln("createVolume Succeeded")
	log.Infoln("createVolume LEAVE")
	return nil
}

func (s *RestServer) findAwsInstanceInScaleIONode(inst *ec2.Instance) *types.ScaleIONode {
	log.Infoln("findAwsInstanceInScaleIONode ENTER")
	for _, node := range s.State.ScaleIO.Nodes {
		log.Infoln(node.Hostname, "=", inst.PrivateDnsName, "or", inst.PublicDnsName, "?")
		if strings.EqualFold(node.Hostname, *inst.PrivateDnsName) ||
			strings.EqualFold(node.Hostname, *inst.PublicDnsName) {
			log.Infoln("Node FOUND!")
			log.Infoln("findAwsInstanceInScaleIONode LEAVE")
			return node
		}
	}

	log.Infoln("Node NOT FOUND")
	log.Infoln("findAwsInstanceInScaleIONode LEAVE")
	return nil
}

func (s *RestServer) checkForFull(state *types.ScaleIOFramework) (*pairDomainPool, error) {
	log.Infoln("checkForFull ENTER")

	client, errClient := s.createScaleioClient(state)
	if errClient != nil {
		log.Errorln("createScaleioClient Failed. Err:", errClient)
		log.Infoln("checkForFull LEAVE")
		return nil, errClient
	}

	system, errSystem := client.FindSystem(s.State.ScaleIO.ClusterID, s.State.ScaleIO.ClusterName, "")
	if errSystem != nil {
		log.Errorln("FindSystem Error:", errSystem)
		log.Infoln("checkForFull LEAVE")
		return nil, errSystem
	}

	tmpDomain, errDomain := system.FindProtectionDomain("", s.Config.ProtectionDomain, "")
	if errDomain != nil {
		log.Errorln("FindProtectionDomain Error:", errDomain)
		log.Infoln("checkForFull LEAVE")
		return nil, errDomain
	}

	log.Infoln("ProtectionDomain exists:", tmpDomain.Name)
	scaleioDomain := goscaleio.NewProtectionDomainEx(client, tmpDomain)

	storagePools, errPools := scaleioDomain.GetStoragePool("")
	if errPools != nil {
		log.Errorln("GetStoragePool Error:", errPools)
		log.Infoln("checkForFull LEAVE")
		return nil, errPools
	}

	poolsNeedExpanding := &pairDomainPool{
		Domain: scaleioDomain,
		Pools:  make([]*goscaleio.StoragePool, 0),
	}

	for _, tmpPool := range storagePools {
		log.Infoln("StoragePool exists:", tmpPool.Name)
		scaleioPool := goscaleio.NewStoragePoolEx(client, tmpPool)

		stats, errStats := scaleioPool.GetStatistics()
		if errStats != nil {
			log.Warnln("GetStatistics Error:", errStats)
			continue
		}

		fakeUsedData := s.State.ScaleIO.FakeUsedData
		s.State.ScaleIO.CapacityData = stats.CapacityLimitInKb
		s.State.ScaleIO.UsedData = stats.CapacityInUseInKb
		log.Infoln("Fake Used Data:", s.State.ScaleIO.FakeUsedData)
		log.Infoln("Capacity:", s.State.ScaleIO.CapacityData)
		log.Infoln("Actual Used Data:", s.State.ScaleIO.UsedData)

		usedSpacePercent := int(float64(stats.CapacityInUseInKb+fakeUsedData) / float64(stats.CapacityLimitInKb) * 100)
		if stats.CapacityLimitInKb == 0 {
			usedSpacePercent = 0
		}
		if usedSpacePercent > s.Config.UsedThreshold {
			log.Infoln("Storage Pool Needs Expanding:", scaleioPool.StoragePool.Name,
				"Percent Used:", usedSpacePercent, "Threshold:", s.Config.UsedThreshold)
			poolsNeedExpanding.Pools = append(poolsNeedExpanding.Pools, scaleioPool)
		} else {
			log.Infoln("Storage Pool:", scaleioPool.StoragePool.Name,
				"Percent Used:", usedSpacePercent, "Threshold:", s.Config.UsedThreshold)
		}
	}

	log.Infoln("checkForFull LEAVE")
	return poolsNeedExpanding, nil
}

func getNextDevicePath(metaData *kvstore.Metadata, node *types.ScaleIONode) string {
	log.Infoln("getNextDevicePath ENTER")

	highestLetter := "f"
	var lastDomain *types.ProtectionDomain
	var lastPool *types.StoragePool
	for _, domain := range node.ProvidesDomains {
		for _, pool := range domain.Pools {
			for _, device := range pool.Devices {
				lenDev := len(device) - 1
				log.Infoln("lenDev:", lenDev)
				if highestLetter[0] < device[lenDev] {
					highestLetter = device[lenDev:]
				}

				lastDomain = domain
				lastPool = pool
			}
		}
	}

	//new device name
	highestLetter = string(highestLetter[0] + 1)
	log.Infoln("highestLetter:", highestLetter)

	newDevPath := "/dev/xvd" + highestLetter
	log.Infoln("newDevPath:", newDevPath)

	//add to metadata
	if metaData.ProtectionDomains == nil {
		log.Debugln("Creating ProtectionDomain Map")
		metaData.ProtectionDomains = make(map[string]*kvstore.ProtectionDomain)
	}
	if metaData.ProtectionDomains[lastDomain.Name] == nil {
		log.Debugln("Creating new ProtectionDomain:", lastDomain.Name)
		metaData.ProtectionDomains[lastDomain.Name] = &kvstore.ProtectionDomain{
			Name: lastDomain.Name,
			Add:  true,
		}
	}
	mDomain := metaData.ProtectionDomains[lastDomain.Name]

	if mDomain.Pools == nil {
		log.Debugln("Creating StoragePool Map")
		mDomain.Pools = make(map[string]*kvstore.StoragePool)
	}
	if mDomain.Pools[lastPool.Name] == nil {
		log.Debugln("Creating new StoragePool:", lastPool.Name)
		mDomain.Pools[lastPool.Name] = &kvstore.StoragePool{
			Name: lastPool.Name,
			Add:  true,
		}
	}
	mPool := mDomain.Pools[lastPool.Name]

	if mPool.Devices == nil {
		log.Debugln("Creating Devices Map")
		mPool.Devices = make(map[string]*kvstore.Device)
	}

	log.Debugln("Creating new Device:", newDevPath)
	metaData.ProtectionDomains[lastDomain.Name].Pools[lastDomain.Name].Devices[newDevPath] = &kvstore.Device{
		Name: newDevPath,
		Add:  true,
	}

	//add to agent object
	lastPool.Devices = append(lastPool.Devices, newDevPath)

	log.Infoln("newDevPath:", newDevPath)
	log.Infoln("getNextDevicePath LEAVE")

	return newDevPath
}

func (s *RestServer) expandPools(pools *pairDomainPool) error {
	log.Infoln("expandPools ENTER")

	//EC2 connect
	client, awsInfo, errConn := awsConnect(s.Config.AccessKey, s.Config.SecretKey)
	if errConn != nil {
		log.Errorln("awsConnect Failed. Ret:", errConn)
		log.Infoln("expandPools LEAVE")
		return errConn
	}

	pairHosts, errHosts := s.getInstanceList(client)
	if errHosts != nil {
		log.Errorln("getInstanceList Failed. Ret:", errHosts)
		log.Infoln("expandPools LEAVE")
		return errHosts
	}

	//Create EBS volumes on all hosts
	for _, pairHost := range pairHosts {
		errCreate := createVolume(client, awsInfo, pairHost, s.Config.VolumeGrowthSize)
		if errCreate != nil {
			log.Errorln("createVolume Error:", errCreate)
		}
	}

	//Attach Device on all hosts
	for _, pool := range pools.Pools {
		for _, pairHost := range pairHosts {
			sdsID := "sds_" + pairHost.ScaleIONode.IPAddress
			tmpSds, errSds := pools.Domain.FindSds("Name", sdsID)
			if errSds != nil {
				log.Errorln("Unable to find SDS:", sdsID)
				continue
			}

			//Get nodes metadata
			metaData, err := s.Store.GetMetadata(pairHost.ScaleIONode.Hostname)
			if err != nil {
				log.Warnln("No metadata for node", pairHost.ScaleIONode.Hostname)
			}

			newDevice := getNextDevicePath(metaData, pairHost.ScaleIONode)

			_, errAttach := pool.AttachDevice(newDevice, tmpSds.ID)
			if errAttach == nil {
				//Save updated metadata
				err = s.Store.SetMetadata(pairHost.ScaleIONode.Hostname, metaData)
				if err == nil {
					log.Debugln("Metadata saved for node:", pairHost.ScaleIONode.Hostname)
				} else {
					log.Errorln("Save metadata failed for node: ", pairHost.ScaleIONode.Hostname, ". Err:", err)
				}
			} else {
				log.Errorln("AttachDevice Error:", errAttach)
			}
		}
	}

	log.Infoln("expandPools LEAVE")
	return nil
}

func setFakeData(w http.ResponseWriter, r *http.Request, server *RestServer) {
	if !server.Config.Debug {
		http.Error(w, "This API is available only in debug mode.", http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		http.Error(w, "Unable to read the HTTP Body stream", http.StatusBadRequest)
		return
	}
	if err := r.Body.Close(); err != nil {
		log.Warnln("Unable to close the HTTP Body stream:", err)
	}

	state := &types.UpdateUsedData{
		Acknowledged: false,
		FakeUsedData: 0,
	}
	if err := json.Unmarshal(body, &state); err != nil {
		http.Error(w, "Unable to marshall the response", http.StatusBadRequest)
		return
	}

	//update the object
	server.Lock()
	server.State.ScaleIO.FakeUsedData = state.FakeUsedData
	server.Unlock()

	//acknowledged the state change
	state.Acknowledged = true

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(state); err != nil {
		http.Error(w, "Unable to marshall the response", http.StatusBadRequest)
	}
}
