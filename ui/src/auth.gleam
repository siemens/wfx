// SPDX-FileCopyrightText: 2026 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
import gleam/http/request.{type Request}
import gleam/result
import gleam/uri
import lustre/effect.{type Effect}
import rsvp.{type Handler}

import config.{type Config}

@external(javascript, "./auth.ffi.mjs", "authenticate")
pub fn authenticate(_config: Config, on_ready: fn() -> Nil) -> Nil {
  on_ready()
}

@external(javascript, "./auth.ffi.mjs", "accessToken")
pub fn access_token() -> String {
  ""
}

pub fn authorize_request(
  request: Request(body),
  token: String,
) -> Request(body) {
  case token {
    "" -> request
    token -> request.set_header(request, "authorization", "Bearer " <> token)
  }
}

pub fn get(url: String, handler: Handler(String, message)) -> Effect(message) {
  case uri.parse(url) |> result.try(request.from_uri) {
    Ok(request) ->
      request
      |> authorize_request(access_token())
      |> rsvp.send(handler)
    Error(_) -> rsvp.get(url, handler)
  }
}
