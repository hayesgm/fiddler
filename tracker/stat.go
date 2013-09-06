package tracker

import (
  "github.com/coreos/go-etcd/etcd"
  "strconv"
  "path"
  "bitbucket.org/hayesgm/systemstat"
  "math/rand"
  //"log"
)

type Stat struct {
  StatType string
  Value float64
}

func NewStat(statType string) (stat *Stat) {
  stat = &Stat{StatType: statType}
  return
}

func (stat *Stat) GetStatValue() float64 {
  defer func() {
    if e := recover(); e != nil {
      // We're going to squash issues pull stats
      //log.Println("Failed to pull `", stat.StatType, "` stat data:", e)
    }
  }()
  switch stat.StatType {
  case "cpu":
    // one := systemstat.GetCPUSample()
    // two := systemstat.GetCPUSample()
    // avg := systemstat.GetSimpleCPUAverage(one, two)
    // return avg.BusyPct
    return rand.Float64()
  case "free-mem":
    return float64(systemstat.GetMemSample().MemFree)
  default:
    return 0.0
  }
}

func (stat *Stat) write(myid string, cli *etcd.Client) {
  stat.Value = stat.GetStatValue()

  value := strconv.FormatFloat(stat.Value, 'f', -1, 64)
  cli.Set(path.Join("fiddler/stats",stat.StatType,myid), value, 60)
}