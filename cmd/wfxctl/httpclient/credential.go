package httpclient

/*
 * SPDX-FileCopyrightText: 2026 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"strings"

	"github.com/Southclaws/fault"
)

type credential struct {
	username       string
	password       string
	authType       string
	authCredential string
	hasUsername    bool
	hasPassword    bool
}

func credentialHelperName(helper string) string {
	if strings.ContainsRune(helper, '/') {
		return helper
	}
	return "wfxctl-credential-" + helper
}

func getCredential(ctx context.Context, helper string, u *url.URL) (credential, error) {
	cmd := exec.CommandContext(ctx, credentialHelperName(helper), "get")
	cmd.Stdin = strings.NewReader(fmt.Sprintf("protocol=%s\nhost=%s\npath=%s\n\n", u.Scheme, u.Host, strings.TrimPrefix(u.Path, "/")))
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	output, err := cmd.Output()
	if err != nil {
		message := strings.TrimSpace(stderr.String())
		if message != "" {
			return credential{}, fmt.Errorf("credential helper failed: %s: %w", message, err)
		}
		return credential{}, fmt.Errorf("credential helper failed: %w", err)
	}

	var result credential
	for line := range strings.SplitSeq(string(output), "\n") {
		line = strings.TrimSuffix(line, "\r")
		if line == "" {
			break
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			return credential{}, fmt.Errorf("invalid response from credential helper: expected key=value")
		}
		switch key {
		case "username":
			result.username, result.hasUsername = value, true
		case "password":
			result.password, result.hasPassword = value, true
		case "authtype":
			result.authType = value
		case "credential":
			result.authCredential = value
		}
	}
	return result, nil
}

func (c credential) authorization() (string, error) {
	switch {
	case c.authType != "" && c.authCredential != "":
		if strings.ContainsAny(c.authType+c.authCredential, "\r\n") {
			return "", fmt.Errorf("invalid credential helper response")
		}
		return c.authType + " " + c.authCredential, nil
	case c.hasUsername && c.hasPassword:
		value := base64.StdEncoding.EncodeToString([]byte(c.username + ":" + c.password))
		return "Basic " + value, nil
	default:
		return "", fmt.Errorf("credential helper returned no complete credential")
	}
}

func addCredential(ctx context.Context, helper string, req *http.Request) error {
	credential, err := getCredential(ctx, helper, req.URL)
	if err != nil {
		return fault.Wrap(err)
	}
	authorization, err := credential.authorization()
	if err != nil {
		return fault.Wrap(err)
	}
	req.Header.Set("Authorization", authorization)
	return nil
}
