package tracker

import (
  "time"
  "github.com/coreos/go-etcd/etcd"
  "github.com/coreos/etcd/store"
  "fmt"
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

func WatchStats(cli *etcd.Client) {
  var statsCh = make(chan *store.Response)
  var endCh = make(chan bool)
  go cli.Watch("/stats", 0, statsCh, endCh)

  for {
    fmt.Printf("Watching...\n")
    response := <- statsCh
    fmt.Printf("Response: %#v\n", response)

    // From here, if we're good, we can keep a map of all of the stats to make that easily accessible
    // But if we're not, we could really just poll etcd to figure out the stats
    
    // If our process breaks a given threshold, we can call ramp-up or ramp-down
  }
}