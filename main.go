package main

import (
	"encoding/json"
	"net"

	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const CONFIG_PATH = "/etc/nebula"

func main() {
	dbPath := CONFIG_PATH + "/nebula.db"
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.POST("/lighthouses", func(c *gin.Context) {
		privateAddress := c.Query("privateAddress")
		privateNetAddress := net.ParseIP(privateAddress)
		if privateNetAddress == nil {
			c.JSON(400, gin.H{"error": "address is invalid"})
			return
		}
		publicAddress := c.Query("publicAddress")
		publicNetAddress := net.ParseIP(publicAddress)
		if publicNetAddress == nil {
			c.JSON(400, gin.H{"error": "address is invalid"})
			return
		}

		privateAddress = privateNetAddress.String()
		publicAddress = publicNetAddress.String()

		err = db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte("lighthouses"))
			if err != nil {
				return err
			}
			err = b.Put([]byte(privateAddress), []byte(publicAddress))
			if err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			c.JSON(500, gin.H{"error": "internal error"})
			log.Error(err)
			return
		}

		c.JSON(200, gin.H{"publicAddress": publicAddress, "privateAddress": privateAddress})
	})

	type LighthouseMapping struct {
		PublicAddress  string `json:"publicAddress"`
		PrivateAddress string `json:"privateAddress"`
	}

	r.GET("/lighthouses", func(c *gin.Context) {
		var lighthouses []LighthouseMapping
		err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("lighthouses"))
			if b == nil {
				return nil
			}
			err = b.ForEach(func(k, v []byte) error {
				lighthouses = append(lighthouses, LighthouseMapping{PublicAddress: string(v), PrivateAddress: string(k)})
				return nil
			})
			return err
		})
		if err != nil {
			c.JSON(500, gin.H{"error": "internal error"})
			log.Error(err)
			return
		}
		c.JSON(200, gin.H{"lighthouses": lighthouses})
	})

	r.DELETE("/lighthouses", func(c *gin.Context) {
		privateAddress := c.Query("privateAddress")
		privateNetAddress := net.ParseIP(privateAddress)
		if privateNetAddress == nil {
			c.JSON(400, gin.H{"error": "address is invalid"})
			return
		}

		err = db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("lighthouses"))
			if b == nil {
				return nil
			}
			err = b.Delete([]byte(privateNetAddress.String()))
			if err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			c.JSON(500, gin.H{"error": "internal error"})
			log.Error(err)
			return
		}

		c.JSON(200, gin.H{"privateAddress": privateAddress})
	})

	type FirewallRule struct {
		Port   string   `json:"port"`
		Proto  string   `json:"proto"`
		Groups []string `json:"groups"`
	}

	r.GET("/lighthouses/firewall", func(c *gin.Context) {
		privateAddress := c.Query("privateAddress")
		privateNetAddress := net.ParseIP(privateAddress)
		if privateNetAddress == nil {
			c.JSON(400, gin.H{"error": "address is invalid"})
			return
		}

		privateAddress = privateNetAddress.String()

		firewallTable := map[string][]FirewallRule{}

		err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("lighthouses"))
			if b == nil {
				return nil
			}
			firewallTableJson := b.Get([]byte(privateAddress + "_firewall"))
			if firewallTableJson != nil {
				err = json.Unmarshal(firewallTableJson, &firewallTable)
				if err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			c.JSON(500, gin.H{"error": "internal error"})
			log.Error(err)
			return
		}

		c.JSON(200, firewallTable)
	})

	r.POST("/lighthouses/firewall", func(c *gin.Context) {
		privateAddress := c.Query("privateAddress")
		privateNetAddress := net.ParseIP(privateAddress)

		port := c.Query("port")
		proto := c.Query("proto")
		groups := c.QueryArray("groups")

		if port == "" || proto == "" || len(groups) == 0 || privateNetAddress == nil {
			c.JSON(400, gin.H{"error": "invalid parameters"})
			return
		}

		if privateNetAddress == nil {
			c.JSON(400, gin.H{"error": "address is invalid"})
			return
		}

		privateAddress = privateNetAddress.String()

		firewallRule := FirewallRule{Port: port, Proto: proto, Groups: groups}

		if err != nil {
			c.JSON(500, gin.H{"error": "internal error"})
			log.Error(err)
			return
		}

		firewallTable := map[string][]FirewallRule{}

		err = db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte("lighthouses"))
			if err != nil {
				return err
			}
			firewallTableJson := b.Get([]byte(privateAddress))
			if firewallTableJson != nil {
				err = json.Unmarshal(firewallTableJson, &firewallTable)
				if err != nil {
					return err
				}
				firewallTable[port] = append(firewallTable[port], firewallRule)
				if err != nil {
					return err
				}
			} else {
				firewallTable := map[string][]FirewallRule{port: {firewallRule}}
				firewallTableJson, err = json.Marshal(firewallTable)
				if err != nil {
					return err
				}
			}
			err = b.Put([]byte(privateAddress+"_firewall"), firewallTableJson)
			c.JSON(200, firewallTable)
			if err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			c.JSON(500, gin.H{"error": "internal error"})
			log.Error(err)
			return
		}

		c.JSON(200, gin.H{"rule": firewallTable})
	})

	r.Run(":8080")
}
