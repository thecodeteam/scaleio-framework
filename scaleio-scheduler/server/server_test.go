package server

import (
	"bytes"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	assert "github.com/stretchr/testify/assert"

	config "github.com/codedellemc/scaleio-framework/scaleio-scheduler/config"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

const (
	TestInputFile  = "/tmp/inputfile.txt"
	TestOutputFile = "/tmp/outputfile.txt"
)

var server *RestServer

func TestMain(m *testing.M) {
	log.SetLevel(log.InfoLevel)
	log.SetOutput(os.Stdout)

	//create config object
	cfg := config.NewConfig()

	//alt executor path
	cfg.AltExecutorPath = TestInputFile

	server = NewRestServer(cfg)

	server.State.ScaleIO.Nodes = append(server.State.ScaleIO.Nodes, &types.ScaleIONode{
		AgentID:    "127.0.0.1",
		ExecutorID: "executor1",
		Persona:    types.PersonaMdmPrimary,
		State:      types.StateUnknown,
	})
	server.State.ScaleIO.Nodes = append(server.State.ScaleIO.Nodes, &types.ScaleIONode{
		AgentID:    "127.0.0.2",
		ExecutorID: "executor2",
		Persona:    types.PersonaMdmSecondary,
		State:      types.StateUnknown,
	})
	server.State.ScaleIO.Nodes = append(server.State.ScaleIO.Nodes, &types.ScaleIONode{
		AgentID:    "127.0.0.3",
		ExecutorID: "executor3",
		Persona:    types.PersonaTb,
		State:      types.StateUnknown,
	})
	server.State.ScaleIO.Nodes = append(server.State.ScaleIO.Nodes, &types.ScaleIONode{
		AgentID:    "127.0.0.4",
		ExecutorID: "executor4",
		Persona:    types.PersonaNode,
		State:      types.StateUnknown,
	})

	server.State.ScaleIO.AdminPassword = "Scaleio123"
	server.State.ScaleIO.Deb.DebMdm = "mymdm"

	//wait 5 seconds for server to come up
	time.Sleep(5 * time.Second)

	//log.Infoln("Start tests")
	m.Run()
}

func TestDownload(t *testing.T) {
	//create a test file to serve
	input, err := os.Create(TestInputFile)
	assert.NotNil(t, input)
	assert.NoError(t, err)

	input.WriteString("testing")
	input.Close()

	//create a downloaded file
	output, err := os.Create(TestOutputFile)
	assert.NotNil(t, output)
	assert.NoError(t, err)

	//get the "executor" file
	resp, err := http.Get("http://" + server.Config.RestAddress + ":" +
		strconv.Itoa(server.Config.RestPort) + "/scaleio-executor")

	assert.Equal(t, resp.StatusCode, http.StatusOK)

	defer resp.Body.Close()

	//save it
	_, err = io.Copy(output, resp.Body)
	assert.NoError(t, err)

	//close and flush
	output.Close()

	//sha512 of source
	sha1 := sha512.New()
	byte1, _ := ioutil.ReadFile(TestInputFile)
	sha1.Write(byte1)
	str1 := fmt.Sprintf("%x", sha1.Sum(nil))
	log.Debugln("Data:", string(byte1), "sha512:", str1)

	//sha512 of downloaded file
	sha2 := sha512.New()
	byte2, _ := ioutil.ReadFile(TestOutputFile)
	sha2.Write(byte2)
	str2 := fmt.Sprintf("%x", sha2.Sum(nil))
	log.Debugln("Data:", string(byte2), "sha512:", str2)

	//check hash to make sure they are equal
	assert.Equal(t, str1, str2)

	os.Remove(TestInputFile)
	os.Remove(TestOutputFile)
}

func TestVersion(t *testing.T) {
	url := "http://" + server.Config.RestAddress + ":" +
		strconv.Itoa(server.Config.RestPort) + "/version"

	req, err := http.NewRequest("GET", url, nil)
	//req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NotNil(t, resp)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1048576))
	assert.NotNil(t, body)
	assert.NoError(t, err)

	var ver types.Version
	err = json.Unmarshal(body, &ver)
	assert.NotNil(t, ver)
	assert.NoError(t, err)

	assert.Equal(t, ver.VersionInt, config.VersionInt)
	assert.Equal(t, ver.VersionStr, config.VersionStr)
}

func TestState(t *testing.T) {
	url := "http://" + server.Config.RestAddress + ":" +
		strconv.Itoa(server.Config.RestPort) + "/api/state"

	req, err := http.NewRequest("GET", url, nil)
	//req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NotNil(t, resp)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1048576))
	assert.NotNil(t, body)
	assert.NoError(t, err)

	var state types.ScaleIOFramework
	err = json.Unmarshal(body, &state)
	assert.NotNil(t, state)
	assert.NoError(t, err)

	assert.Equal(t, "scaleio", state.ScaleIO.ClusterName)
	assert.Equal(t, "pd", state.ScaleIO.ProtectionDomain)
	assert.Equal(t, "sp", state.ScaleIO.StoragePool)
	assert.Equal(t, "Scaleio123", state.ScaleIO.AdminPassword)

	for _, node := range state.ScaleIO.Nodes {
		switch node.ExecutorID {
		case "executor1":
			log.Debugln("ExecutorID:", node.ExecutorID, "Persona:", node.Persona)
			assert.Equal(t, types.PersonaMdmPrimary, node.Persona)
		case "executor2":
			log.Debugln("ExecutorID:", node.ExecutorID, "Persona:", node.Persona)
			assert.Equal(t, types.PersonaMdmSecondary, node.Persona)
		case "executor3":
			log.Debugln("ExecutorID:", node.ExecutorID, "Persona:", node.Persona)
			assert.Equal(t, types.PersonaTb, node.Persona)
		case "executor4":
			log.Debugln("ExecutorID:", node.ExecutorID, "Persona:", node.Persona)
			assert.Equal(t, types.PersonaNode, node.Persona)
		}
	}

	assert.Equal(t, "mymdm", state.ScaleIO.Deb.DebMdm)
}

func TestNodeStateOk(t *testing.T) {
	url := "http://" + server.Config.RestAddress + ":" +
		strconv.Itoa(server.Config.RestPort) + "/api/node/state"

	state := types.UpdateNode{
		Acknowledged: false,
		ExecutorID:   "executor1",
		State:        types.StatePrerequisitesInstalled,
	}

	response, err := json.MarshalIndent(state, "", "  ")
	assert.NotNil(t, response)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(response))
	assert.NotNil(t, req)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NotNil(t, resp)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1048576))
	assert.NotNil(t, body)
	assert.NoError(t, err)

	log.Debugln("response Status:", resp.Status)
	log.Debugln("response Headers:", resp.Header)
	log.Debugln("response Body:", string(body))

	var newstate types.UpdateNode
	err = json.Unmarshal(body, &newstate)
	assert.NotNil(t, state)
	assert.NoError(t, err)

	log.Debugln("Acknowledged:", newstate.Acknowledged)
	log.Debugln("ExecutorID:", newstate.ExecutorID)
	log.Debugln("State:", newstate.State)

	assert.Equal(t, true, newstate.Acknowledged)
	assert.Equal(t, "executor1", newstate.ExecutorID)
	assert.Equal(t, types.StatePrerequisitesInstalled, newstate.State)
}

func TestNodeStateBad(t *testing.T) {
	url := "http://" + server.Config.RestAddress + ":" +
		strconv.Itoa(server.Config.RestPort) + "/api/node/state"

	state := types.UpdateNode{
		Acknowledged: false,
		ExecutorID:   "executor5",
		State:        types.StatePrerequisitesInstalled,
	}

	response, err := json.MarshalIndent(state, "", "  ")
	assert.NotNil(t, response)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(response))
	assert.NotNil(t, req)
	assert.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NotNil(t, resp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1048576))
	assert.NotNil(t, body)
	assert.NoError(t, err)

	assert.Equal(t, "Unable to find the Executor", strings.TrimSpace(string(body)))
}

func TestNodeAdd(t *testing.T) {
	url := "http://" + server.Config.RestAddress + ":" +
		strconv.Itoa(server.Config.RestPort) + "/api/node/cluster"

	state := types.AddNode{
		Acknowledged: false,
		ExecutorID:   "executor1",
	}

	response, err := json.MarshalIndent(state, "", "  ")
	assert.NotNil(t, response)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(response))
	assert.NotNil(t, req)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NotNil(t, resp)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1048576))
	assert.NotNil(t, body)
	assert.NoError(t, err)

	log.Debugln("response Status:", resp.Status)
	log.Debugln("response Headers:", resp.Header)
	log.Debugln("response Body:", string(body))

	var newstate types.AddNode
	err = json.Unmarshal(body, &newstate)
	assert.NotNil(t, state)
	assert.NoError(t, err)

	log.Debugln("Acknowledged:", newstate.Acknowledged)
	log.Debugln("ExecutorID:", newstate.ExecutorID)

	assert.Equal(t, true, newstate.Acknowledged)
	assert.Equal(t, "executor1", newstate.ExecutorID)
}

func TestNodePing(t *testing.T) {
	url := "http://" + server.Config.RestAddress + ":" +
		strconv.Itoa(server.Config.RestPort) + "/api/node/ping"

	state := types.PingNode{
		Acknowledged: false,
		ExecutorID:   "executor1",
	}

	response, err := json.MarshalIndent(state, "", "  ")
	assert.NotNil(t, response)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(response))
	assert.NotNil(t, req)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NotNil(t, resp)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1048576))
	assert.NotNil(t, body)
	assert.NoError(t, err)

	log.Debugln("response Status:", resp.Status)
	log.Debugln("response Headers:", resp.Header)
	log.Debugln("response Body:", string(body))

	var newstate types.AddNode
	err = json.Unmarshal(body, &newstate)
	assert.NotNil(t, state)
	assert.NoError(t, err)

	log.Debugln("Acknowledged:", newstate.Acknowledged)
	log.Debugln("ExecutorID:", newstate.ExecutorID)

	assert.Equal(t, true, newstate.Acknowledged)
	assert.Equal(t, "executor1", newstate.ExecutorID)
}
