// Package gdctx is the runtime context of GlusterD
//
// This file implements the global runtime context for GlusterD.
// Any package that needs access to the GlusterD global runtime context just
// needs to import this package.
package gdctx

import (
	"errors"
	"io/ioutil"
	"os"
	"path"

	"github.com/gluster/glusterd2/utils"
	"github.com/gluster/glusterd2/version"

	log "github.com/sirupsen/logrus"
	"github.com/pborman/uuid"
	config "github.com/spf13/viper"
)

// Any object that is a part of the GlusterD context and needs to be available
// to other packages should be declared here as exported global variables
var (
	MyUUID    uuid.UUID
	Restart   bool // Indicates if its a fresh install or not (based on presence/absence of UUID file)
	OpVersion int
	HostIP    string
	HostName  string
)

// SetHostnameAndIP will initialize HostIP and HostName global variables
func SetHostnameAndIP() error {
	hostIP, err := utils.GetLocalIP()
	if err != nil {
		return err
	}
	HostIP = hostIP

	hostName, err := os.Hostname()
	if err != nil {
		return err
	}
	HostName = hostName

	return nil
}

// SetUUID will generate (or use if present) and set MyUUID global variable
func SetUUID() error {
	uuidFile := path.Join(config.GetString("localstatedir"), "uuid")
	ubytes, err := ioutil.ReadFile(uuidFile)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			// generate new UUID and write to file
			MyUUID = uuid.NewRandom()
			if err := ioutil.WriteFile(uuidFile, []byte(MyUUID.String()), 0644); err != nil {
				log.WithError(err).WithField("path", uuidFile).Debug(
					"failed to write UUID to file")
				return err
			}
			log.WithField("uuid", MyUUID.String()).Info("Generated new UUID")
			return nil
		default:
			log.WithError(err).WithField("path", uuidFile).Debug(
				"failed to read UUID from file")
			return err
		}
	}
	// use the UUID found in file
	MyUUID = uuid.Parse(string(ubytes))
	if MyUUID == nil {
		return errors.New("failed to parse UUID found in file")
	}
	log.WithField("uuid", MyUUID.String()).Info("Found existing UUID")

	Restart = true

	return nil
}

func init() {
	OpVersion = version.MaxOpVersion
}
