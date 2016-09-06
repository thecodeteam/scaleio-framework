package exec

import (
	"bufio"
	"bytes"
	"errors"
	"os/exec"
	"regexp"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

var (
	//ErrBufferCreateFailed creating the buffer failed
	ErrBufferCreateFailed = errors.New("Unable to create the buffer object")

	//ErrScannerCreateFailed creating the scanner failed
	ErrScannerCreateFailed = errors.New("Unable to create the scanner object")

	//ErrReaderCreateFailed creating the reader failed
	ErrReaderCreateFailed = errors.New("Unable to create the reader object")

	//ErrCommandCreateFailed creating the command failed
	ErrCommandCreateFailed = errors.New("Unable to create the command object")

	//ErrExecuteFailed installation package failed
	ErrExecuteFailed = errors.New("The command line failed to execute correctly")
)

//RunCommand executes a command that monitors output for success or failure
func RunCommand(cmdLine string, successRegex string, failureRegex string) error {
	log.Debugln("RunCommand ENTER")
	log.Debugln("Cmdline:", cmdLine)
	log.Debugln("SuccessRegex:", successRegex)
	log.Debugln("FailureRegex:", failureRegex)

	cmd := exec.Command("bash", "-c", cmdLine)
	if cmd == nil {
		log.Errorln("Error creating cmd")
		log.Debugln("RunCommand LEAVE")
		return ErrCommandCreateFailed
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorln("Error starting Cmd:", err)
		log.Debugln("RunCommand LEAVE")
		return err
	}

	readbuffer := bytes.NewBuffer(out)
	if readbuffer == nil {
		log.Errorln("Error creating buffer")
		log.Debugln("RunCommand LEAVE")
		return ErrBufferCreateFailed
	}

	reader := bufio.NewScanner(readbuffer)
	if reader == nil {
		log.Errorln("Error creating reader")
		log.Debugln("RunCommand LEAVE")
		return ErrReaderCreateFailed
	}

	failure := false
	succeeded := false
	for reader.Scan() {
		line := reader.Text()
		log.Debugln("Line:", line)
		if failure {
			continue
		}
		if len(failureRegex) > 0 {
			myfail, _ := regexp.MatchString(failureRegex, line)
			if myfail {
				log.Debugln("Line Matched - FAILURE!")
				failure = true
			}
		}
		if succeeded {
			continue
		}
		if len(successRegex) > 0 {
			mysucceed, _ := regexp.MatchString(successRegex, line)
			if mysucceed {
				log.Debugln("Line Matched - SUCCEEDED!")
				succeeded = true
			}
		}
	}

	if failure {
		log.Debugln("Cmdline explicitly failed to execute correctly")
		log.Debugln("RunCommand LEAVE")
		return ErrExecuteFailed
	}
	if succeeded {
		log.Debugln("Cmdline executed successful")
		log.Debugln("RunCommand LEAVE")
		return nil
	}

	log.Debugln("Cmdline implicitly failed to execute correctly")
	log.Debugln("RunCommand LEAVE")
	return ErrExecuteFailed
}

//RunCommandEx executes a command that monitors output for success or failure
func RunCommandEx(cmdLine string, successRegex string, failureRegex string, waitInSec int) error {
	log.Debugln("RunCommandEx ENTER")
	log.Debugln("Cmdline:", cmdLine)
	log.Debugln("SuccessRegex:", successRegex)
	log.Debugln("FailureRegex:", failureRegex)

	cmd := exec.Command("bash", "-c", cmdLine)
	if cmd == nil {
		log.Errorln("Error creating cmd")
		log.Debugln("RunCommandEx LEAVE")
		return ErrCommandCreateFailed
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Errorln("Error getting StdoutPipe:", err)
		log.Debugln("RunCommandEx LEAVE")
		return err
	}

	err = cmd.Start()
	if err != nil {
		log.Errorln("Error on cmd start:", err)
		log.Debugln("RunCommandEx LEAVE")
		return err
	}

	stdoutScanner := bufio.NewScanner(stdout)
	if cmd == nil {
		log.Errorln("Error creating scanner")
		log.Debugln("RunCommandEx LEAVE")
		return ErrScannerCreateFailed
	}

	output := ""
	go func() {
		for stdoutScanner.Scan() {
			line := stdoutScanner.Text()
			log.Infoln(line)
			output += line
		}
	}()

	err = cmd.Wait()
	if err != nil {
		log.Warnln("Error on cmd wait:", err)
	}

	cmd.Process.Wait() //this should wait until all child processes are gone

	time.Sleep(time.Duration(waitInSec) * time.Second)

	outputBuffer := bytes.NewBuffer([]byte(output))
	if outputBuffer == nil {
		log.Errorln("Error creating buffer")
		log.Debugln("RunCommandEx LEAVE")
		return ErrBufferCreateFailed
	}

	outputScanner := bufio.NewScanner(outputBuffer)
	if outputScanner == nil {
		log.Errorln("Error creating reader")
		log.Debugln("RunCommandEx LEAVE")
		return ErrScannerCreateFailed
	}

	failure := false
	succeeded := false
	for outputScanner.Scan() {
		line := outputScanner.Text()
		log.Debugln("Line:", line)
		if failure {
			continue
		}
		if len(failureRegex) > 0 {
			myfail, _ := regexp.MatchString(failureRegex, line)
			if myfail {
				log.Debugln("Line Matched - FAILURE!")
				failure = true
			}
		}
		if succeeded {
			continue
		}
		if len(successRegex) > 0 {
			mysucceed, _ := regexp.MatchString(successRegex, line)
			if mysucceed {
				log.Debugln("Line Matched - SUCCEEDED!")
				succeeded = true
			}
		}
	}

	if failure {
		log.Debugln("Cmdline explicitly failed to execute correctly")
		log.Debugln("RunCommandEx LEAVE")
		return ErrExecuteFailed
	}
	if succeeded {
		log.Debugln("Cmdline executed successful")
		log.Debugln("RunCommandEx LEAVE")
		return nil
	}

	log.Debugln("Cmdline implicitly failed to execute correctly")
	log.Debugln("RunCommandEx LEAVE")
	return ErrExecuteFailed
}

//RunCommandOutput executes a command that returns the output
func RunCommandOutput(cmdLine string) (string, error) {
	log.Debugln("RunCommandOutput ENTER")
	log.Debugln("Cmdline:", cmdLine)

	cmd := exec.Command("bash", "-c", cmdLine)
	if cmd == nil {
		log.Errorln("Error creating cmd")
		log.Debugln("RunCommandOutput LEAVE")
		return "", ErrCommandCreateFailed
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorln("Error getting output:", err)
		log.Debugln("RunCommandOutput LEAVE")
		return "", err
	}

	output := strings.TrimSpace(string(out))

	log.Debugln("RunCommandOutput Succeeded")
	log.Debugln(output)
	log.Debugln("RunCommandOutput LEAVE")
	return output, nil
}
