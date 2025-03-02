package main

import (
	"github.com/humbertovnavarro/orion/pkg/lighthouse"
	"github.com/sirupsen/logrus"
)

func main() {
	err := lighthouse.SeedCertificates()
	if err != nil {
		logrus.Error(err)
		logrus.Error("If this is the first time you started orion - you need to generate a CA cert")
	}

}
