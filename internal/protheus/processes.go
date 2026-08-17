package protheus

import (
	"context"
	"strings"

	"github.com/shirou/gopsutil/v4/process"
)

type ProcessInfo struct {
	PID int32 `json:"pid"`
	Name string `json:"name"`
	CPU float64 `json:"cpu_percent"`
	MemoryMB float64 `json:"memory_mb"`
	CreateAge float64 `json:"uptime_hours"`
}

type ProcessList struct { ProcessPattern string `json:"process_pattern"`; Count int `json:"count"`; Processes []ProcessInfo `json:"processes"` }

func FindProcesses(ctx context.Context, pattern string) (ProcessList,error) {
	processes,err:=process.ProcessesWithContext(ctx);if err!=nil{return ProcessList{},err};pattern=strings.ToLower(strings.TrimSuffix(pattern,".exe"));result:=ProcessList{ProcessPattern:pattern,Processes:[]ProcessInfo{}}
	for _,p:=range processes{name,err:=p.NameWithContext(ctx);if err!=nil{continue};normalized:=strings.ToLower(strings.TrimSuffix(name,".exe"));if !strings.Contains(normalized,pattern){continue};memInfo,_:=p.MemoryInfoWithContext(ctx);cpuPercent,_:=p.CPUPercentWithContext(ctx);createTime,_:=p.CreateTimeWithContext(ctx);memoryMB:=0.0;if memInfo!=nil{memoryMB=float64(memInfo.RSS)/(1024*1024)};uptimeHours:=0.0;if createTime>0{nowMS,err:=processNowMS();if err==nil&&nowMS>createTime{uptimeHours=float64(nowMS-createTime)/1000/3600}};result.Processes=append(result.Processes,ProcessInfo{PID:p.Pid,Name:name,CPU:round(cpuPercent,2),MemoryMB:round(memoryMB,2),CreateAge:round(uptimeHours,2)})};result.Count=len(result.Processes);return result,nil
}
