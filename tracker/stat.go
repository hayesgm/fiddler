package tracker

import (
  "github.com/coreos/go-etcd/etcd"
  "strconv"
  "path"
  "math/rand"
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
  switch stat.StatType {
  case "cpu":
    return rand.Float64()
  default:
    return 0.0
  }
}

func (stat *Stat) write(myid string, cli *etcd.Client) {
  stat.Value = stat.GetStatValue()

  value := strconv.FormatFloat(stat.Value, 'f', -1, 64)
  cli.Set(path.Join("/stats",myid,stat.StatType), value, 60)
}