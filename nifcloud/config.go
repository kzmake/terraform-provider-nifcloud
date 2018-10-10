package nifcloud

import (
	"fmt"
	"github.com/alice02/nifcloud-sdk-go/nifcloud"
	"github.com/alice02/nifcloud-sdk-go/nifcloud/credentials"
	"github.com/alice02/nifcloud-sdk-go/nifcloud/session"
	"github.com/alice02/nifcloud-sdk-go/service/computing"
)

type Config struct {
	AccessKey string
	SecretKey string
	Region    string

	Ec2Endpoint string
}

type NifcloudClient struct {
	computingconn *computing.Computing
}

func (c *Config) Client() (interface{}, error) {
	if c.AccessKey == "" {
		return nil, fmt.Errorf("[Err] No Access key for Nifcloud")
	}

	if c.SecretKey == "" {
		return nil, fmt.Errorf("[Err] No Secret key for Nifcloud")
	}

	if c.Region == "" {
		return nil, fmt.Errorf("[Err] No Region Name for Nifcloud")
	}

	sess := session.Must(session.NewSession(&nifcloud.Config{
		Region:      nifcloud.String(c.Region),
		Credentials: credentials.NewStaticCredentials(c.AccessKey, c.SecretKey, ""),
	}))

	var client NifcloudClient

	client.computingconn = computing.New(sess)

	return &client, nil
}
