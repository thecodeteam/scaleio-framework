package mgr

import (
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	xplatform "github.com/dvonthenen/goxplatform"

	common "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/common"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//constants for verifying that the command lines executed properly
const (
	clusterConfigCheck       = "Mode: 3_node"
	createClusterCheck       = "Successfully created the MDM Cluster"
	loggedInCheck            = "Logged in"
	setPasswordCheck         = "Password changed successfully"
	addMdmToClusterCheck     = "Successfully added a standby MDM"
	changeClusterModeCheck   = "Successfully switched the cluster mode"
	clusterNotInitialedCheck = "Query-all-SDS returned 0 SDS nodes"
	clusterRenameCheck       = "Successfully renamed system to"
	addProtectionDomainCheck = "Successfully created protection domain"
	addStoragePoolCheck      = "Successfully created a storage pool"
	addSdsCheck              = "Successfully created SDS"
	addVolumeCheck           = "Successfully created volume of size"
)

//ManagementSetup for setting up the MDM packages
func (mm *MdmManager) ManagementSetup(state *types.ScaleIOFramework, isPriOrSec bool) error {
	log.Infoln("ManagementSetup ENTER")

	mdmPair, errBase := common.CreateMdmPairString(state)
	if errBase != nil {
		log.Errorln("Error downloading MDM package:", errBase)
		log.Infoln("ManagementSetup LEAVE")
		return errBase
	}
	log.Infoln("MDM Pair String:", mdmPair)

	//MDM Install
	mdmVer, mdmVerErr := xplatform.GetInstance().Inst.ParseVersionFromFilename(mm.MdmPackageDownload)
	mdmInst, mdmInstErr := xplatform.GetInstance().Inst.GetInstalledVersion(mm.MdmPackageName, true)
	log.Debugln("mdmVer:", mdmVer)
	log.Debugln("mdmVerErr:", mdmVerErr)
	log.Debugln("mdmInst:", mdmInst)
	log.Debugln("mdmInstErr:", mdmInstErr)

	if mdmVerErr != nil || mdmInstErr != nil || mdmVer != mdmInst {
		log.Infoln("Installing", mm.MdmPackageName)

		localMdm, err := xplatform.GetInstance().Inst.DownloadPackage(mm.MdmPackageDownload)
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

		mdmInstallCmd := strings.Replace(mm.MdmInstallCmd, "{PriOrSec}", strPriOrSec, -1)
		mdmInstallCmd = strings.Replace(mdmInstallCmd, "{LocalMdm}", localMdm, -1)
		log.Infoln("mdmInstallCmd:", mdmInstallCmd)

		err = xplatform.GetInstance().Run.Command(mdmInstallCmd, mm.MdmInstallCheck, "")
		if err != nil {
			log.Errorln("Install MDM Failed:", err)
			log.Infoln("ManagementSetup LEAVE")
			return err
		}
	} else {
		log.Infoln(mm.MdmPackageName, "is already installed")
		time.Sleep(time.Duration(common.DelayIfInstalledInSeconds) * time.Second)
	}

	log.Infoln("ManagementSetup Succeeded")
	log.Infoln("ManagementSetup LEAVE")
	return nil
}

//NodeSetup for setting up the SDS and SDC packages
func (nm *NodeManager) NodeSetup(state *types.ScaleIOFramework) error {
	log.Infoln("NodeSetup ENTER")

	mdmPair, errBase := common.CreateMdmPairString(state)
	if errBase != nil {
		log.Errorln("Error downloading MDM package:", errBase)
		log.Infoln("NodeSetup LEAVE")
		return errBase
	}
	log.Infoln("MDM Pair String:", mdmPair)

	//SDS Install
	sdsVer, sdsVerErr := xplatform.GetInstance().Inst.ParseVersionFromFilename(nm.SdsPackageDownload)
	sdsInst, sdsInstErr := xplatform.GetInstance().Inst.GetInstalledVersion(nm.SdsPackageName, true)
	log.Debugln("sdsVer:", sdsVer)
	log.Debugln("sdsVerErr:", sdsVerErr)
	log.Debugln("sdsInst:", sdsInst)
	log.Debugln("sdsInstErr:", sdsInstErr)

	if sdsVerErr != nil || sdsInstErr != nil || sdsVer != sdsInst {
		log.Infoln("Installing", nm.SdsPackageName)

		localSds, err := xplatform.GetInstance().Inst.DownloadPackage(nm.SdsPackageDownload)
		if err != nil {
			log.Errorln("Error downloading SDS package:", err)
			log.Infoln("NodeSetup LEAVE")
			return err
		}

		sdsInstallCmd := strings.Replace(nm.SdsInstallCmd, "{LocalSds}", localSds, -1)
		log.Infoln("sdsInstallCmd:", sdsInstallCmd)

		err = xplatform.GetInstance().Run.Command(sdsInstallCmd, nm.SdsInstallCheck, "")
		if err != nil {
			log.Errorln("Install SDS Failed:", err)
			log.Infoln("NodeSetup LEAVE")
			return err
		}
	} else {
		log.Infoln(nm.SdsPackageName, "is already installed")
		time.Sleep(time.Duration(common.DelayIfInstalledInSeconds) * time.Second)
	}

	//SDC Install
	sdcVer, sdcVerErr := xplatform.GetInstance().Inst.ParseVersionFromFilename(nm.SdcPackageDownload)
	sdcInst, sdcInstErr := xplatform.GetInstance().Inst.GetInstalledVersion(nm.SdcPackageName, true)
	log.Debugln("sdcVer:", sdcVer)
	log.Debugln("sdcVerErr:", sdcVerErr)
	log.Debugln("sdcInst:", sdcInst)
	log.Debugln("sdcInstErr:", sdcInstErr)

	if sdcVerErr != nil || sdcInstErr != nil || sdcVer != sdcInst {
		log.Infoln("Installing", nm.SdcPackageName)

		localSdc, err := xplatform.GetInstance().Inst.DownloadPackage(nm.SdcPackageDownload)
		if err != nil {
			log.Errorln("Error downloading SDC package:", err)
			log.Infoln("NodeSetup LEAVE")
			return err
		}

		sdcInstallCmd := strings.Replace(nm.SdcInstallCmd, "{MdmPair}", mdmPair, -1)
		sdcInstallCmd = strings.Replace(sdcInstallCmd, "{LocalSdc}", localSdc, -1)
		log.Infoln("sdcInstallCmd:", sdcInstallCmd)

		err = xplatform.GetInstance().Run.Command(sdcInstallCmd, nm.SdcInstallCheck, "")
		if err != nil {
			log.Errorln("Install SDC Failed:", err)
			log.Infoln("NodeSetup LEAVE")
			return err
		}
	} else {
		log.Infoln(nm.SdcPackageName, "is already installed")
		time.Sleep(time.Duration(common.DelayIfInstalledInSeconds) * time.Second)
	}

	log.Infoln("NodeSetup Succeeded")
	log.Infoln("NodeSetup LEAVE")
	return nil
}

//CreateCluster creates the ScaleIO cluster
func (mm *MdmManager) CreateCluster(state *types.ScaleIOFramework) error {
	log.Infoln("CreateCluster ENTER")

	if state.ScaleIO.Configured {
		log.Infoln("ScaleIO cluster is already installed")
		time.Sleep(time.Duration(common.DelayIfInstalledInSeconds) * time.Second * 2)
		log.Infoln("CreateCluster LEAVE")
		return nil
	}

	//Needed to setup cluster
	pri, err := common.GetPrimaryMdmNode(state)
	if err != nil {
		log.Errorln("Cannot find Primary MDM node")
		log.Infoln("CreateCluster LEAVE")
		return err
	}
	sec, err := common.GetSecondaryMdmNode(state)
	if err != nil {
		log.Errorln("Cannot find Secondary MDM node")
		log.Infoln("CreateCluster LEAVE")
		return err
	}
	tb, err := common.GetTiebreakerMdmNode(state)
	if err != nil {
		log.Errorln("Cannot find TieBreaker MDM node")
		log.Infoln("CreateCluster LEAVE")
		return err
	}

	createCmdline := "scli --create_mdm_cluster --master_mdm_ip " + pri.IPAddress +
		" --master_mdm_management_ip " + pri.IPAddress + " --master_mdm_name mdm1 --accept_license " +
		"--approve_certificate"
	err = xplatform.GetInstance().Run.Command(createCmdline, createClusterCheck, "")
	if err != nil {
		log.Errorln("Init First Node Failed:", err)
		log.Infoln("CreateCluster LEAVE")
		return err
	}

	time.Sleep(time.Duration(common.DelayBetweenCommandsInSeconds) * time.Second)

	loginCmdline := "scli --login --username admin --password admin"
	err = xplatform.GetInstance().Run.Command(loginCmdline, loggedInCheck, "")
	if err != nil {
		log.Errorln("ScaleIO Login Failed:", err)
		log.Infoln("CreateCluster LEAVE")
		return err
	}

	time.Sleep(time.Duration(common.DelayBetweenCommandsInSeconds) * time.Second)

	setPassCmdline := "scli --set_password --old_password admin --new_password " +
		state.ScaleIO.AdminPassword
	err = xplatform.GetInstance().Run.Command(setPassCmdline, setPasswordCheck, "")
	if err != nil {
		log.Errorln("ScaleIO Set Password Failed:", err)
		log.Infoln("CreateCluster LEAVE")
		return err
	}

	time.Sleep(time.Duration(common.DelayBetweenCommandsInSeconds) * time.Second)

	loginCmdline = "scli --login --username admin --password " + state.ScaleIO.AdminPassword
	err = xplatform.GetInstance().Run.Command(loginCmdline, loggedInCheck, "")
	if err != nil {
		log.Errorln("ScaleIO Login with new Password Failed:", err)
		log.Infoln("CreateCluster LEAVE")
		return err
	}

	time.Sleep(time.Duration(common.DelayBetweenCommandsInSeconds) * time.Second)

	secondaryCmdline := "scli --add_standby_mdm --new_mdm_ip " + sec.IPAddress +
		" --mdm_role manager --new_mdm_management_ip " + sec.IPAddress + " --new_mdm_name mdm2"
	err = xplatform.GetInstance().Run.Command(secondaryCmdline, addMdmToClusterCheck, "")
	if err != nil {
		log.Errorln("Add Secondary MDM Failed:", err)
		log.Infoln("CreateCluster LEAVE")
		return err
	}

	time.Sleep(time.Duration(common.DelayBetweenCommandsInSeconds) * time.Second)

	tiebreakerCmdline := "scli --add_standby_mdm --new_mdm_ip " + tb.IPAddress +
		" --mdm_role tb --new_mdm_name tb"
	err = xplatform.GetInstance().Run.Command(tiebreakerCmdline, addMdmToClusterCheck, "")
	if err != nil {
		log.Errorln("Add Tiebreaker MDM Failed:", err)
		log.Infoln("CreateCluster LEAVE")
		return err
	}

	time.Sleep(time.Duration(common.DelayBetweenCommandsInSeconds) * time.Second)

	changeClusterCmdline := "scli --switch_cluster_mode --cluster_mode 3_node " +
		"--add_slave_mdm_name mdm2 --add_tb_name tb"
	err = xplatform.GetInstance().Run.Command(changeClusterCmdline, changeClusterModeCheck, "")
	if err != nil {
		log.Errorln("Change ScaleIO to 3 Node Cluster Failed:", err)
		log.Infoln("CreateCluster LEAVE")
		return err
	}

	time.Sleep(time.Duration(common.DelayBetweenCommandsInSeconds) * time.Second)

	renameCmdline := "scli --mdm_ip " + pri.IPAddress + " --rename_system --new_name scaleio"
	err = xplatform.GetInstance().Run.Command(renameCmdline, clusterRenameCheck, "")
	if err != nil {
		log.Errorln("Cluster Rename Failed:", err)
		log.Infoln("InitializeCluster LEAVE")
		return err
	}

	time.Sleep(time.Duration(common.DelayBetweenCommandsInSeconds) * time.Second)

	log.Infoln("CreateCluster Succeeded")
	log.Infoln("CreateCluster LEAVE")
	return nil
}

//GatewaySetup for setting up the ScaleIO gateway for API use
func (mm *MdmManager) GatewaySetup(state *types.ScaleIOFramework) (bool, error) {
	log.Infoln("GatewaySetup ENTER")

	pri, errPri := common.GetPrimaryMdmNode(state)
	if errPri != nil {
		log.Errorln("getPrimaryMdmNode Failed:", errPri)
		log.Infoln("GatewaySetup LEAVE")
		return false, errPri
	}
	sec, errSec := common.GetSecondaryMdmNode(state)
	if errSec != nil {
		log.Errorln("getSecondaryMdmNode Failed:", errSec)
		log.Infoln("GatewaySetup LEAVE")
		return false, errSec
	}

	//Install LIA
	liaVer, liaVerErr := xplatform.GetInstance().Inst.ParseVersionFromFilename(mm.LiaPackageDownload)
	liaInst, liaInstErr := xplatform.GetInstance().Inst.GetInstalledVersion(mm.LiaPackageName, true)
	log.Debugln("liaVer:", liaVer)
	log.Debugln("liaVerErr:", liaVerErr)
	log.Debugln("liaInst:", liaInst)
	log.Debugln("liaInstErr:", liaInstErr)

	if liaVerErr != nil || liaInstErr != nil || liaVer != liaInst {
		log.Infoln("Installing", mm.LiaPackageName)

		localLia, err := xplatform.GetInstance().Inst.DownloadPackage(mm.LiaPackageDownload)
		if err != nil {
			log.Errorln("Error downloading LIA package:", err)
			log.Infoln("PrimaryMDM LEAVE")
			return false, err
		}

		liaInstallCmd := strings.Replace(mm.LiaInstallCmd, "{LocalLia}", localLia, -1)
		log.Infoln("liaInstallCmd:", liaInstallCmd)

		err = xplatform.GetInstance().Run.Command(liaInstallCmd, mm.LiaInstallCheck, "")
		if err != nil {
			log.Errorln("Install LIA Failed:", err)
			log.Infoln("GatewaySetup LEAVE")
			return false, err
		}

		installIDCmdline := "scli --query_all | grep \"Installation ID\" | sed -n -e 's/^.*ID: //p'"
		output, err := xplatform.GetInstance().Run.CommandOutput(installIDCmdline)
		if err != nil {
			log.Errorln("Install LIA Failed:", err)
			log.Infoln("GatewaySetup LEAVE")
			return false, err
		}

		dumpIDCmdline := "echo " + output + " > /opt/emc/scaleio/lia/cfg/installation_id.txt"
		output, err = xplatform.GetInstance().Run.CommandOutput(dumpIDCmdline)
		if err != nil || len(output) > 0 {
			log.Errorln("Install LIA Failed:", err)
			log.Infoln("GatewaySetup LEAVE")
			return false, err
		}
	} else {
		log.Infoln(mm.LiaPackageName, "is already installed")
		time.Sleep(time.Duration(common.DelayIfInstalledInSeconds) * time.Second)
	}

	//Install Gateway
	gwVer, gwVerErr := xplatform.GetInstance().Inst.ParseVersionFromFilename(mm.GatewayPackageDownload)
	gwInst, gwInstErr := xplatform.GetInstance().Inst.GetInstalledVersion(mm.GatewayPackageName, true)
	log.Debugln("gwVer:", gwVer)
	log.Debugln("gwVerErr:", gwVerErr)
	log.Debugln("gwInst:", gwInst)
	log.Debugln("gwInstErr:", gwInstErr)

	if gwVerErr != nil || gwInstErr != nil || gwVer != gwInst {
		log.Infoln("Installing", mm.GatewayPackageName)

		localGw, err := xplatform.GetInstance().Inst.DownloadPackage(mm.GatewayPackageDownload)
		if err != nil {
			log.Errorln("Error downloading Gateway package:", err)
			log.Infoln("PrimaryMDM LEAVE")
			return false, err
		}

		gatewayInstallCmd := strings.Replace(mm.GatewayInstallCmd, "{LocalGw}", localGw, -1)

		err = xplatform.GetInstance().Run.Command(gatewayInstallCmd, mm.GatewayInstallCheck, "")
		if err != nil {
			log.Errorln("Install GW Failed:", err)
			log.Infoln("GatewaySetup LEAVE")
			return false, err
		}

		bypasssecCmdline := "sed -i 's/security.bypass_certificate_check=false/security.bypass_certificate_check=true/' /opt/emc/scaleio/gateway/webapps/ROOT/WEB-INF/classes/gatewayUser.properties"
		output, err := xplatform.GetInstance().Run.CommandOutput(bypasssecCmdline)
		if err != nil || len(output) > 0 {
			log.Errorln("Configure By-Pass Security Check Failed:", err)
			log.Infoln("GatewaySetup LEAVE")
			return false, err
		}

		writemdmCmdline := "sed -i 's/mdm.ip.addresses=/mdm.ip.addresses='" + pri.IPAddress +
			"','" + sec.IPAddress + "'/' /opt/emc/scaleio/gateway/webapps/ROOT/WEB-INF/classes/gatewayUser.properties"
		output, err = xplatform.GetInstance().Run.CommandOutput(writemdmCmdline)
		if err != nil || len(output) > 0 {
			log.Errorln("Configure MDM to Gateway Failed:", err)
			log.Infoln("GatewaySetup LEAVE")
			return false, err
		}
	} else {
		log.Infoln(mm.GatewayPackageName, "is already installed")
		time.Sleep(time.Duration(common.DelayIfInstalledInSeconds) * time.Second)
	}

	if gwInst == "" && gwInstErr == nil {
		log.Debugln("No previous install of", mm.GatewayPackageName,
			"exists. Reboot required!")
		log.Infoln("GatewaySetup LEAVE")
		return true, nil
	}

	log.Infoln("GatewaySetup Succeeded")
	log.Infoln("GatewaySetup LEAVE")
	return false, nil
}
