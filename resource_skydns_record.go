package main

import (
	"encoding/json"
	"github.com/coreos/etcd/client"
	"github.com/hashicorp/terraform/helper/schema"
	"golang.org/x/net/context"
	"strings"
)

func skydnsRecord() *schema.Resource {
	return &schema.Resource{
		Create: createSkyDnsRecord,
		Read:   readSkyDnsRecord,
		Delete: deleteSkyDnsRecord,
		Exists: existsSkyDnsRecord,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ttl": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"records": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
				ForceNew: true,
				Set:      schema.HashString,
			},
		},
	}
}

type SkyDnsRecord struct {
	Host  string `json:"host"`
	TTL   int    `json:"ttl,omitempty"`
	Group string `json:"group,omitempty"`
}

func reverse(xs []string) []string {
	for i, j := 0, len(xs)-1; i < j; i, j = i+1, j-1 {
		xs[i], xs[j] = xs[j], xs[i]
	}
	return xs
}

func createSkyDnsRecord(d *schema.ResourceData, meta interface{}) error {
	config := meta.(Config)
	c := config.Client
	kapi := client.NewKeysAPI(*c)

	name := d.Get("name").(string)
	ttl := d.Get("ttl").(int)
	records := d.Get("records").(*schema.Set).List()

	// TODO check if there's something in etcd (a directory or value node)
	keyPrefix := config.SkyDnsPrefix + strings.Join(reverse(strings.Split(name, ".")), "/")

	if len(records) > 0 {
		for _, record := range records {
			key := keyPrefix + "/record-" + strings.Replace(record.(string), ".", "_", -1)
			value := SkyDnsRecord{
				Host:  record.(string),
				TTL:   ttl,
				Group: name,
			}
			// TODO error handling
			jsonValueBytes, _ := json.Marshal(value)

			// TODO inspect response && error handling
			kapi.Set(context.Background(), key, string(jsonValueBytes), nil)
		}
	} else {
		// TODO inspect response && error handling
		keys, _ := kapi.Get(context.Background(), keyPrefix, nil)
		for _, node := range keys.Node.Nodes {
			// TODO inspect response && error handling
			kapi.Delete(context.Background(), node.Key, nil)
		}
		// TODO inspect response && error handling
		kapi.Delete(context.Background(), keyPrefix, &client.DeleteOptions{Dir: true})
	}

	d.SetId(name)
	return nil
}

func readSkyDnsRecord(d *schema.ResourceData, meta interface{}) error {
	config := meta.(Config)
	c := config.Client
	kapi := client.NewKeysAPI(*c)

	name := d.Get("name").(string)
	keyPrefix := config.SkyDnsPrefix + strings.Join(reverse(strings.Split(name, ".")), "/")

	keys, err := kapi.Get(context.Background(), keyPrefix, nil)
	if err != nil {
		return err
	}

	records := make([]string, 0, len(keys.Node.Nodes))
	ttl := -1
	// TODO check if the node really is a directory node
	for _, node := range keys.Node.Nodes {
		var record SkyDnsRecord
		err = json.Unmarshal([]byte(node.Value), &record)
		if err != nil {
			return err
		}
		ttl = record.TTL
		records = append(records, record.Host)
	}

	d.Set("records", records)
	if ttl > 0 {
		d.Set("ttl", ttl)
	}
	return nil
}

func deleteSkyDnsRecord(d *schema.ResourceData, meta interface{}) error {
	config := meta.(Config)
	c := config.Client
	kapi := client.NewKeysAPI(*c)

	name := d.Get("name").(string)
	keyPrefix := config.SkyDnsPrefix + strings.Join(reverse(strings.Split(name, ".")), "/")

	// TODO error handling
	keys, _ := kapi.Get(context.Background(), keyPrefix, nil)
	// TODO check if the node really is a directory node
	for _, node := range keys.Node.Nodes {
		// TODO inspect response && error handling
		kapi.Delete(context.Background(), node.Key, nil)
	}
	// TODO inspect response && error handling
	kapi.Delete(context.Background(), keyPrefix, &client.DeleteOptions{Dir: true})

	return nil
}

func existsSkyDnsRecord(d *schema.ResourceData, meta interface{}) (bool, error) {
	config := meta.(Config)
	c := config.Client
	kapi := client.NewKeysAPI(*c)

	name := d.Get("name").(string)
	keyPrefix := config.SkyDnsPrefix + strings.Join(reverse(strings.Split(name, ".")), "/")

	// TODO: check if the node is a directory node?
	_, err := kapi.Get(context.Background(), keyPrefix, nil)
	if err != nil {
		if client.IsKeyNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	} else {
		return true, nil
	}
}
