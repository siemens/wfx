package main

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"log"
	"time"

	"github.com/Southclaws/fault"
	"github.com/cavaliergopher/grab/v3"
)

func download(url string, dir string) (string, error) {
	req, err := grab.NewRequest(dir, url)
	if err != nil {
		return "", fault.Wrap(err)
	}

	// start download
	log.Printf("Downloading %v...\n", req.URL())
	resp := grab.NewClient().Do(req)
	log.Printf("  %v\n", resp.HTTPResponse.Status)

	// start UI loop
	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()

Loop:
	for {
		select {
		case <-t.C:
			log.Printf("  transferred %v / %v bytes (%.2f%%)\n",
				resp.BytesComplete(),
				resp.Size(),
				100*resp.Progress())

		case <-resp.Done:
			// download is complete
			break Loop
		}
	}

	if err := resp.Err(); err != nil {
		return "", fault.Wrap(err)
	}

	log.Printf("Download saved to %s\n", resp.Filename)
	return resp.Filename, nil
}
