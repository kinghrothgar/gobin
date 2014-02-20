package store

import (
	"github.com/kinghrothgar/goblin/storage"
	"sort"
	"time"
)

// A data structure to hold a key/value pair.
type UIDTime struct {
	UID  string
	Time time.Time
}

// A slice of Pairs that implements sort.Interface to sort by Value.
type UIDTimeList []UIDTime

func (p UIDTimeList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p UIDTimeList) Len() int      { return len(p) }
func (p UIDTimeList) Less(i, j int) bool {
	if p[i].Time.Equal(p[j].Time) {
		return p[i].UID < p[j].UID
	}
	return p[i].Time.Before(p[j].Time)
}

// A function to turn a map into a PairList, then sort and return it.
func sortHordeByTime(h storage.Horde) UIDTimeList {
	l := make(UIDTimeList, len(h))
	i := 0
	for uid, created := range h {
		l[i] = UIDTime{UID: uid, Time: created}
		i++
	}
	sort.Sort(l)
	return l
}
