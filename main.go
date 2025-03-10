package main

import (
	"flag"
	"fmt"
	"os"

	"encoding/json"

	wrapping "github.com/hashicorp/go-kms-wrapping"
	log "github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"

	"github.com/ruizink/vault-kms-keyid-updater/version"
)

func main() {

	dbFile := flag.String("db", "", "Path to the BoltDB file")
	boltBucket := flag.String("bucket", "", "Name of the bucket")
	boltKey := flag.String("boltkey", "", "Bolt key to look for")
	keyID := flag.String("keyid", "", "KeyID to look for")
	newKeyID := flag.String("newkeyid", "", "New KeyID to set")
	logLevel := flag.String("loglevel", "info", "Log level (debug, info, warn, error, fatal, panic)")
	version := flag.Bool("version", false, "Print the version and exit")

	flag.Parse()

	if *version {
		printVersion()
		os.Exit(0)
	}

	// Configure logging
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{})
	level, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Fatalf("Error parsing log level: %s", err)
	}
	log.SetLevel(level)

	if *dbFile == "" || *boltBucket == "" || *boltKey == "" || *keyID == "" || *newKeyID == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	log.Infof("Loading database %s", *dbFile)

	// Check if the database file exists
	if _, err := os.Stat(*dbFile); err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("BoltDB database file '%s' does not exist", *dbFile)
		} else {
			log.Fatal(err)
		}
	}

	db, err := bolt.Open(*dbFile, 0600, nil)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	defer db.Close()

	blob, err := readBoltKey(db, *boltBucket, *boltKey)
	if err != nil {
		log.Fatalf("Error reading key: %s", err)
	}

	log.Debugf("Key value (hex): %#X", blob)

	blobInfo := &wrapping.EncryptedBlobInfo{}

	if err := proto.Unmarshal(blob, blobInfo); err != nil {
		log.Fatalf("Failed to proto decode blob: %s", err)
	}

	blobStr := prettyPrint(blobInfo)
	log.Debugf("blobInfo=%s", blobStr)

	if blobInfo.KeyInfo.KeyID != *keyID {
		log.Infof("Current KeyID '%s' does not match the provided one '%s'. Skipping...", blobInfo.KeyInfo.KeyID, *keyID)
		return
	}

	log.Infof("Changing KeyID from '%s' to '%s'", *keyID, *newKeyID)
	blobInfo.KeyInfo.KeyID = *newKeyID

	blobInfoBytes, err := proto.Marshal(blobInfo)
	if err != nil {
		log.Fatalf("Failed to proto encode blob: %s", err)
	}

	if err := writeBoltKey(db, *boltBucket, *boltKey, blobInfoBytes); err != nil {
		log.Fatalf("Error writing key: %s", err)
	}

	log.Debugf("New key value (hex): %#X", blobInfoBytes)

	log.Debugf("blobInfo=%s", prettyPrint(blobInfo))

}

func writeBoltKey(db *bolt.DB, bucket string, key string, value []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		return b.Put([]byte(key), value)
	})
}

func readBoltKey(db *bolt.DB, bucket string, key string) ([]byte, error) {
	var value []byte
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("Bucket '%s' not found", bucket)
		}
		value = b.Get([]byte(key))
		if value == nil {
			log.Warnf("Key '%s' not found or empty", key)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return value, nil
}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

func printVersion() {
	fmt.Fprintf(os.Stderr, "Version: %s\n", version.Version)
	fmt.Fprintf(os.Stderr, "(Build date: %s, Git commit: %s)\n", version.BuildDate, version.GitCommit)
}
