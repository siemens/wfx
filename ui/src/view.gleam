// SPDX-FileCopyrightText: 2025 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//
// Author: Michael Adler <michael.adler@siemens.com>
import gleam/http/response
import gleam/int
import gleam/json
import gleam/list
import gleam/option.{None, Some}
import gleam/string_tree
import gleam/time/calendar
import gleam/time/timestamp
import lustre/attribute.{class, colspan, href, src, title}
import lustre/element.{type Element}
import lustre/element/html.{
  a, button, div, h1, img, nav, p, table, td, text, tfoot, th, thead, tr,
}
import lustre/element/keyed
import lustre/event
import plinth/javascript/date
import rsvp

import highlightjs
import model.{type Model}
import msg.{type Msg}
import wfx
import wfx/encoder
import wfx/utils

const class_table_header = "px-4 py-2 text-left bg-gray-100"

const class_table_row = "px-4 py-2 text-left"

const class_link = "text-blue-500 hover:text-blue-700 underline"

pub fn view(model: Model) -> Element(Msg) {
  let empty = div([], [])

  let loading = p([], [text("Waiting for response from wfx...")])

  let content =
    div(
      [
        class(
          "flex-1 flex flex-col items-center justify-start mt-6 text-center",
        ),
      ],
      [
        case model.route {
          model.RouteJobs(_, _) ->
            model.paginated_jobs
            |> option.map(view_paginated_jobs(_, model.base_path))
            |> option.unwrap(loading)
          model.RouteWorkflows(_, _) ->
            model.paginated_workflows
            |> option.map(view_paginated_workflows(_, model.base_path))
            |> option.unwrap(loading)
          model.RouteWorkflowDetails(_) ->
            model.workflow
            |> option.map(view_single_workflow)
            |> option.unwrap(loading)
          model.RouteJobDetails(_) ->
            model.job |> option.map(view_single_job) |> option.unwrap(loading)
        },
      ],
    )

  div(
    [
      class("flex flex-col min-h-screen"),
    ],
    [
      // Top Bar
      nav(
        [
          class(
            "bg-gray-800 text-white px-4 py-2 flex justify-between items-center",
          ),
        ],
        [
          empty,
          div([class("space-x-4")], [
            a([href("https://github.com/siemens/wfx")], [
              img([
                class("w-12 h-12"),
                src(model.base_path <> "/logo.svg"),
              ]),
            ]),
          ]),
          empty,
        ],
      ),
      // Tab Bar
      div([class("flex justify-center")], {
        let css_active =
          "px-4 py-2 cursor-pointer text-blue-600 border-b-2 border-blue-600 font-semibold"
        let css_inactive =
          "px-4 py-2 cursor-pointer text-gray-600 hover:text-blue-600"
        [
          div(
            [
              class(case model.route {
                model.RouteJobs(_, _) -> css_active
                _ -> css_inactive
              }),
            ],
            [
              a([href(model.base_path <> "/jobs")], [text("Jobs")]),
            ],
          ),
          div(
            [
              class(case model.route {
                model.RouteWorkflows(_, _) -> css_active
                _ -> css_inactive
              }),
            ],
            [
              a([href(model.base_path <> "/workflows")], [
                text("Workflows"),
              ]),
            ],
          ),
        ]
      }),
      content,
    ],
  )
}

fn format_rsvp_error(err: rsvp.Error) -> String {
  case err {
    rsvp.BadBody -> "Bad body"
    rsvp.BadUrl(_) -> "Bad URL"
    rsvp.HttpError(response.Response(status: _, headers: _, body: body)) -> body
    rsvp.JsonError(_) -> "Failed to unmarshal JSON"
    rsvp.NetworkError -> "Network Error"
    rsvp.UnhandledResponse(_) -> "Unexpected response type"
  }
}

fn view_paginated_jobs(
  paginated: model.PaginatedJobs,
  base_path: String,
) -> Element(Msg) {
  let view_job_row = fn(job: wfx.Job) {
    #(
      job.id,
      tr([class("text-left odd:bg-white even:bg-gray-100")], [
        td([class(class_table_row <> " font-mono")], [
          a(
            [
              href(base_path <> "/jobs/" <> job.id),
              class(class_link),
            ],
            [
              text(job.id),
            ],
          ),
        ]),
        td([class(class_table_row)], [
          text(job.client_id),
        ]),
        td([class(class_table_row)], [
          a(
            [
              href(base_path <> "/workflows/" <> job.workflow.name),
              class(class_link),
            ],
            [
              text(job.workflow.name),
            ],
          ),
        ]),
        td([class(class_table_row)], [
          text(
            job.status
            |> option.map(fn(x) { x.state })
            |> option.unwrap(""),
          ),
        ]),
        td([class(class_table_row)], [
          text(
            utils.job_group(job)
            |> option.map(fn(group) { group.name })
            |> option.unwrap(""),
          ),
        ]),
        td([class(class_table_row <> " font-mono")], [
          text(
            job.stime
            |> option.map(format_timestamp)
            |> option.unwrap("n/a"),
          ),
        ]),
        td([class(class_table_row <> " font-mono")], [
          text(
            job.mtime
            |> option.map(format_timestamp)
            |> option.unwrap("n/a"),
          ),
        ]),
      ]),
    )
  }

  let th_elements = [
    th([class(class_table_header <> " w-100")], [text("ID")]),
    th([class(class_table_header <> " w-32")], [text("Client")]),
    th([class(class_table_header <> " w-64")], [text("Workflow")]),
    th([class(class_table_header <> " w-32")], [text("Status")]),
    th([class(class_table_header <> " w-32")], [text("Group")]),
    th([class(class_table_header <> " w-64")], [text("Created")]),
    th([class(class_table_header <> " w-64")], [text("Last Modified")]),
  ]

  case paginated.response {
    Ok(jobs) ->
      div([class("overflow-x-auto")], [
        table([attribute.id("jobs-table"), class("table-fixed")], [
          thead([], [
            tr([class("border-b")], th_elements),
          ]),
          keyed.tbody([], list.map(jobs.content, view_job_row)),
          create_pagination(
            base_path <> "/jobs",
            jobs.pagination,
            list.length(th_elements),
          ),
        ]),
      ])
    Error(err) -> view_api_error(err)
  }
}

fn view_paginated_workflows(
  paginated: model.PaginatedWorkflows,
  base_path: String,
) -> Element(Msg) {
  let th_elements = [
    th([class(class_table_header)], [text("Name")]),
    th([class(class_table_header)], [text("Description")]),
  ]

  case paginated.response {
    Ok(workflows) ->
      div([class("overflow-x-auto")], [
        table(
          [
            attribute.id("workflows-table"),
            class("min-w-full border border-gray-300 rounded-lg shadow-sm"),
          ],
          [
            thead([], [
              tr([class("border-b")], th_elements),
            ]),
            keyed.tbody(
              [],
              workflows.content
                |> list.map(fn(workflow) {
                  #(
                    workflow.name,
                    tr([class("text-left odd:bg-white even:bg-gray-100")], [
                      td([class(class_table_row)], [
                        a(
                          [
                            href(base_path <> "/workflows/" <> workflow.name),
                            class(class_link),
                          ],
                          [
                            text(workflow.name),
                          ],
                        ),
                      ]),
                      td([class(class_table_row)], [
                        text(workflow.description |> option.unwrap("")),
                      ]),
                    ]),
                  )
                }),
            ),
            create_pagination(
              base_path <> "/workflows",
              workflows.pagination,
              list.length(th_elements),
            ),
          ],
        ),
      ])
    Error(err) -> view_api_error(err)
  }
}

fn view_api_error(err: model.ApiError) -> Element(Msg) {
  case err {
    model.RsvpError(err) -> p([], [text("Error: " <> format_rsvp_error(err))])
  }
}

fn ceil(a: Int, b: Int) -> Int {
  { a + b - 1 } / b
}

fn create_pagination(
  prefix: String,
  pagination: wfx.Pagination,
  column_count: Int,
) -> Element(Msg) {
  let pages = ceil(pagination.total, pagination.limit)
  let current_page = utils.page_from_pagination(pagination)

  let links = case pages > 1 {
    False -> []
    True ->
      list.range(1, pages)
      |> list.map(fn(page) {
        case page == current_page {
          True -> p([], [text(int.to_string(page))])
          False ->
            a(
              [
                href(prefix <> "?page=" <> int.to_string(page)),
                class(class_link),
              ],
              [
                text(int.to_string(page)),
              ],
            )
        }
      })
  }

  let link_prev =
    a(
      [
        href(prefix <> "?page=" <> int.to_string(current_page - 1)),
        class(class_link),
      ],
      [
        text("«"),
      ],
    )

  let link_next =
    a(
      [
        href(prefix <> "?page=" <> int.to_string(current_page + 1)),
        class(class_link),
      ],
      [
        text("»"),
      ],
    )

  // add left button
  let links = case current_page {
    1 -> links
    _ -> [link_prev, ..links]
  }
  let links = case current_page < pages {
    True -> list.append(links, [link_next])
    False -> links
  }

  let elements = [
    button(
      [
        class("rounded cursor-pointer"),
        title("Refresh"),
        event.on_click(msg.UserClickedRefresh),
      ],
      [text("↻")],
    ),
    ..links
  ]

  tfoot([], [
    tr([], [
      td([class("px-6 py-4 text-center"), colspan(column_count)], [
        div(
          [class("inline-flex justify-center items-center space-x-2")],
          elements,
        ),
      ]),
    ]),
  ])
}

fn view_single_workflow(model: model.Workflow) -> Element(Msg) {
  case model.response {
    Ok(workflow) -> {
      div([class("flex flex-col justify-center items-center")], [
        h1([class("text-xl font-bold")], [text(workflow.name)]),
        case model.mermaid_diagram {
          None -> p([], [text("Rendering...")])
          Some(html) ->
            html.pre([class("mermaid mb-4")], [
              element.unsafe_raw_html("", "mermaid", [], html),
            ])
        },
        html.h2([class("text-xl font-bold")], [text("JSON")]),
        render_code(
          id: "workflow-json",
          code: encoder.workflow(workflow) |> json_to_string,
          lang: "json",
          copied: model.copied_to_clipboard,
        ),
      ])
    }
    Error(err) -> view_api_error(err)
  }
}

fn view_single_job(model: model.Job) -> Element(Msg) {
  case model.response {
    Ok(job) -> {
      div([class("flex flex-col justify-center items-center")], [
        h1([class("text-xl font-bold mb-4")], [text("Job Details")]),

        div([class("overflow-x-auto")], [
          table([attribute.id("job-details"), class("table-fixed mb-4")], [
            thead([], [
              tr([class("border-b")], [
                th([class(class_table_header <> " w-32")], [text("Attribute")]),
                th([class(class_table_header <> " w-100")], [text("Value")]),
              ]),
            ]),
            html.tbody([], [
              tr([class("text-left odd:bg-white even:bg-gray-100")], [
                td([class(class_table_row <> " text-left")], [text("Job ID")]),
                td([class(class_table_row <> " text-center font-mono")], [
                  text(job.id),
                ]),
              ]),

              tr([class("text-left odd:bg-white even:bg-gray-100")], [
                td([class(class_table_row <> " text-left")], [text("Client ID")]),
                td([class(class_table_row <> " text-center font-mono")], [
                  text(job.client_id),
                ]),
              ]),

              tr([class("text-left odd:bg-white even:bg-gray-100")], [
                td([class(class_table_row <> " text-left")], [text("State")]),
                td([class(class_table_row <> " text-center font-mono")], [
                  text(
                    job.status
                    |> option.map(fn(x) { x.state })
                    |> option.unwrap("n/a"),
                  ),
                ]),
              ]),

              tr([class("text-left odd:bg-white even:bg-gray-100")], [
                td([class(class_table_row <> " text-left")], [text("Group")]),
                td([class(class_table_row <> " text-center font-mono")], [
                  text(
                    utils.job_group(job)
                    |> option.map(fn(group) { group.name })
                    |> option.unwrap(""),
                  ),
                ]),
              ]),

              tr([class("text-left odd:bg-white even:bg-gray-100")], [
                td([class(class_table_row <> " text-left")], [text("Tags")]),
                td([class(class_table_row <> " text-center font-mono")], [
                  text(
                    job.tags
                    |> list.map(string_tree.from_string)
                    |> string_tree.join(", ")
                    |> string_tree.to_string,
                  ),
                ]),
              ]),

              tr([class("text-left odd:bg-white even:bg-gray-100")], [
                td([class(class_table_row <> " text-left")], [text("Created")]),
                td([class(class_table_row <> " text-center font-mono")], [
                  text(
                    job.stime
                    |> option.map(format_timestamp)
                    |> option.unwrap("n/a"),
                  ),
                ]),
              ]),

              tr([class("text-left odd:bg-white even:bg-gray-100")], [
                td([class(class_table_row <> " text-left")], [text("Modified")]),
                td([class(class_table_row <> " text-center font-mono")], [
                  text(
                    job.mtime
                    |> option.map(format_timestamp)
                    |> option.unwrap("n/a"),
                  ),
                ]),
              ]),
            ]),
          ]),
        ]),
        div([class("mb-4")], [
          html.h2([class("text-xl font-bold")], [text(job.workflow.name)]),
          case model.mermaid_diagram {
            None -> p([], [text("Rendering...")])
            Some(html) ->
              html.pre([class("mermaid mb-4")], [
                element.unsafe_raw_html("", "mermaid", [], html),
              ])
          },
        ]),
        html.h2([class("text-xl font-bold")], [text("JSON")]),
        render_code(
          id: "job-json",
          code: encoder.job(job) |> json_to_string,
          lang: "json",
          copied: model.copied_to_clipboard,
        ),
      ])
    }
    Error(err) -> view_api_error(err)
  }
}

fn format_timestamp(ts: timestamp.Timestamp) -> String {
  let #(date, tod) = timestamp.to_calendar(ts, calendar.local_offset())
  let #(year, month, day) = #(date.year, date.month, date.day)
  let #(hours, minutes, seconds) = #(tod.hours, tod.minutes, tod.seconds)
  let to_padded_str = fn(x) {
    case x < 10 {
      True -> "0" <> int.to_string(x)
      False -> int.to_string(x)
    }
  }
  let date_str =
    string_tree.new()
    |> string_tree.append(month |> calendar.month_to_int |> to_padded_str)
    |> string_tree.append("/")
    |> string_tree.append(to_padded_str(day))
    |> string_tree.append("/")
    |> string_tree.append(to_padded_str(year))
  let time_str =
    string_tree.new()
    |> string_tree.append(to_padded_str(hours))
    |> string_tree.append(":")
    |> string_tree.append(to_padded_str(minutes))
    |> string_tree.append(":")
    |> string_tree.append(to_padded_str(seconds))

  string_tree.join([date_str, time_str], " ")
  |> string_tree.to_string
}

@external(javascript, "./view.ffi.mjs", "jsonToString")
fn json_to_string(a: json.Json) -> String

fn render_code(
  id id: String,
  code code: String,
  lang lang: String,
  copied copied: Bool,
) {
  let highlighted = highlightjs.highlight_code(code, lang)
  div([class("relative w-full mx-auto my-8")], [
    html.button(
      [
        event.on_click(msg.CopyToClipboard(code)),
        attribute.title("Copy to clipboard"),
        class(
          "absolute bottom-3 right-2 bg-transparent text-gray-800 px-3 py-1 rounded text-xl transition cursor-pointer",
        ),
      ],
      [
        text(case copied {
          False -> "📋"
          True -> "Copied!"
        }),
      ],
    ),
    html.pre(
      [
        class(
          "text-left bg-gray-100 text-gray-800 rounded p-4 overflow-x-auto font-mono text-sm border border-gray-200",
        ),
      ],
      [
        element.unsafe_raw_html(
          "",
          "code",
          [
            attribute.id(id),
            class(
              "block max-w-4xl max-h-[40rem] w-full p-2 bg-white border rounded font-mono text-sm overflow-auto",
            ),
          ],
          highlighted,
        ),
      ],
    ),
  ])
}
