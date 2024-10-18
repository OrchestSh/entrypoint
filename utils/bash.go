package utils

import (
	"math"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/google/shlex"
)

// GetUserUIDs returns the user UID and default GID
func GetUserUIDs(username string) (uint32, uint32, error) {
	u, err := user.Lookup(username)
	if err != nil {
		return 0, 0, err
	}

	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return 0, 0, err
	}
	if uid < 0 || uid > math.MaxUint32 {
		uid = 0
	}

	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return 0, 0, err
	}
	if gid < 0 || uid > math.MaxUint32 {
		gid = 0
	}

	return uint32(uid), uint32(gid), nil
}

// RunAndLogCmdAs is just a wrapper for exec, run the command and print the stdout and stderr,
// also can be executed as another user if user != ""
func RunAndLogCmdAs(cmd string, userName string, env []string) error {
	var err error
	if userName == "" {
		userN, err := user.Current()
		if err != nil {
			return err
		}
		userName = userN.Name
	}
	uid, gid, err := GetUserUIDs(userName)
	if err != nil {
		return err
	}

	parts, err := shlex.Split(cmd)
	if err != nil {
		return err
	}
	cmdS := exec.Command(parts[0], parts[1:]...) // #nosec G204
	cmdS.SysProcAttr = &syscall.SysProcAttr{}
	cmdS.SysProcAttr.Credential = &syscall.Credential{
		Uid: uid,
		Gid: gid,
	}

	cmdS.Env = os.Environ()
	if env != nil {
		cmdS.Env = append(cmdS.Env, env...)
	}

	cmdS.Stdout = os.Stdout
	cmdS.Stderr = os.Stdout
	log.Debugf("Starting command: %s", cmd)
	err = cmdS.Start()
	if err != nil {
		log.Errorf("Error starting the command: %s", err.Error())
		return err
	}

	err = cmdS.Wait()
	if err != nil {
		log.Errorf("Error running the command: %s", err.Error())
		return err
	}
	return nil
}

// Copy copies a single file from src to dst
func Copy(src, dst string) error {
	srcF, err := os.Open(filepath.Clean(src))
	if err != nil {
		return err
	}
	defer srcF.Close()
	dstF, err := os.Create(filepath.Clean(dst))
	if err != nil {
		return err
	}
	defer dstF.Close()
	_, err = dstF.ReadFrom(srcF)
	return err
}
