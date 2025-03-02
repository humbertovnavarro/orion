package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/humbertovnavarro/orion/pkg/config"
	"github.com/humbertovnavarro/orion/pkg/etcd"
	"github.com/humbertovnavarro/orion/pkg/utils"
)

type Client struct {
	IpAddress net.IP
	Groups    []string
	Subnets   []net.IPNet
	Key       []byte
	Cert      []byte
	Updated   time.Time
}

func (c *Client) GroupCSV() string {
	return strings.Join(c.Groups, ",")
}

func (c *Client) SubnetCSV() string {
	out := make([]string, len(c.Subnets))
	for i, sub := range c.Subnets {
		out[i] = sub.String()
	}
	return strings.Join(out, ",")
}

func (c *Client) Push() error {
	ipString := c.ToString()
	clientKey := fmt.Sprintf(etcd.KEY_CLIENTS, ipString)
	clientCa := fmt.Sprintf(etcd.KEY_CLIENT_CERT, ipString)
	clientGroups := fmt.Sprintf(etcd.KEY_CLIENT_GROUPS, c.GroupCSV())
	clientSubnets := fmt.Sprintf(etcd.KEY_CLIENT_SUBNETS, c.SubnetCSV())

	err := etcd.Put(clientKey, c.IpAddress.String())
	if err != nil {
		return err
	}

	err = etcd.Put(clientCa, string(c.Cert))
	if err != nil {
		return err
	}

	err = etcd.Put(clientKey, string(c.Key))
	if err != nil {
		return err
	}

	err = etcd.Put(clientGroups, strings.Join(c.Groups, ","))
	if err != nil {
		return err
	}

	err = etcd.Put(clientSubnets, c.SubnetCSV())
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Pull() error {
	ipString := c.ToString()
	clientKey := fmt.Sprintf(etcd.KEY_CLIENTS, ipString)
	clientCert := fmt.Sprintf(etcd.KEY_CLIENT_CERT, ipString)
	clientGroups := fmt.Sprintf(etcd.KEY_CLIENT_GROUPS, c.GroupCSV())
	clientSubnets := fmt.Sprintf(etcd.KEY_CLIENT_SUBNETS, c.SubnetCSV())

	fetchedIp, err := etcd.Get(ipString)
	if err != nil {
		return err
	}
	c.IpAddress = net.ParseIP(fetchedIp)

	fetchedClientKey, err := etcd.Get(clientKey)
	if err != nil {
		return err
	}
	c.Key = []byte(fetchedClientKey)

	fetchedClientCert, err := etcd.Get(clientCert)
	if err != nil {
		return err
	}
	c.Cert = []byte(fetchedClientCert)

	fetchedClientGroups, err := etcd.Get(clientGroups)
	if err != nil {
		return err
	}
	c.Groups = strings.Split(fetchedClientGroups, ",")

	fetchedClientSubnets, err := etcd.Get(clientSubnets)
	if err != nil {
		return err
	}

	fetchedClientSubnetStrings := strings.Split(fetchedClientSubnets, ",")
	fetchedSubnets := make([]net.IPNet, len(fetchedClientSubnets))

	for i, s := range fetchedClientSubnetStrings {
		_, n, err := net.ParseCIDR(s)
		if err != nil {
			return err
		}
		if n == nil {
			return errors.New("invalid subnet in client subnets")
		}
		fetchedSubnets[i] = *n
	}

	return nil
}

func (c *Client) ToDisk() (err error) {
	subnetArgs := ""
	numSubnets := len(c.Subnets)

	for i, s := range c.Subnets {
		subnetArgs += s.String()
		if i != numSubnets {
			subnetArgs += ","
		}
	}

	groupArgs := ""
	numGroups := len(c.Groups)

	for i, s := range c.Groups {

		if !utils.IsAlphanumeric(s) {
			err = fmt.Errorf("invalid group string while generating cert for client %s", c.IpAddress.String())
			return
		}

		if i != numGroups {
			groupArgs += ","
		}

	}

	certPath := path.Join(config.ClientConfigLocation, c.IpAddress.String()+".crt")
	keyPath := path.Join(config.ClientConfigLocation, c.IpAddress.String()+".key")

	args := []string{
		"sign",
		"-name",
		c.IpAddress.String(),
		"-ca-crt",
		config.CertificateAuthorityLocation,
		"-ca-key",
		config.CertificateAuthorityKeyLocation,
		"-duration",
		config.CertificateDuration,
		"-ip",
		c.IpAddress.String(),
		"-subnets",
		subnetArgs,
		"-out-crt",
		certPath,
		"-out-key",
		keyPath,
	}

	if numGroups > 0 {
		args = append(args, groupArgs)
	}

	cmd := exec.Command(config.NebulaCertBin, args...)

	output, err := cmd.CombinedOutput()

	if err != nil {
		err = fmt.Errorf("%s\n%s\n%s", strings.Join(args, " "), output, err.Error())
		return
	}

	return
}

func (c *Client) ToString() string {
	out, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return err.Error()
	}
	return string(out)
}

func Clone() {

}
