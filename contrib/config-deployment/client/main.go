package main

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	flag "github.com/spf13/pflag"
)

var ( // CLI flags
	host         string
	port         int
	clientID     string
	pollInterval time.Duration
)

var done atomic.Bool

func init() {
	flag.StringVarP(&host, "host", "h", "localhost", "wfx host")
	flag.IntVarP(&port, "port", "p", 8080, "wfx port")
	flag.StringVarP(&clientID, "client-id", "c", "", "client id to use")
	flag.DurationVarP(&pollInterval, "interval", "i", time.Second*10, "polling interval")
}

func main() {
	flag.Parse()
	if clientID == "" {
		fmt.Fprintf(os.Stderr, "FATAL: Argument --client-id is missing\n")
		os.Exit(1)
	}

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	log.Println("Press CTRL+C to quit...")

	go worker()

	<-signalChannel
	done.Store(true)
	log.Println("Bye bye")
}
