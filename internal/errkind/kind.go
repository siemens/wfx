package errkind

import "github.com/Southclaws/fault/ftag"

// TOCTOU is an error kind indicating a Time-Of-Check to Time-Of-Use race.
// For example, an entity (e.g. a job) was concurrently modified between the moment its state was read and the moment an
// update was attempted.
// It will typically result in a 400 response, so clients can retry the operation with fresh data instead of receiving
// an opaque server error.
const TOCTOU = ftag.Kind("TOCTOU")
