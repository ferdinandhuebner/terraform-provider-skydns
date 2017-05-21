package main

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"strings"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"skydns_prefix": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "/skydns/",
			},
			"endpoints": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"username": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"password": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ca_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"cert_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"key_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"skydns_record": skydnsRecord(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(data *schema.ResourceData) (interface{}, error) {

	endpointsSchema := data.Get("endpoints").([]interface{})
	endpoints := make([]string, 0, len(endpointsSchema))
	for _, endpoint := range endpointsSchema {
		endpoints = append(endpoints, endpoint.(string))
	}

	prefix := data.Get("skydns_prefix").(string)
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}

	config := EtcdConfig{
		Endpoints: endpoints,
		Username:  data.Get("username").(string),
		Password:  data.Get("password").(string),
		CaFile:    data.Get("ca_file").(string),
		CertFile:  data.Get("cert_file").(string),
		KeyFile:   data.Get("key_file").(string),
	}

	client, err := config.Client()
	if err != nil {
		return nil, err
	}

	return Config{
		Client:       client,
		SkyDnsPrefix: prefix,
	}, nil
}
