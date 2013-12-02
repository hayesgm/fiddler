package orchestrator

import (
  "github.com/hayesgm/fiddler/config"
  "github.com/coreos/go-etcd/etcd"
  "github.com/hayesgm/go-etcd-lock/daemon"
  "encoding/json"
  "log"
  "path"
  "time"
)

type Role struct {
  Docker *config.DockerConf
  // Pipes []piper.Pipe
}

/*
  Orchestrator is going to take a cloud configuration
  and a cluster (generally, etcd bundle).  It will
  assign each server roles which it must take on.
  Orchestrator will, now-and-then, adjust roles
  for servers.

  Additionally, Orchestrator will caller
  Piper to create sockets for communication
  between the different nodes.
*/

/* If we're the leader [and we get to decide the world], 
  this function will come up with a view of what the
  world of servers should look like.
  This is going to be built around a cost-minization
  function that will have penalties for changing
  roles on servers. */
func decideWorld(conf *config.DockerConf, myid string, cli *etcd.Client) (world map[string][]Role) {
  world = make(map[string][]Role) // Make a map of the world

  // We're going to set everything to ourselves, for now
  roles := make([]Role, 1)
  roles[0].Docker = conf
  
  world[myid] = roles

  return
}

/* Attempts to rule the state of the world by
  attaining the `fiddler/ruler` lock.

  The ruler will run and dictate world state
  with decideWorld().
*/
func tryToRuleTheWorld(conf *config.DockerConf, myid string, cli *etcd.Client) {
  rule := func(stopCh chan int) {
    for {
      world := decideWorld(conf, myid, cli)

      for id, roles := range world {

        str, err := json.Marshal(roles)
        if err != nil {
          log.Fatal(err)
        }
        log.Printf("Setting Roles for %v to %v", id, string(str))
        cli.Set(path.Join("fiddler/roles", id), string(str), 0)  
      }
      
      time.Sleep(3000 * time.Millisecond)
    }
  }

  daemon.RunOne(cli, "fiddler/ruler", rule, 10)
}

/*
  This will return a channel which dicates which Roles this
  Fiddler should launch.  It is the responsibility of
  each Fiddler to keep its launches limited to Roles
  returns on this channel.
*/
func MyRoles(conf *config.DockerConf, myid string, cli *etcd.Client) (rolesCh chan []Role) {
  rolesCh = make(chan []Role)
  var watcherCh = make(chan *etcd.Response)
  var endCh = make(chan bool) // Do we use this?

  go cli.Watch(path.Join("fiddler/roles", myid), 0, false, watcherCh, endCh)

  // We'll start another goroutine which will parse the response and return roles
  go func() {
    var roles []Role

    for {
      response := <- watcherCh

      if err := json.Unmarshal([]byte(response.Value), &roles); err != nil {
        log.Fatalf("json.Unmarshal failed: %v", err)
      }

      rolesCh <- roles // print our current roles
    }
  }()

  go tryToRuleTheWorld(conf, myid, cli)

  return
}

