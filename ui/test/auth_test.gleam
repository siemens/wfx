// SPDX-FileCopyrightText: 2026 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
import gleam/http/request

import auth

pub fn authorize_request_test() {
  let authorization =
    request.new()
    |> auth.authorize_request("secret")
    |> request.get_header("authorization")

  assert authorization == Ok("Bearer secret")
}

pub fn authorize_request_without_token_test() {
  let authorization =
    request.new()
    |> auth.authorize_request("")
    |> request.get_header("authorization")

  assert authorization == Error(Nil)
}
