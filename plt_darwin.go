//-----------------------------------------------------------------------------
// Copyright (C) Microsoft. All rights reserved.
// Licensed under the MIT license.
// See LICENSE.txt file in the project root for full license information.
//-----------------------------------------------------------------------------
package main

import (
	"bufio"
	"context"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	tm "github.com/nsf/termbox-go"
)

type ethrNetDevInfo struct {
	bytes      uint64
	packets    uint64
	drop       uint64
	errs       uint64
	fifo       uint64
	frame      uint64
	compressed uint64
	multicast  uint64
}

func getNetDevStats(stats *ethrNetStat) {
	ifs, err := net.Interfaces()
	if err != nil {
		ui.printErr("%v", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, "netstat", "-w1")

	rc, err := cmd.StdoutPipe()
	if err != nil {
		ui.printErr("%v", err)
		return
	}

	reader := bufio.NewReader(rc)

	go func() { // Run command in another goroutine
		if err := cmd.Run(); err != nil {
			ui.printErr("%v", err)
			return
		}
	}()

	if _, err = reader.ReadString('\n'); err != nil {
		ui.printErr("%v", err)
		return
	}

	if _, err = reader.ReadString('\n'); err != nil {
		ui.printErr("%v", err)
		return
	}

	var line string
	for err == nil {
		line, err = reader.ReadString('\n')
		if line == "" {
			continue
		}
		// input        (Total)           output
		// packets  errs      bytes    packets  errs      bytes colls
		// 0     0          0          0     0          0     0
		// 1     0         66          1     0         66     0
		if strings.Contains(line, "input") || strings.Contains(line, "packets") {
			continue
		}

		netDevStat := buildNetDevStat("", line)
		if isIfUp(netDevStat.interfaceName, ifs) {
			stats.netDevStats = append(stats.netDevStats, buildNetDevStat("", line))
		}
	}
}

func buildNetDevStat(interfaceName, line string) ethrNetDevStat {
	fields := strings.Fields(line)
	// spew.Dump(fields[:3])
	// spew.Dump(fields[3:])
	rxInfo := toNetDevInfo(fields[:3])
	txInfo := toNetDevInfo(fields[3:])
	return ethrNetDevStat{
		interfaceName: interfaceName,
		rxBytes:       rxInfo.bytes,
		txBytes:       txInfo.bytes,
		rxPkts:        rxInfo.packets,
		txPkts:        txInfo.packets,
	}
}

func toNetDevInfo(fields []string) ethrNetDevInfo {
	return ethrNetDevInfo{
		packets: toInt(fields[0]),
		errs:    toInt(fields[1]),
		bytes:   toInt(fields[2]),
		// drop:       toInt(fields[3]),
		// fifo:       toInt(fields[4]),
		// frame:      toInt(fields[5]),
		// compressed: toInt(fields[6]),
		// multicast:  toInt(fields[7]),
	}
}

func isIfUp(ifName string, ifs []net.Interface) bool {
	for _, ifi := range ifs {
		if ifi.Name == ifName {
			if (ifi.Flags & net.FlagUp) != 0 {
				return true
			}
			return false
		}
	}
	return false
}

func toInt(str string) uint64 {
	res, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		panic(err)
	}
	return res
}

func getTcpStats(stats *ethrNetStat) {
	var err error
	for err != nil {
		time.Sleep(time.Second)
	}

	// snmpStatsFile, err := os.Open("/proc/net/snmp")
	// if err != nil {
	// 	ui.printErr("%v", err)
	// 	return
	// }
	// defer snmpStatsFile.Close()
	//
	// reader := bufio.NewReader(snmpStatsFile)
	//
	// var line string
	// for err == nil {
	// 	// Tcp: RtoAlgorithm RtoMin RtoMax MaxConn ActiveOpens PassiveOpens AttemptFails EstabResets
	// 	//      CurrEstab InSegs OutSegs RetransSegs InErrs OutRsts InCsumErrors
	// 	line, err = reader.ReadString('\n')
	// 	if line == "" || !strings.HasPrefix(line, "Tcp") {
	// 		continue
	// 	}
	// 	// Skip the first line starting with Tcp
	// 	line, err = reader.ReadString('\n')
	// 	if !strings.HasPrefix(line, "Tcp") {
	// 		break
	// 	}
	// 	fields := strings.Fields(line)
	// 	stats.tcpStats.segRetrans = toInt(fields[12])
	// }
}

func hideCursor() {
	tm.SetCursor(0, 0)
}

func blockWindowResize() {
}
