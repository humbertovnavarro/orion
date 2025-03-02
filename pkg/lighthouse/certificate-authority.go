package lighthouse

import (
	"os"
	"strings"
	"time"

	"github.com/humbertovnavarro/orion/pkg/config"
	"github.com/humbertovnavarro/orion/pkg/etcd"
	"github.com/humbertovnavarro/orion/pkg/utils"
	"github.com/sirupsen/logrus"
)

var FETCH_TIMEOUT = 2 * time.Second

type certificateAuthority struct {
	Key  string
	Cert string
}

func Update() error {
	cert := &certificateAuthority{}
	err := cert.Pull()
	if err != nil {
		return err
	}
	err = cert.ToDisk()
	if err != nil {
		return err
	}
	return nil
}

func SeedCertificates() (err error) {
	cert := &certificateAuthority{}

	err = cert.Pull()

	if err != nil {
		logrus.Info("Could not pull lighthouse certificate. Populating from disk.")
		err := cert.FromDisk()
		if err != nil {
			return err
		}
		logrus.Info("Pushing lighthouse certificates to etcd.")
		err = cert.Push()
		if err != nil {
			return err
		}
		return nil
	}

	logrus.Info("loaded cluster ca certs")

	timeString := time.Now().Local().Format(time.DateTime)
	timeString = strings.ReplaceAll(timeString, " ", "_")

	err = utils.FileCopy(config.CertificateAuthorityKeyLocation, config.CertificateAuthorityKeyLocation+timeString+".bak")
	if err != nil {
		return err
	}

	logrus.Info("backed up ca authority key")

	err = utils.FileCopy(config.CertificateAuthorityLocation, config.CertificateAuthorityLocation+timeString+".bak")
	if err != nil {
		return err
	}

	logrus.Info("backed up ca cert")

	err = cert.ToDisk()

	if err != nil {
		return err
	}

	logrus.Info("wrote cluster ca certs to disk")
	return
}

// Syncs certs on disk with etcd
func (c *certificateAuthority) Update() error {
	err := c.Pull()
	if err != nil {
		return err
	}
	err = c.ToDisk()
	if err != nil {
		return err
	}
	return nil
}

func (c *certificateAuthority) ToDisk() error {
	// Write Key
	file, err := os.Create(config.CertificateAuthorityKeyLocation)
	if err != nil {
		return err
	}

	_, err = file.WriteString(c.Key)
	if err != nil {
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}
	// Write CA
	file, err = os.Create(config.CertificateAuthorityLocation)
	if err != nil {
		return err
	}

	_, err = file.WriteString(c.Cert)
	if err != nil {
		return err
	}

	err = file.Close()
	if err != nil {
		return err
	}

	return nil
}

func (c *certificateAuthority) FromDisk() error {
	key, err := os.ReadFile(config.CertificateAuthorityKeyLocation)
	if err != nil {
		return err
	}

	cert, err := os.ReadFile(config.CertificateAuthorityLocation)
	if err != nil {
		return err
	}

	c.Cert = string(cert)
	c.Key = string(key)
	return nil
}

func (c *certificateAuthority) Pull() error {
	key, err := etcd.Get(etcd.KEY_NEBULA_CA)
	if err != nil {
		return err
	}

	cert, err := etcd.Get(etcd.KEY_NEBULA_CA_KEY)
	if err != nil {
		return err
	}

	c.Key = key
	c.Cert = cert

	return nil
}

func (c *certificateAuthority) Push() error {
	err := etcd.Put(etcd.KEY_NEBULA_CA_KEY, c.Key)
	if err != nil {
		return err
	}
	err = etcd.Put(etcd.KEY_NEBULA_CA, c.Cert)
	if err != nil {
		return err
	}
	return nil
}
