package main

import (
  "github.com/hayesgm/fiddler/config"
  "github.com/hayesgm/fiddler/tracker"
  "github.com/hayesgm/fiddler/launcher"
  "github.com/coreos/go-etcd/etcd"
  "github.com/go-contrib/uuid"
  "log"
  "flag"
  "os/exec"
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

func main() {
  // First, let's make sure we have a configuration file
  var install = flag.Bool("i", false, "install fiddler")
  var launch = flag.Bool("l", false, "launch docker directly")
  var c = flag.String("c", "", "location of configuration file (http okay)")
  flag.Parse()

  if *c == "" {
    log.Fatal("Must include configuration (-c) option")
  }

  // Next, we'll load the config
  var conf *config.FiddlerConf
  conf, err := config.LoadFiddlerConfig(*c)

  if err != nil {
    log.Fatal("Unable to load config file:", err)
  }

  if *install {
    err := launcher.InstallFiddler(*c, *conf, *launch)
    if err != nil {
      log.Fatal("Error installing Fiddler: ", err)
    }
  } else {
    // Now, we're going to make sure we're monitoring our stats
    go tracker.TrackMyStats(cli, myid, []string{"cpu"})
    go tracker.WatchStats(cli) // TODO: This should only be for leader node

    if *launch {
      // Next, we're going to spawn the desired Docker process
      var cmd *exec.Cmd
      cmd, err = launcher.Launch(conf.Docker)
      if err != nil {
        log.Fatal("Unable to launch",conf.Docker,err)
      }

      err = cmd.Wait() // We'll stay open so long as the docker is open.  We can ensure the docker stays open, etc.
      if err != nil {
        log.Fatal("Unable to complete",conf.Docker,err)
      }

      log.Println("Exiting...")
    }
  }
}