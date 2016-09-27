package core

import (
	"time"

	log "github.com/Sirupsen/logrus"

	nodestate "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/node"
	"github.com/codedellemc/scaleio-framework/scaleio-executor/native"
	"github.com/codedellemc/scaleio-framework/scaleio-executor/native/exec"
	"github.com/codedellemc/scaleio-framework/scaleio-executor/native/installers"
	"github.com/codedellemc/scaleio-framework/scaleio-executor/native/installers/deb"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//constants for verifying that the command lines executed properly
const (
	aiozipCheck                = "[0-9]+ upgraded|[0-9]+ newly"
	genericInstallCheck        = "1 upgraded|1 newly"
	requiredKernelVersionCheck = "4.2.0-30-generic"
	rebootCheck                = "reboot in 1 minute"
	mdmInstallCheck            = "mdm start/running"
	sdsInstallCheck            = "sds start/running"
	sdcInstallCheck            = "Success configuring module"
	clusterConfigCheck         = "Mode: 3_node"
	createClusterCheck         = "Successfully created the MDM Cluster"
	loggedInCheck              = "Logged in"
	setPasswordCheck           = "Password changed successfully"
	addMdmToClusterCheck       = "Successfully added a standby MDM"
	changeClusterModeCheck     = "Successfully switched the cluster mode"
	clusterNotInitialedCheck   = "Query-all-SDS returned 0 SDS nodes"
	liaInstallCheck            = "lia start/running"
	liaRestartCheck            = liaInstallCheck
	gatewayInstallCheck        = "The EMC ScaleIO Gateway is running"
	gatewayRestartCheck        = "scaleio-gateway start/running"
	clusterRenameCheck         = "Successfully renamed system to"
	addProtectionDomainCheck   = "Successfully created protection domain"
	addStoragePoolCheck        = "Successfully created a storage pool"
	addSdsCheck                = "Successfully created SDS"
	addVolumeCheck             = "Successfully created volume of size"
)

func environmentSetup(state *types.ScaleIOFramework) (bool, error) {
	log.Infoln("EnvironmentSetup ENTER")

	aioErr := deb.IsInstalled("libaio1")
	zipErr := deb.IsInstalled("zip")
	if aioErr != nil || zipErr != nil {
		log.Infoln("Installing libaio1 and zip")

		miscCmdline := "apt-get -y install libaio1 zip"
		err := exec.RunCommand(miscCmdline, aiozipCheck, "")
		if err != nil {
			log.Errorln("Install Prerequisites Failed:", err)
			log.Infoln("EnvironmentSetup LEAVE")
			return false, err
		}
	} else {
		log.Infoln("libaio1 and zip are already installed")
	}

	kernelErr := deb.IsInstalled("linux-image-4.2.0-30-generic")
	if kernelErr != nil {
		log.Infoln("Installing linux-image-4.2.0-30-generic")

		kernelCmdline := "apt-get -y install linux-image-4.2.0-30-generic"
		err := exec.RunCommand(kernelCmdline, genericInstallCheck, "")
		if err != nil {
			log.Errorln("Install Kernel Failed:", err)
			log.Infoln("EnvironmentSetup LEAVE")
			return false, err
		}
	} else {
		log.Infoln("linux-image-4.2.0-30-generic is already installed")
	}

	kernelVer, kernelVerErr := native.GetRunningKernelVersion()
	if kernelVerErr != nil {
		log.Errorln("Kernel Version Check Failed:", kernelVerErr)
		log.Infoln("EnvironmentSetup LEAVE")
		return false, kernelVerErr
	}

	if kernelVer != requiredKernelVersionCheck {
		log.Errorln("Reboot Required!")
		log.Infoln("EnvironmentSetup LEAVE")
		return true, nil
	}

	log.Infoln("Already running kernel version", requiredKernelVersionCheck)

	log.Infoln("EnvironmentSetup Succeeded")
	log.Infoln("EnvironmentSetup LEAVE")
	return false, nil
}

func managementSetup(state *types.ScaleIOFramework, isPriOrSec bool) error {
	log.Infoln("ManagementSetup ENTER")

	mdmPair, errBase := createMdmPairString(state)
	if errBase != nil {
		log.Errorln("Error downloading MDM package:", errBase)
		log.Infoln("ManagementSetup LEAVE")
		return errBase
	}
	log.Infoln("MDM Pair String:", mdmPair)

	//MDM Install
	mdmVer, mdmVerErr := installers.ParseVersionFromFilename(state.ScaleIO.Deb.DebMdm)
	mdmInst, mdmInstErr := deb.GetInstalledVersion(types.DebMdmPackageName, true)
	log.Debugln("mdmVer:", mdmVer)
	log.Debugln("mdmVerErr:", mdmVerErr)
	log.Debugln("mdmInst:", mdmInst)
	log.Debugln("mdmInstErr:", mdmInstErr)

	if mdmVerErr != nil || mdmInstErr != nil || mdmVer != mdmInst {
		log.Infoln("Installing", types.DebMdmPackageName)

		localMdm, err := installers.DownloadPackage(state.ScaleIO.Deb.DebMdm)
		if err != nil {
			log.Errorln("Error downloading MDM package:", err)
			log.Infoln("ManagementSetup LEAVE")
			return err
		}

		var strPriOrSec string
		if isPriOrSec {
			strPriOrSec = "1"
		} else {
			strPriOrSec = "0"
		}

		mdmCmdline := "MDM_ROLE_IS_MANAGER=" + strPriOrSec + " dpkg -i " + localMdm
		err = exec.RunCommand(mdmCmdline, mdmInstallCheck, "")
		if err != nil {
			log.Errorln("Install MDM Failed:", err)
			log.Infoln("ManagementSetup LEAVE")
			return err
		}
	} else {
		log.Infoln(types.DebMdmPackageName, "is already installed")
	}

	log.Infoln("ManagementSetup Succeeded")
	log.Infoln("ManagementSetup LEAVE")
	return nil
}

func nodeSetup(state *types.ScaleIOFramework) error {
	log.Infoln("NodeSetup ENTER")

	mdmPair, errBase := createMdmPairString(state)
	if errBase != nil {
		log.Errorln("Error downloading MDM package:", errBase)
		log.Infoln("NodeSetup LEAVE")
		return errBase
	}
	log.Infoln("MDM Pair String:", mdmPair)

	//SDS Install
	sdsVer, sdsVerErr := installers.ParseVersionFromFilename(state.ScaleIO.Deb.DebSds)
	sdsInst, sdsInstErr := deb.GetInstalledVersion(types.DebSdsPackageName, true)
	log.Debugln("sdsVer:", sdsVer)
	log.Debugln("sdsVerErr:", sdsVerErr)
	log.Debugln("sdsInst:", sdsInst)
	log.Debugln("sdsInstErr:", sdsInstErr)

	if sdsVerErr != nil || sdsInstErr != nil || sdsVer != sdsInst {
		log.Infoln("Installing", types.DebSdsPackageName)

		localSds, err := installers.DownloadPackage(state.ScaleIO.Deb.DebSds)
		if err != nil {
			log.Errorln("Error downloading SDS package:", err)
			log.Infoln("NodeSetup LEAVE")
			return err
		}

		sdsCmdline := "dpkg -i " + localSds
		err = exec.RunCommand(sdsCmdline, sdsInstallCheck, "")
		if err != nil {
			log.Errorln("Install SDS Failed:", err)
			log.Infoln("NodeSetup LEAVE")
			return err
		}
	} else {
		log.Infoln(types.DebSdsPackageName, "is already installed")
	}

	//SDC Install
	sdcVer, sdcVerErr := installers.ParseVersionFromFilename(state.ScaleIO.Deb.DebSdc)
	sdcInst, sdcInstErr := deb.GetInstalledVersion(types.DebSdcPackageName, true)
	log.Debugln("sdcVer:", sdcVer)
	log.Debugln("sdcVerErr:", sdcVerErr)
	log.Debugln("sdcInst:", sdcInst)
	log.Debugln("sdcInstErr:", sdcInstErr)

	if sdcVerErr != nil || sdcInstErr != nil || sdcVer != sdcInst {
		log.Infoln("Installing", types.DebSdcPackageName)

		localSdc, err := installers.DownloadPackage(state.ScaleIO.Deb.DebSdc)
		if err != nil {
			log.Errorln("Error downloading SDC package:", err)
			log.Infoln("NodeSetup LEAVE")
			return err
		}

		sdcCmdline := "MDM_IP=" + mdmPair + " dpkg -i " + localSdc
		err = exec.RunCommand(sdcCmdline, sdcInstallCheck, "")
		if err != nil {
			log.Errorln("Install SDC Failed:", err)
			log.Infoln("NodeSetup LEAVE")
			return err
		}
	} else {
		log.Infoln(types.DebSdcPackageName, "is already installed")
	}

	log.Infoln("NodeSetup Succeeded")
	log.Infoln("NodeSetup LEAVE")
	return nil
}

func isClusterInstalled() error {
	log.Infoln("isClusterInstalled ENTER")

	queryCmdline := "scli --query_cluster"
	err := exec.RunCommand(queryCmdline, clusterConfigCheck, "")
	if err != nil {
		log.Errorln("Query Cluster Failed:", err)
		log.Infoln("isClusterInstalled LEAVE")
		return err
	}

	log.Debugln("isClusterInstalled Succeeded")
	log.Infoln("isClusterInstalled LEAVE")
	return nil
}

func createCluster(state *types.ScaleIOFramework) error {
	log.Infoln("CreateCluster ENTER")

	errCheck := isClusterInstalled()
	if errCheck == nil {
		log.Infoln("ScaleIO cluster is already installed")
		log.Infoln("CreateCluster LEAVE")
		return nil
	}

	//Needed to setup cluster
	pri, err := getPrimaryMdmNode(state)
	if err != nil {
		log.Errorln("Cannot find Primary MDM node")
		log.Infoln("CreateCluster LEAVE")
		return err
	}
	sec, err := getSecondaryMdmNode(state)
	if err != nil {
		log.Errorln("Cannot find Secondary MDM node")
		log.Infoln("CreateCluster LEAVE")
		return err
	}
	tb, err := getTiebreakerMdmNode(state)
	if err != nil {
		log.Errorln("Cannot find TieBreaker MDM node")
		log.Infoln("CreateCluster LEAVE")
		return err
	}

	createCmdline := "scli --create_mdm_cluster --master_mdm_ip " + pri.IPAddress +
		" --master_mdm_management_ip " + pri.IPAddress + " --master_mdm_name mdm1 --accept_license " +
		"--approve_certificate"
	err = exec.RunCommand(createCmdline, createClusterCheck, "")
	if err != nil {
		log.Errorln("Init First Node Failed:", err)
		log.Infoln("CreateCluster LEAVE")
		return err
	}

	time.Sleep(time.Duration(DelayBetweenCommandsInSeconds) * time.Second)

	loginCmdline := "scli --login --username admin --password admin"
	err = exec.RunCommand(loginCmdline, loggedInCheck, "")
	if err != nil {
		log.Errorln("ScaleIO Login Failed:", err)
		log.Infoln("CreateCluster LEAVE")
		return err
	}

	time.Sleep(time.Duration(DelayBetweenCommandsInSeconds) * time.Second)

	setPassCmdline := "scli --set_password --old_password admin --new_password " +
		state.ScaleIO.AdminPassword
	err = exec.RunCommand(setPassCmdline, setPasswordCheck, "")
	if err != nil {
		log.Errorln("ScaleIO Set Password Failed:", err)
		log.Infoln("CreateCluster LEAVE")
		return err
	}

	time.Sleep(time.Duration(DelayBetweenCommandsInSeconds) * time.Second)

	loginCmdline = "scli --login --username admin --password " + state.ScaleIO.AdminPassword
	err = exec.RunCommand(loginCmdline, loggedInCheck, "")
	if err != nil {
		log.Errorln("ScaleIO Login with new Password Failed:", err)
		log.Infoln("CreateCluster LEAVE")
		return err
	}

	time.Sleep(time.Duration(DelayBetweenCommandsInSeconds) * time.Second)

	secondaryCmdline := "scli --add_standby_mdm --new_mdm_ip " + sec.IPAddress +
		" --mdm_role manager --new_mdm_management_ip " + sec.IPAddress + " --new_mdm_name mdm2"
	err = exec.RunCommand(secondaryCmdline, addMdmToClusterCheck, "")
	if err != nil {
		log.Errorln("Add Secondary MDM Failed:", err)
		log.Infoln("CreateCluster LEAVE")
		return err
	}

	time.Sleep(time.Duration(DelayBetweenCommandsInSeconds) * time.Second)

	tiebreakerCmdline := "scli --add_standby_mdm --new_mdm_ip " + tb.IPAddress +
		" --mdm_role tb --new_mdm_name tb"
	err = exec.RunCommand(tiebreakerCmdline, addMdmToClusterCheck, "")
	if err != nil {
		log.Errorln("Add Tiebreaker MDM Failed:", err)
		log.Infoln("CreateCluster LEAVE")
		return err
	}

	time.Sleep(time.Duration(DelayBetweenCommandsInSeconds) * time.Second)

	changeClusterCmdline := "scli --switch_cluster_mode --cluster_mode 3_node " +
		"--add_slave_mdm_name mdm2 --add_tb_name tb"
	err = exec.RunCommand(changeClusterCmdline, changeClusterModeCheck, "")
	if err != nil {
		log.Errorln("Change ScaleIO to 3 Node Cluster Failed:", err)
		log.Infoln("CreateCluster LEAVE")
		return err
	}

	time.Sleep(time.Duration(DelayBetweenCommandsInSeconds) * time.Second)

	queryCmdline := "scli --query_cluster"
	err = exec.RunCommand(queryCmdline, clusterConfigCheck, "")
	if err != nil {
		log.Errorln("Query Cluster Failed:", err)
		log.Infoln("CreateCluster LEAVE")
		return err
	}

	time.Sleep(time.Duration(DelayBetweenCommandsInSeconds) * time.Second)

	log.Infoln("CreateCluster Succeeded")
	log.Infoln("CreateCluster LEAVE")
	return nil
}

func isClusterInitialized() error {
	log.Infoln("isClusterInitialized ENTER")

	queryCmdline := "scli --query_all_sds"
	err := exec.RunCommand(queryCmdline, "", clusterNotInitialedCheck)
	if err != nil {
		log.Errorln("Check Cluster Failed:", err)
		log.Infoln("isClusterInitialized LEAVE")
		return err
	}

	time.Sleep(time.Duration(DelayBetweenCommandsInSeconds) * time.Second)

	log.Errorln("Query Cluster Succeeded")
	log.Infoln("isClusterInitialized LEAVE")
	return nil
}

func initializeCluster(state *types.ScaleIOFramework) error {
	log.Infoln("InitializeCluster ENTER")

	errCheck := isClusterInitialized()
	if errCheck == nil {
		log.Infoln("ScaleIO cluster is already initialized")
		log.Infoln("CreateCluster LEAVE")
		return nil
	}

	//Needed to setup cluster
	pri, err := getPrimaryMdmNode(state)
	if err != nil {
		log.Errorln("Unable to find the Primary MDM node")
		log.Infoln("CreateCluster LEAVE")
		return err
	}

	renameCmdline := "scli --mdm_ip " + pri.IPAddress + " --rename_system --new_name scaleio"
	err = exec.RunCommand(renameCmdline, clusterRenameCheck, "")
	if err != nil {
		log.Errorln("Cluster Rename Failed:", err)
		log.Infoln("InitializeCluster LEAVE")
		return err
	}

	time.Sleep(time.Duration(DelayBetweenCommandsInSeconds) * time.Second)

	addProtectionDomainCmdline := "scli --add_protection_domain --protection_domain_name " +
		state.ScaleIO.ProtectionDomain
	err = exec.RunCommand(addProtectionDomainCmdline, addProtectionDomainCheck, "")
	if err != nil {
		log.Errorln("Add Protection Domain Failed:", err)
		log.Infoln("InitializeCluster LEAVE")
		return err
	}

	time.Sleep(time.Duration(DelayBetweenCommandsInSeconds) * time.Second)

	addStoragePoolCmdline := "scli --add_storage_pool --protection_domain_name " +
		state.ScaleIO.ProtectionDomain + " --storage_pool_name " + state.ScaleIO.StoragePool
	err = exec.RunCommand(addStoragePoolCmdline, addStoragePoolCheck, "")
	if err != nil {
		log.Errorln("Add Storage Pool Failed:", err)
		log.Infoln("InitializeCluster LEAVE")
		return err
	}

	time.Sleep(time.Duration(DelayBetweenCommandsInSeconds) * time.Second)

	err = addSdsNodesToCluster(state, false)
	if err != nil {
		log.Errorln("Failed to add node to ScaleIO cluster:", err)
		log.Infoln("InitializeCluster LEAVE")
		return err
	}

	if state.DemoMode {
		time.Sleep(time.Duration(DelayOnVolumeCreateInSeconds) * time.Second)

		addVolumeCmdline := "scli --mdm_ip " + pri.IPAddress + " --add_volume --size_gb 1 " +
			"--volume_name test --protection_domain_name " + state.ScaleIO.ProtectionDomain +
			" --storage_pool_name " + state.ScaleIO.StoragePool
		err = exec.RunCommand(addVolumeCmdline, addVolumeCheck, "")
		if err != nil {
			log.Errorln("Add Test Volume Failed:", err)
			log.Infoln("InitializeCluster LEAVE")
			return err
		}
	}

	time.Sleep(time.Duration(DelayBetweenCommandsInSeconds) * time.Second)

	log.Infoln("InitializeCluster Succeeded")
	log.Infoln("InitializeCluster LEAVE")

	return nil
}

func addSdsNodesToCluster(state *types.ScaleIOFramework, needsLogin bool) error {
	log.Infoln("AddSdsNodesToCluster ENTER")
	log.Infoln("needsLogin:", needsLogin)

	loggedIn := false

	for _, node := range state.ScaleIO.Nodes {
		if node.InCluster {
			log.Infoln("Node", node.ExecutorID, "has already been added to the cluster")
			continue
		}
		if node.State < types.StateBasePackagedInstalled {
			log.Infoln("Node", node.ExecutorID, "is not in the cluster but has also not "+
				"finished the installation of the ScaleIO packages yet.")
			continue
		}

		if needsLogin && !loggedIn {
			loggedIn = true

			loginCmdline := "scli --login --username admin --password " + state.ScaleIO.AdminPassword
			err := exec.RunCommand(loginCmdline, loggedInCheck, "")
			if err != nil {
				log.Errorln("ScaleIO Login with new Password Failed:", err)
				log.Infoln("CreateCluster LEAVE")
				return err
			}
		}

		time.Sleep(time.Duration(DelayBetweenCommandsInSeconds) * time.Second)

		log.Infoln("Adding Node", node.ExecutorID, "/", node.IPAddress,
			"to the ScaleIO cluster.")

		addSdsCmdline := "scli --add_sds --sds_ip " + node.IPAddress +
			" --device_path " + state.ScaleIO.BlockDevice + " --storage_pool_name " +
			state.ScaleIO.StoragePool + " --protection_domain_name " +
			state.ScaleIO.ProtectionDomain + " --sds_name " + generateSdsName(node)
		err := exec.RunCommand(addSdsCmdline, addSdsCheck, "")
		if err != nil {
			log.Errorln("Add SDS node Failed:", err)
			log.Infoln("AddSdsNodesToCluster LEAVE")
			return err
		}

		errState := nodestate.UpdateAddNode(state.SchedulerAddress, node.ExecutorID)
		if errState != nil {
			log.Errorln("Failed to signal add node change:", errState)
		}
	}

	log.Infoln("AddSdsNodesToCluster Succeeded")
	log.Infoln("AddSdsNodesToCluster LEAVE")

	return nil
}

func gatewaySetup(state *types.ScaleIOFramework) (bool, error) {
	log.Infoln("GatewaySetup ENTER")

	pri, errPri := getPrimaryMdmNode(state)
	if errPri != nil {
		log.Errorln("getPrimaryMdmNode Failed:", errPri)
		log.Infoln("GatewaySetup LEAVE")
		return false, errPri
	}
	sec, errSec := getSecondaryMdmNode(state)
	if errSec != nil {
		log.Errorln("getSecondaryMdmNode Failed:", errSec)
		log.Infoln("GatewaySetup LEAVE")
		return false, errSec
	}

	//Install LIA
	liaVer, liaVerErr := installers.ParseVersionFromFilename(state.ScaleIO.Deb.DebLia)
	liaInst, liaInstErr := deb.GetInstalledVersion(types.DebLiaPackageName, true)
	log.Debugln("liaVer:", liaVer)
	log.Debugln("liaVerErr:", liaVerErr)
	log.Debugln("liaInst:", liaInst)
	log.Debugln("liaInstErr:", liaInstErr)

	if liaVerErr != nil || liaInstErr != nil || liaVer != liaInst {
		log.Infoln("Installing", types.DebLiaPackageName)

		localLia, err := installers.DownloadPackage(state.ScaleIO.Deb.DebLia)
		if err != nil {
			log.Errorln("Error downloading LIA package:", err)
			log.Infoln("PrimaryMDM LEAVE")
			return false, err
		}

		liaCmdline := "TOKEN=" + state.ScaleIO.AdminPassword + " dpkg -i " + localLia
		err = exec.RunCommand(liaCmdline, liaInstallCheck, "")
		if err != nil {
			log.Errorln("Install LIA Failed:", err)
			log.Infoln("GatewaySetup LEAVE")
			return false, err
		}

		installIDCmdline := "scli --query_all | grep \"Installation ID\" | sed -n -e 's/^.*ID: //p'"
		output, err := exec.RunCommandOutput(installIDCmdline)
		if err != nil {
			log.Errorln("Install LIA Failed:", err)
			log.Infoln("GatewaySetup LEAVE")
			return false, err
		}

		dumpIDCmdline := "echo " + output + " > /opt/emc/scaleio/lia/cfg/installation_id.txt"
		output, err = exec.RunCommandOutput(dumpIDCmdline)
		if err != nil || len(output) > 0 {
			log.Errorln("Install LIA Failed:", err)
			log.Infoln("GatewaySetup LEAVE")
			return false, err
		}

		serviceliaCmdline := "service lia restart"
		err = exec.RunCommand(serviceliaCmdline, liaRestartCheck, "")
		if err != nil {
			log.Errorln("Restart LIA Failed:", err)
			log.Infoln("GatewaySetup LEAVE")
			return false, err
		}
	} else {
		log.Infoln(types.DebLiaPackageName, "is already installed")
	}

	//Install Gateway
	gwVer, gwVerErr := installers.ParseVersionFromFilename(state.ScaleIO.Deb.DebGw)
	gwInst, gwInstErr := deb.GetInstalledVersion(types.DebGwPackageName, true)
	log.Debugln("gwVer:", gwVer)
	log.Debugln("gwVerErr:", gwVerErr)
	log.Debugln("gwInst:", gwInst)
	log.Debugln("gwInstErr:", gwInstErr)

	if gwVerErr != nil || gwInstErr != nil || gwVer != gwInst {
		log.Infoln("Installing", types.DebGwPackageName)

		localGw, err := installers.DownloadPackage(state.ScaleIO.Deb.DebGw)
		if err != nil {
			log.Errorln("Error downloading Gateway package:", err)
			log.Infoln("PrimaryMDM LEAVE")
			return false, err
		}

		gwCmdline := "GATEWAY_ADMIN_PASSWORD=" + state.ScaleIO.AdminPassword + " dpkg -i " + localGw
		err = exec.RunCommand(gwCmdline, gatewayInstallCheck, "")
		if err != nil {
			log.Errorln("Install GW Failed:", err)
			log.Infoln("GatewaySetup LEAVE")
			return false, err
		}

		bypasssecCmdline := "sed -i 's/security.bypass_certificate_check=false/security.bypass_certificate_check=true/' /opt/emc/scaleio/gateway/webapps/ROOT/WEB-INF/classes/gatewayUser.properties"
		output, err := exec.RunCommandOutput(bypasssecCmdline)
		if err != nil || len(output) > 0 {
			log.Errorln("Configure By-Pass Security Check Failed:", err)
			log.Infoln("GatewaySetup LEAVE")
			return false, err
		}

		writemdmCmdline := "sed -i 's/mdm.ip.addresses=/mdm.ip.addresses='" + pri.IPAddress +
			"','" + sec.IPAddress + "'/' /opt/emc/scaleio/gateway/webapps/ROOT/WEB-INF/classes/gatewayUser.properties"
		output, err = exec.RunCommandOutput(writemdmCmdline)
		if err != nil || len(output) > 0 {
			log.Errorln("Configure MDM to Gateway Failed:", err)
			log.Infoln("GatewaySetup LEAVE")
			return false, err
		}
	} else {
		log.Infoln(types.DebGwPackageName, "is already installed")
	}

	if gwInst == "" && gwInstErr == nil {
		log.Debugln("No previous install of", types.DebGwPackageName,
			"exists. Reboot required!")
		log.Infoln("GatewaySetup LEAVE")
		return true, nil
	}

	log.Infoln("GatewaySetup Succeeded")
	log.Infoln("GatewaySetup LEAVE")
	return false, nil
}
