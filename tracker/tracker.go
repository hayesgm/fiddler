package tracker

import (
  "fmt"
  "time"
  "github.com/coreos/go-etcd/etcd"
  "log"
  "github.com/hayesgm/fiddler/lock"
  "github.com/hayesgm/fiddler/config"
  "strconv"
  "strings"
  "errors"
  "path"
)

// TrackStats is going to pull all pertinent system stats and update our state in etcd
func TrackMyStats(cli *etcd.Client, myid string, stats []string) {
  for {
    cli.Set(path.Join("fiddler/servers",myid), "present", 60)

    for _, stat := range stats {
      // Get the value for the stat and set cli key for the value
      stat := NewStat(stat)
      stat.write(myid, cli)
    }

    time.Sleep(3*time.Second)
  }
}

func getStats(cli *etcd.Client, metric string) (values []float64, err error) {
  serverResp, err := cli.Get(path.Join("fiddler/stats",metric))
  if err != nil {
    return nil, err
  }
  
  values = make([]float64, len(serverResp))
  for i, resp := range serverResp {
    if values[i], err = strconv.ParseFloat(resp.Value, 64); err != nil {
      return nil, err
    }
  }
  log.Printf("Metric `%s` values: %v\n", metric, values)

  return
}

func aggregate(agg string, values []float64) (value float64, err error) {
  switch agg {
  case "avg":
    sum := 0.0
    for _, v := range values {
      sum += v
    }
    return sum/float64(len(values)), nil
  default:
    return 0, errors.New(fmt.Sprintf("Unknown agg: %s", agg))
  }
}

func compare(a float64, comp byte, b float64) (bool, error) {
  switch comp {
  case '<':
    return a < b, nil
  case '>':
    return a > b, nil
  case '=':
    return a == b, nil
  default:
    return false, errors.New(fmt.Sprintf("Unknown comp: %s", comp))
  }
}

func check(cli *etcd.Client, stat, val string) (pass bool, err error) {
  // We're going to do a quick job parsing stat and val
  statItems := strings.Split(stat, "-")
  if len(statItems) != 2 {
    return false, errors.New(fmt.Sprintf("stat would be <agg>-<stat>, found: %s",stat))
  }
  agg, metric := statItems[0], statItems[1]

  comp := []byte(val)[0]
  amt, err := strconv.ParseFloat(string([]byte(val)[1:]),64)
  if err != nil {
    return false, err
  }

  values, err := getStats(cli, metric)
  if err != nil {
    return false, err
  }

  value, err := aggregate(agg, values)
  if err != nil {
    return false, err
  }

  pass, err = compare(value, comp, amt)
  if err != nil {
    return false, err
  }

  log.Printf("Checked (%v) `%s` of `%s` = %v ?%s %v", pass, agg, metric, value, string(comp), amt)

  return
}

func checkStats(cli *etcd.Client, conf *config.FiddlerConf) (err error) {
  // Pull aggregates of the information from config
  serverResp, err := cli.Get("fiddler/servers")
  if err != nil {
    return err
  }

  serverCount := len(serverResp)
  grow, shrink := false, false // We'll track both

  if serverCount < conf.Scale.Max {
    // We might want to Grow

    // Let's see if we fit any of the parameters
    for stat, val := range conf.Scale.Grow {
      if pass, err := check(cli, stat, val); err != nil {
        return err
      } else if pass {
        grow = true
      }
    }
  }

  if serverCount > conf.Scale.Min {
    // We might want to Shrink

    // Let's see if we fit any of the parameters
    for stat, val := range conf.Scale.Shrink {
      if pass, err := check(cli, stat, val); err != nil {
        return err
      } else if pass {
        shrink = true
      }
    }
  }

  if grow && shrink {
    log.Println("Fail as we want to grow and shrink.")
  } else if grow {
    log.Println("I want to grow")
  } else if shrink {
    log.Println("I want to shrink")
  }

  return
}
func WatchStats(cli *etcd.Client, myid string, conf *config.FiddlerConf) {
  goChan, stopChan := lock.Acquire(cli, "fiddler/watcher", 20)

  go func() {
    <- goChan
    
    run := true

    for run {
      select {
      case <-stopChan:
        run = false // We're going to exit
      default:
        log.Println("I am king")

        // We're going to look at the stats we want to look at
        // and determine the correct count of servers

        err := checkStats(cli, conf)
        if err != nil {
          log.Printf("Encountered error: %s", err)
        }

        time.Sleep(5*time.Second)

        /*
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
        */
      }
    }
  }()
/*
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

      
    }
  }
*/
}
