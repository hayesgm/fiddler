package main

import (
  "github.com/hayesgm/fiddler/config"
  "github.com/hayesgm/fiddler/tracker"
  "github.com/hayesgm/fiddler/installer"
  "github.com/hayesgm/fiddler/launcher"
  "github.com/hayesgm/fiddler/spawner"
  "github.com/coreos/go-etcd/etcd"
  "github.com/go-contrib/uuid"
  "log"
  "flag"
  "os/exec"
  "fmt"
)

/*
  Fiddler is a daemon that will check stats amoungst a
  set of nodes in an etcd cluster.  If certain thresholds
  are exceeded across the stack, fiddle will issue
  auto-scale commands to expand the cluster.
*/

/*
  Fiddler knows how to spawn itself into CoreOS
  and start docker containers.  Thus, Fiddler is
  quickly able to coordinate replicating itself
  as load dictates.  It's also able to detect drops
  in load, where it will kill excess servers.
*/

/*
  Fiddler uses a Fidderfile for all configuration settings.
  ./fiddler -c fiddler.conf
*/

var cli = etcd.NewClient()
var myid = uuid.NewV1().String()

func printUsage() {
  fmt.Println("Fiddler is an app to auto-scale containers.")
  fmt.Println("")
  fmt.Println("Usage:")
  fmt.Println("fiddler --config=<conf> install")
  fmt.Println("\tInstalls Fiddler into systemd")
  fmt.Println("fiddler --config=<conf> launch")
  fmt.Println("\tLaunches a new container within Fiddler")
  fmt.Println("fiddler --config=<conf> spawn <name>")
  fmt.Println("\tSpawns a new fiddler environment in the cloud")
  fmt.Println("fiddler --config=<conf> daemon")
  fmt.Println("\tRuns Fiddler like a daemon to monitor machine")
}

func main() {
  // First, let's make sure we have a configuration file
  var c = flag.String("config", "", "location of configuration file (http okay)")
  flag.Parse()

  if len(flag.Args()) < 1 {
    printUsage()
    return
  }

  if *c == "" {
    fmt.Println("Missing config option")
    printUsage()
    return
  }

  // Next, we'll load the config
  var conf *config.FiddlerConf
  conf, err := config.LoadFiddlerConfig(*c)

  if err != nil {
    log.Fatal("Unable to load config file:", err)
  }

  var cmd = flag.Args()[0]

  switch cmd {
  case "i", "install":
    // We should install Fiddler
    err := installer.InstallFiddler(*c, *conf)
    if err != nil {
      log.Fatal("Error installing Fiddler:", err)
    }
  case "s", "spawn":
    if len(flag.Args()) < 2 {
      fmt.Println("Must provide env name")
      printUsage()
      return
    }

    name := flag.Args()[1]

    // We'll grab spawn pool and grow it
    pool, err := spawner.GetSpawnPool(name)
    if err != nil {
      log.Fatal("Error getting spawn pool:", err)
    }

    err = pool.Grow(*c)
    if err != nil {
      log.Fatal("Error growing spawn pool:", err)
    }
  case "l", "launch":
    // We should launch our container
    var cmd *exec.Cmd
    cmd, err = launcher.Launch(conf.Docker)
    if err != nil {
      log.Fatal("Error launching container:",conf.Docker,err)
    }

    err = cmd.Wait() // We'll stay open so long as the docker is open.  We can ensure the docker stays open, etc.
    if err != nil {
      log.Fatal("Failed to run container:",conf.Docker,err)
    }
  case "d", "daemon":
    // We should be like a daemon, tracking stats

    // Now, we're going to make sure we're monitoring our stats
    go tracker.TrackMyStats(cli, myid, []string{"cpu"})
    go tracker.WatchStats(cli, myid)

    ch := make(chan int)
    <- ch // Hold forever
  default:
    printUsage()
    return
  }
}