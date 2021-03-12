package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "github.com/papertrail/go-tail/follower"
    "io"
    "log"
    "strconv"
    "strings"
    "time"
)

type logSourceInfo struct  {
    loggedCount             int64
    collectedCount          int64

    firstSeq                int64
}

type reportStatistics struct  {
    totalLogsCollectedCount  int64
}

func main() {

    var fluentLogFileName string
    var reportCount int64

    flag.StringVar(&fluentLogFileName, "f", "0.log", "fluent log file to tail")
    flag.Int64Var(&reportCount, "c", 100, "number of logs between reports")
    flag.Parse()

    logSourceInfoMap := make(map[string]logSourceInfo)
    reportData := reportStatistics{}

    t, err := follower.New(fluentLogFileName , follower.Config{
        Whence: io.SeekStart,
        Offset: 0,
        Reopen: true,
    })

    if err != nil {
        log.Fatal(err)
    }

    for line := range t.Lines() {
        lineSplit := strings.Split(line.String(), "{")
        if len(lineSplit) < 2 {
            continue
        }
        jsonString := "{"+strings.Split(lineSplit[1],"}")[0]+"}"
        var j map[string]interface{}
        if err1 := json.Unmarshal([]byte(jsonString), &j); err1 != nil {
            continue
        }

        if _, ok := j["path"]; !ok {
            continue
        }

        if _, ok := j["log"]; !ok {
            continue
        }

        // get the file name of log (container name)
        nameSliced := strings.Split(fmt.Sprintf("%s", j["path"]), "_")
        if len(nameSliced) < 1 {
            continue
        }
        nameSliced = strings.Split(fmt.Sprintf("%s", nameSliced[0]), "/")
        if len(nameSliced) < 4 {
            continue
        }

        name := nameSliced[4]
        if _, ok := logSourceInfoMap[name]; !ok {
            logSourceInfoMap[name] = logSourceInfo{}
        }

        // get the sequence number of the log
        logSliced := strings.Split(fmt.Sprintf("%s", j["log"]), "-")
        if len(logSliced) < 2 {
            continue
        }
        seqStr := strings.TrimSpace(logSliced[2])
        seq, err1 := strconv.ParseInt(seqStr, 10, 0)
        if err1 != nil{
            continue
        }

        // calculate metrics
        entry:= logSourceInfoMap[name]
        if entry.firstSeq == 0 {
            entry.firstSeq = seq-1
        }
        entry.collectedCount += 1
        entry.loggedCount = seq - entry.firstSeq

        // calculate global metrics
        reportData.totalLogsCollectedCount +=1

        // persist
        logSourceInfoMap[name] = entry

        if reportData.totalLogsCollectedCount % reportCount == 0 {
            report(reportData, logSourceInfoMap)
        }
    }
}

func report(reportData reportStatistics, logSourceInfoMap map[string]logSourceInfo) {
    log.Printf("Report at: %s\n", time.Now().String())
    log.Printf("-==-=-=-=-=\n")
    log.Printf("Total number of collected logs : %d\n", reportData.totalLogsCollectedCount)
    log.Printf("-==-=-=-=-=\n")
    tableFormat := "| %-40v | %-20v | %-20v | %-20v |\n"
    log.Printf(tableFormat,
        "Container name",
        "Lines Logged",
        "Lines Collected",
        "Lines Loss")
    log.Printf(strings.Repeat("-", len(fmt.Sprintf(tableFormat,0,0,0,0))-1) )
    for name, entry  := range logSourceInfoMap {
        log.Printf(tableFormat,name, entry.loggedCount, entry.collectedCount,entry.loggedCount- entry.collectedCount)

    }
    log.Printf("\n\n")
}
