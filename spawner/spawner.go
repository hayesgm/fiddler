package spawner

// TODO: Make platform agnostic
// TODO: Take keys not from ENV?
// TODO: Find a way to get Fiddler script to run
// Are we going to have to run an SSH to run the install?

import (
  "fmt"
  "os"
  "errors"
  "io/ioutil"
)

type SpawnPool interface {
  // Grows the pool by one instance
  Grow(conf string) (err error)
  // Shrinsk the pool one instance
  Shrink() (err error)
}

// GetSpawnPool will return spawn pool from the environment or etcd
func GetSpawnPool(name string) (pool SpawnPool, err error) {
  // First, we'll try to grab the environment from env
  if len(os.Getenv("cloud")) == 0 {
    return nil, errors.New("Must provide cloud env variable to run spawner")
  }

  // We have a cloud environment variable
  switch os.Getenv("cloud") {
  case "aws":
    var key, secret, zone string
    key, secret, zone = os.Getenv("key"), os.Getenv("secret"), os.Getenv("zone")
    if len(key) == 0 || len(secret) == 0 || len(zone) == 0 {
      return nil, errors.New("Must provide key, secret and zone to spawn AWS instances")
    }
    // There's hacks all around for this
    privateKeyFile := os.Getenv("privateKeyFile")
    if len(privateKeyFile) == 0 {
      return nil, errors.New("Must provide privateKey")
    }
    
    privateKey, err := ioutil.ReadFile(privateKeyFile)
    if err != nil {
      return nil, err
    }

    return NewAmazonSpawnPool(name, key, secret, zone, string(privateKey))
  default:
    return nil, errors.New(fmt.Sprintf("Unknown cloud: %s", os.Getenv("cloud")))
  }
}