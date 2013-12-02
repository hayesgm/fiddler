package main

import (
  "github.com/hayesgm/fiddler/config"
  "github.com/hayesgm/fiddler/tracker"
  "github.com/hayesgm/fiddler/installer"
  "github.com/hayesgm/fiddler/launcher"
  "github.com/hayesgm/fiddler/spawner"
  "github.com/hayesgm/fiddler/orchestrator"
  "github.com/hayesgm/crates/client"
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

var cli *etcd.Client
var myid = uuid.NewV1().String()

func printUsage() {
  fmt.Println("Fiddler is an app to auto-scale containers.")
  fmt.Println("")
  fmt.Println("Usage:")
  fmt.Println("fiddler --config=<conf> install")
  fmt.Println("\tInstalls Fiddler into systemd")
  fmt.Println("fiddler --config=<conf> launch")
  fmt.Println("\tLaunches a new container within Fiddler")
  fmt.Println("fiddler --config=<conf> spawn")
  fmt.Println("\tSpawns a new fiddler environment in the cloud")
  fmt.Println("fiddler --config=<conf> daemon")
  fmt.Println("\tRuns Fiddler like a daemon to monitor machine")
  fmt.Println("fiddler --config=<conf> --cloud=<cloud> register")
  fmt.Println("\tRuns Fiddler to register a machine to a Lease Cloud")
  fmt.Println("fiddler --config=<conf> join")
  fmt.Println("\tJoins a Fiddler ring")
}

func runLaunch(conf *config.FiddlerConf) {
  // We should launch our container
  var cmd *exec.Cmd
  cmd, err := launcher.Launch(conf.Docker)
  if err != nil {
    log.Fatal("Error launching container:",conf.Docker,err)
  }

  err = cmd.Wait() // We'll stay open so long as the docker is open.  We can ensure the docker stays open, etc.
  if err != nil {
    log.Fatal("Failed to run container:",conf.Docker,err)
  }
}

func joinRing(myid string, conf *config.FiddlerConf) {
  // This is going to ask Orchestrator for roles, and fulfill them as required
  rolesCh := orchestrator.MyRoles(conf.Docker, myid, cli)

  for {
    roles := <- rolesCh
    
    // Well, should this kind of be, kill or launch?
    for _, role := range roles {
      // We should launch our container
      var cmd *exec.Cmd
      cmd, err := launcher.Launch(role.Docker)
      if err != nil {
        log.Fatal("Error launching container:",role.Docker,err)
      }
      
      err = cmd.Wait() // We'll stay open so long as the docker is open.  We can ensure the docker stays open, etc.
      if err != nil {
        log.Fatal("Failed to run container:",role.Docker,err)
      }
    }
  }
}

func loadConfig(c *string) (conf *config.FiddlerConf) {
  if *c == "" {
    printUsage()
    log.Fatal("Missing config option")
  }

  // Next, we'll load the config
  conf, err := config.LoadFiddlerConfig(*c)

  if err != nil {
    log.Fatal("Unable to load config file:", err)
  }

  log.Printf("Connecting to etcd at %#v...\n", conf.Ring)
  cli = etcd.NewClient(conf.Ring)

  return
}

func main() {
  // First, let's make sure we have a configuration file
  var c = flag.String("config", "", "location of configuration file (http okay)")
  var cloud = flag.String("cloud", "", "crate cloud to connect to")
  flag.Parse()

  if len(flag.Args()) < 1 {
    printUsage()
    return
  }

  var cmd = flag.Args()[0]

  switch cmd {
  case "i", "install":
    conf := loadConfig(c)

    // We should install Fiddler
    err := installer.InstallFiddler(*c, *conf)
    if err != nil {
      log.Fatal("Error installing Fiddler:", err)
    }
  case "s", "spawn":
    conf := loadConfig(c)

    if len(conf.Env) == 0 {
      fmt.Println("Must provide env name in config")
      return
    }

    // We'll grab spawn pool and grow it
    pool, err := spawner.GetSpawnPool(conf.Env)
    if err != nil {
      log.Fatal("Error getting spawn pool:", err)
    }

    err = pool.Grow(*c)
    if err != nil {
      log.Fatal("Error growing spawn pool:", err)
    }
  case "l", "launch":
    conf := loadConfig(c)

    runLaunch(conf)
  case "j", "join":
    conf := loadConfig(c)

    joinRing(myid, conf)
  case "d", "daemon":
    conf := loadConfig(c)

    // We should be like a daemon, tracking stats

    // Now, we're going to make sure we're monitoring our stats
    go tracker.TrackMyStats(cli, myid, []string{"cpu"})
    go tracker.WatchStats(cli, myid, conf)

    ch := make(chan int)
    <- ch // Hold forever
  case "r", "register":
    // We should regiser to a Crate cload
    regResp, err := client.Register(*cloud)

    if err != nil {
      log.Fatal("Failed to register crate:", err)
    }

    log.Printf("Registered as %s\n", regResp.Crate.Id)

    acqResp, err := client.Acquire(*cloud, regResp.Crate.Id.Hex())
    if err != nil {
      log.Fatal("Failed to acquire lease:", err)
    }

    log.Printf("Acquired lease: %#v\n", acqResp)

    conf := loadConfig(&acqResp.Lease.Conf)
    
    runLaunch(conf)
  default:
    printUsage()
    return
  }
}