package logic

import (
	"net/http"
	"time"

	"github.com/hash-rabbit/auto-build/model"
)

type CommonInfo struct {
	TodayCount      int      `json:"today_count"`
	YesterdayCount  int      `json:"yesterday_count"`
	MonthCount      int      `json:"month_count"`
	LastMontCount   int      `json:"last_month_count"`
	MonthCountDate  []string `json:"month_count_date"`
	MonthCountGraph []int    `json:"month_count_graph"`
}

func HomeInfo(w http.ResponseWriter, r *http.Request) {
	info := &CommonInfo{
		MonthCountDate:  make([]string, 0),
		MonthCountGraph: make([]int, 0),
	}

	cur := time.Now()
	today := time.Date(cur.Year(), cur.Month(), cur.Day(), 0, 0, 0, 0, time.Local)
	yesterday := today.AddDate(0, 0, -1)

	count, err := model.CountTaskLog(today, cur)
	if err != nil {
		writeError(w, "logic error", err.Error())
		return
	}
	info.TodayCount = int(count)

	count, err = model.CountTaskLog(yesterday, today)
	if err != nil {
		writeError(w, "logic error", err.Error())
		return
	}
	info.YesterdayCount = int(count)

	tomonth := time.Date(cur.Year(), cur.Month(), 1, 0, 0, 0, 0, time.Local)
	lastMonth := tomonth.AddDate(0, -1, 0)
	count, err = model.CountTaskLog(tomonth, cur)
	if err != nil {
		writeError(w, "logic error", err.Error())
		return
	}

	info.MonthCount = int(count)
	count, err = model.CountTaskLog(lastMonth, tomonth)
	if err != nil {
		writeError(w, "logic error", err.Error())
		return
	}
	info.LastMontCount = int(count)

	ds, err := model.Count30DayTaskLog()
	if err != nil {
		writeError(w, "logic error", err.Error())
		return
	}

	dateMap := make(map[string]int)
	for _, v := range ds {
		dateMap[v.Date] = v.Count
	}

	starttemp := time.Now().AddDate(0, 0, -30)
	start := time.Date(starttemp.Year(), starttemp.Month(), starttemp.Day(), 0, 0, 0, 0, time.Local)

	for time.Now().After(start) {
		info.MonthCountDate = append(info.MonthCountDate, start.Format("2006-01-02"))
		info.MonthCountGraph = append(info.MonthCountGraph, dateMap[start.Format("2006-01-02")])
		start = start.AddDate(0, 0, 1)
	}

	writeJson(w, info)
}
