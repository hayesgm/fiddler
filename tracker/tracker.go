package tracker

import (
  "time"
  "github.com/coreos/go-etcd/etcd"
  "github.com/coreos/etcd/store"
  "fmt"
  "log"
)

// TrackStats is going to pull all pertinent system stats and update our state in etcd
func TrackMyStats(cli *etcd.Client, myid string, stats []string) {
  for {
    for _, stat := range stats {
      // Get the value for the stat and set cli key for the value
      stat := NewStat(stat)
      stat.write(myid, cli)
    }

    time.Sleep(3*time.Second)
  }
}

func WatchStats(cli *etcd.Client, myid string) {
  // This is going to be mutually locked against all nodes
  log.Println("Hello, I am:",myid)
  
  for {
    resp, acq, err := cli.TestAndSet("/fiddler/watcher", "", myid, 30)
    log.Println("Lock Resp:",acq,resp,err)
    if !acq { // We are locked out
      // We want to watch for a change in the lock, and we'll repeat
      var watcherCh = make(chan *store.Response)
      var endCh = make(chan bool)

      go cli.Watch("/fiddler/watcher", 0, watcherCh, endCh)
      <- watcherCh

      // Now, we'll try to acquire the lock, again
    } else {
      var endCh = make(chan bool)

      // We got a lock, we want to keep it
      go func() {
        for {
          resp, acq, err := cli.TestAndSet("/fiddler/watcher", myid, myid, 30) // Keep the lock alive
          log.Println("Reset Resp:",acq,resp,err)
          if !acq {
            <- endCh // Let's boot ourselves, we're no longer the leader
          }
          
          time.Sleep(15*time.Second)
        }
      }()

      // Let's run the watch code
      var statsCh = make(chan *store.Response)

      // We're going to see when thresholds are passed
      // Based on the configuration (/defaults)

      go cli.Watch("/stats", 0, statsCh, endCh)

      for {
        fmt.Printf("Watching...\n")
        <- statsCh
        //response := <- statsCh
        // fmt.Printf("Response: %#v\n", response)
      }
    }
  }
}