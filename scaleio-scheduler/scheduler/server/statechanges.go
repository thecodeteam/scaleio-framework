package server

import (
	log "github.com/Sirupsen/logrus"
	goscaleio "github.com/codedellemc/goscaleio"

	common "github.com/codedellemc/scaleio-framework/scaleio-scheduler/scheduler/common"
	kvstore "github.com/codedellemc/scaleio-framework/scaleio-scheduler/scheduler/kvstore"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

func doesNeedleExist(haystack []string, needle string) bool {
	if haystack == nil {
		return false
	}
	for _, value := range haystack {
		log.Debugln(haystack, "=", needle, "?")
		if value == needle {
			log.Debugln(haystack, "=", needle, "? FOUND!")
			return true
		}
	}
	log.Debugln("doesNeedleExist(", needle, ") NOT FOUND!")
	return false
}

func (s *RestServer) processDeletions(metaData *kvstore.Metadata, node *types.ScaleIONode) bool {
	log.Debugln("processDeletions ENTER")

	if node == nil {
		log.Debugln("node == nil. Invalid node!")
		log.Debugln("processDeletions LEAVE")
		return false
	}
	if metaData == nil {
		log.Debugln("metaData == nil. Means no previous state! First start!")
		log.Debugln("processDeletions LEAVE")
		return false
	}

	bHasChange := false
	log.Debugln("metaData.ProtectionDomains size:", len(metaData.ProtectionDomains))
	for keyD, mDomain := range metaData.ProtectionDomains {
		log.Debugln("Domain:", mDomain.Name)
		nDomain := node.ProvidesDomains[keyD]
		if nDomain == nil {
			log.Debugln("Delete Domain:", mDomain.Name)
			mDomain.Delete = true
			bHasChange = true
			continue
		} else if len(nDomain.Pools) == 0 {
			log.Debugln("Delete Domain (", mDomain.Name, ") because contains no Pools")
			mDomain.Delete = true
			bHasChange = true
		}

		log.Debugln("domain.Sdss size:", len(mDomain.Sdss))
		for keyS, mSds := range mDomain.Sdss {
			log.Debugln("Sds:", keyS)
			if mDomain.Delete {
				log.Debugln("Delete SDS :", mSds.Name)
				mSds.Delete = true
				bHasChange = true
			}
		}

		log.Debugln("domain.Pools size:", len(mDomain.Pools))
		for keyP, mPool := range mDomain.Pools {
			log.Debugln("Pool:", keyP)
			nPool := nDomain.Pools[keyP]
			if nPool == nil {
				log.Debugln("Delete Pool:", mPool.Name)
				mPool.Delete = true
				bHasChange = true
				continue
			} else if mDomain.Delete || len(nPool.Devices) == 0 {
				log.Debugln("Delete Pool (", mPool.Name, ") because contains no Devices")
				mPool.Delete = true
				bHasChange = true
			}

			log.Debugln("pool.Devices size:", len(mPool.Devices))
			for keyDv, mDevice := range mPool.Devices {
				log.Debugln("Device:", keyDv)
				nDevices := nPool.Devices
				if nDevices == nil {
					log.Debugln("nDevices == nil")
					continue
				}
				if mDomain.Delete || mPool.Delete || !doesNeedleExist(nDevices, mDevice.Name) {
					log.Debugln("Delete Device:", mDevice.Name)
					mDevice.Delete = true
					bHasChange = true
				}
			}
		}
	}

	log.Debugln("processDeletions Succeeded. Changes:", bHasChange)
	log.Debugln("processDeletions LEAVE")

	return bHasChange
}

func (s *RestServer) processAdditions(metaData *kvstore.Metadata, node *types.ScaleIONode) bool {
	log.Debugln("processAdditions ENTER")

	if node == nil {
		log.Debugln("node == nil. Invalid node!")
		log.Debugln("processAdditions LEAVE")
		return false
	}
	if metaData == nil {
		log.Debugln("metaData == nil. Means no previous state! First start!")
		log.Debugln("processAdditions LEAVE")
		return false
	}

	bHasChange := false
	//Domain
	for keyD, nDomain := range node.ProvidesDomains {
		if metaData.ProtectionDomains == nil {
			metaData.ProtectionDomains = make(map[string]*kvstore.ProtectionDomain)
		}
		if metaData.ProtectionDomains[keyD] == nil {
			log.Debugln("Creating new ProtectionDomain:", nDomain.Name)
			metaData.ProtectionDomains[keyD] = &kvstore.ProtectionDomain{
				Name: keyD,
				Add:  true,
			}
			bHasChange = true
		} else {
			log.Debugln("ProtectionDomain (", keyD, ") already exists")
		}
		mDomain := metaData.ProtectionDomains[keyD]

		//Sds
		//TODO assume only a single Sds per ProtectionDomain. Will change later.
		//As such, the SDS is implicitly created.
		sdsName := "sds_" + node.IPAddress
		if mDomain.Sdss == nil {
			mDomain.Sdss = make(map[string]*kvstore.Sds)
		}
		if mDomain.Sdss[sdsName] == nil {
			mDomain.Sdss[sdsName] = &kvstore.Sds{
				Name: sdsName,
				Add:  true,
				Mode: kvstore.SdsModeAll,
			}
			bHasChange = true
		} else {
			log.Debugln("SDS (", sdsName, ") already exists")
		}

		//Pool
		for keyP, nPool := range nDomain.Pools {
			if mDomain.Pools[keyP] == nil {
				log.Debugln("Creating new StoragePool (", nPool.Name, ") for domain (", nDomain.Name, ")")
				if mDomain.Pools == nil {
					mDomain.Pools = make(map[string]*kvstore.StoragePool)
				}
				mDomain.Pools[keyP] = &kvstore.StoragePool{
					Name: keyP,
					Add:  true,
				}
				bHasChange = true
			} else {
				log.Debugln("StoragePool (", nPool.Name, ") already exists for domain (", nDomain.Name, ")")
			}
			mPool := mDomain.Pools[keyP]

			//Device
			for _, device := range nPool.Devices {
				if mPool.Devices[device] == nil {
					log.Debugln("Creating new Device (", device, ") for domain (", nDomain.Name, ") and pool (", nPool.Name, ")")
					if mPool.Devices == nil {
						mPool.Devices = make(map[string]*kvstore.Device)
					}
					mPool.Devices[device] = &kvstore.Device{
						Name: device,
						Add:  true,
					}
					bHasChange = true
				} else {
					log.Debugln("Device (", device, ") already exists for domain (", nDomain.Name, ") and pool (", nPool.Name, ")")
				}
			}
		}
	}

	log.Debugln("processAdditions Succeeded. Changes:", bHasChange)
	log.Debugln("processAdditions LEAVE")

	return bHasChange
}

func (s *RestServer) processMetadata(client *goscaleio.Client, node *types.ScaleIONode,
	metaData *kvstore.Metadata) error {
	log.Debugln("processMetadata ENTER")

	system, err := client.FindSystem(s.State.ScaleIO.ClusterID, s.State.ScaleIO.ClusterName, "")
	if err != nil {
		log.Errorln("FindSystem Error:", err)
		log.Debugln("processMetadata LEAVE")
		return err
	}

	//ProtectionDomain
	for _, domain := range metaData.ProtectionDomains {
		tmpDomain, errDomain := system.FindProtectionDomain("", domain.Name, "")
		if errDomain != nil {
			if !domain.Delete && domain.Add {
				_, err := system.CreateProtectionDomain(domain.Name)
				if err == nil {
					log.Infoln("ProtectionDomain created:", domain.Name)
				} else {
					log.Errorln("CreateProtectionDomain Error:", err)
					log.Debugln("processMetadata LEAVE")
					return err
				}
			}
			tmpDomain, errDomain = system.FindProtectionDomain("", domain.Name, "")
			if errDomain == nil {
				log.Infoln("ProtectionDomain found:", domain.Name)
			} else {
				log.Errorln("FindProtectionDomain Error:", errDomain)
				log.Debugln("processMetadata LEAVE")
				return errDomain
			}
		} else {
			log.Infoln("ProtectionDomain exists:", domain.Name)
		}
		scaleioDomain := goscaleio.NewProtectionDomainEx(client, tmpDomain)

		//Sds
		var scaleioSds *goscaleio.Sds

		for _, sds := range domain.Sdss {
			tmpSds, errSds := scaleioDomain.FindSds("Name", sds.Name)
			if errSds != nil {
				if !sds.Delete && sds.Add {
					//TODO fix the IPAddress when ServerOnly and ClientOnly is implemented
					_, err := scaleioDomain.CreateSds(sds.Name, []string{node.IPAddress})
					if err == nil {
						log.Infoln("SDS created:", sds.Name)
					} else {
						log.Errorln("CreateSds Error:", err)
						log.Debugln("processMetadata LEAVE")
						return err
					}
					sds.Add = false
				}
				tmpSds, errSds = scaleioDomain.FindSds("Name", sds.Name)
				if errSds == nil {
					log.Infoln("SDS found:", sds.Name)
				} else {
					log.Errorln("FindSds Error:", errSds)
					log.Debugln("processMetadata LEAVE")
					return errSds
				}
			} else {
				log.Infoln("SDS exists:", sds.Name)
			}

			if sds.Mode == kvstore.SdsModeClient {
				log.Debugln("Skipping SDS", sds.Name, "as it is client only.")
				continue
			}

			log.Debugln("Using SDS", sds.Name, "as it is server or all.")
			scaleioSds = goscaleio.NewSdsEx(client, tmpSds)
		}

		//StoragePool
		for _, pool := range domain.Pools {
			tmpPool, errPool := scaleioDomain.FindStoragePool("", pool.Name, "")
			if errPool != nil {
				if !pool.Delete && pool.Add {
					_, err := scaleioDomain.CreateStoragePool(pool.Name)
					if err == nil {
						log.Infoln("StoragePool created:", pool.Name)
					} else {
						log.Errorln("CreateStoragePool Error:", err)
						log.Debugln("processMetadata LEAVE")
						return err
					}
					pool.Add = false
				}
				tmpPool, errPool = scaleioDomain.FindStoragePool("", pool.Name, "")
				if errPool == nil {
					log.Infoln("StoragePool found:", pool.Name)
				} else {
					log.Errorln("FindStoragePool Error:", errPool)
					log.Debugln("processMetadata LEAVE")
					return errPool
				}
			} else {
				log.Infoln("StoragePool exists:", pool.Name)
			}
			scaleioPool := goscaleio.NewStoragePoolEx(client, tmpPool)

			for _, device := range pool.Devices {
				if device.Delete {
					//TODO API Call DEL Device from Pool
				} else if device.Add {
					_, err := scaleioPool.AttachDevice(device.Name, scaleioSds.Sds.ID)
					if err == nil {
						log.Infoln("Device attached:", device.Name)
					} else {
						log.Errorln("AttachDevice Error:", err)
						log.Debugln("processMetadata LEAVE")
						return err
					}
				}
			}

			if pool.Delete {
				//TODO API Call DEL Pool
			}
		}

		if domain.Delete {
			//TODO API Call DEL SDS
			//TODO API Call DEL Domain
		}
	}

	return nil
}

func (s *RestServer) createScaleioClient(state *types.ScaleIOFramework) (*goscaleio.Client, error) {
	log.Debugln("createScaleioClient ENTER")

	ip, err := common.GetGatewayAddress(state)
	if err != nil {
		log.Errorln("GetGatewayAddress Error:", err)
		log.Debugln("createScaleioClient LEAVE")
		return nil, err
	}

	endpoint := "https://" + ip + "/api"
	log.Infoln("Endpoint:", endpoint)
	log.Infoln("APIVersion:", s.Config.APIVersion)

	client, err := goscaleio.NewClientWithArgs(endpoint, s.Config.APIVersion, true, false)
	if err != nil {
		log.Errorln("NewClientWithArgs Error:", err)
		log.Debugln("createScaleioClient LEAVE")
		return nil, err
	}

	_, err = client.Authenticate(&goscaleio.ConfigConnect{
		Endpoint: endpoint,
		Username: "admin",
		Password: s.Config.AdminPassword,
	})
	if err != nil {
		log.Errorln("Authenticate Error:", err)
		log.Debugln("createScaleioClient LEAVE")
		return nil, err
	}
	log.Infoln("Successfuly logged in to ScaleIO Gateway at", client.SIOEndpoint.String())

	log.Debugln("createScaleioClient Succeeded")
	log.Debugln("createScaleioClient LEAVE")

	return client, nil
}

func (s *RestServer) addResourcesToScaleIO(state *types.ScaleIOFramework) error {
	log.Debugln("addResourcesToScaleIO ENTER")

	var client *goscaleio.Client
	client = nil

	for _, node := range state.ScaleIO.Nodes {
		log.Debugln("Processing node:", node.Hostname)

		if !node.Imperative && !node.Advertised {
			log.Warnln("This node has not advertised its devices yet. Skip!")
			continue
		}

		//Get metadata
		metaData, err := s.Store.GetMetadata(node.Hostname)
		if err != nil {
			log.Warnln("No metadata for node", node.Hostname)
		}

		//if no metadata exists (ie first time running), create new object
		if metaData == nil {
			log.Debugln("Creating new metadata object. No prior state.")
			metaData = new(kvstore.Metadata)
		}

		//look for deletions
		dChanges := s.processDeletions(metaData, node)

		//look for additions
		aChanges := s.processAdditions(metaData, node)

		if !dChanges && !aChanges {
			log.Debugln("There are no new changes for this node.")
			continue
		}

		if client == nil {
			client, err = s.createScaleioClient(state)
			if err != nil {
				log.Errorln("createScaleioClient Failed. Err:", err)
				log.Debugln("addResourcesToScaleIO LEAVE")
				return err
			}
		}

		//process metadata model
		s.processMetadata(client, node, metaData)

		//Save metadata
		err = s.Store.SetMetadata(node.Hostname, metaData)
		if err == nil {
			log.Debugln("Metadata saved for node:", node.Hostname)
		} else {
			log.Errorln("Save metadata failed for node: ", node.Hostname, ". Err:", err)
		}
	}

	log.Debugln("addResourcesToScaleIO Succeeded!")
	log.Debugln("addResourcesToScaleIO LEAVE")

	return nil
}
