package nifcloud

import (
	"fmt"
	"github.com/kzmake/nifcloud-sdk-go/nifcloud"
	"github.com/kzmake/nifcloud-sdk-go/nifcloud/credentials"
	"github.com/kzmake/nifcloud-sdk-go/nifcloud/session"
	"github.com/kzmake/nifcloud-sdk-go/service/computing"
)

type Config struct {
	AccessKey string
	SecretKey string
	Region    string
	Endpoint string
}

type NifcloudClient struct {
	computingconn *computing.Computing
}

func (c *Config) Client() (interface{}, error) {
	if c.Region == "" {
		return nil, fmt.Errorf("[Err] No Region Name for Nifcloud")
	}

	var credential *credentials.Credentials
	if c.AccessKey != "" && c.SecretKey != "" {
		credential = credentials.NewStaticCredentials(c.AccessKey, c.SecretKey, "")
	} else {
		credential = credentials.NewEnvCredentials()
	}

	config := nifcloud.Config{
		Region:      nifcloud.String(c.Region),
		Credentials: credential,
	}

	if c.Endpoint != "" {
		config.WithEndpoint(*nifcloud.String(c.Endpoint))
	}

	sess := session.Must(session.NewSession(&config))

	var client NifcloudClient

	client.computingconn = computing.New(sess)

	return &client, nil
}
