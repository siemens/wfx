// Code generated by go-swagger; DO NOT EDIT.

// SPDX-FileCopyrightText: 2023 Siemens AG
//
// SPDX-License-Identifier: Apache-2.0
//

package model

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"strconv"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// Job job
//
// swagger:model Job
type Job struct {

	// client Id
	// Example: client42
	ClientID string `json:"clientId,omitempty"`

	// definition
	// Example: { \"userDefined\": {} }\n
	Definition map[string]interface{} `json:"definition,omitempty"`

	// The job's history. Last in, first out (LIFO). Array is truncated if its length exceeds the maximum allowed length.
	// By default, the job history is omitted from responses unless explicitly requested by the client (see the `history` URL parameter).
	//
	// Max Items: 8192
	History []*History `json:"history,omitempty"`

	// Unique job ID (wfx-generated)
	// Example: 3307e5cb-074c-49b7-99d4-5e61839a4c2d
	// Read Only: true
	// Max Length: 36
	ID string `json:"id,omitempty"`

	// Date and time (ISO8601) when the job was last modified (set by wfx)
	// Read Only: true
	// Format: date-time
	Mtime *strfmt.DateTime `json:"mtime,omitempty"`

	// status
	Status *JobStatus `json:"status,omitempty"`

	// Date and time (ISO8601) when the job was created (set by wfx). Although stime conceptually always exists, it's nullable because we don't want to serialize stime in some cases (e.g. for job events).
	// Read Only: true
	// Format: date-time
	Stime *strfmt.DateTime `json:"stime,omitempty"`

	// tags
	Tags []string `json:"tags,omitempty"`

	// workflow
	Workflow *Workflow `json:"workflow,omitempty"`
}

// Validate validates this job
func (m *Job) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateHistory(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateID(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateMtime(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateStatus(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateStime(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateWorkflow(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Job) validateHistory(formats strfmt.Registry) error {
	if swag.IsZero(m.History) { // not required
		return nil
	}

	iHistorySize := int64(len(m.History))

	if err := validate.MaxItems("history", "body", iHistorySize, 8192); err != nil {
		return err
	}

	for i := 0; i < len(m.History); i++ {
		if swag.IsZero(m.History[i]) { // not required
			continue
		}

		if m.History[i] != nil {
			if err := m.History[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("history" + "." + strconv.Itoa(i))
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("history" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

func (m *Job) validateID(formats strfmt.Registry) error {
	if swag.IsZero(m.ID) { // not required
		return nil
	}

	if err := validate.MaxLength("id", "body", m.ID, 36); err != nil {
		return err
	}

	return nil
}

func (m *Job) validateMtime(formats strfmt.Registry) error {
	if swag.IsZero(m.Mtime) { // not required
		return nil
	}

	if err := validate.FormatOf("mtime", "body", "date-time", m.Mtime.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *Job) validateStatus(formats strfmt.Registry) error {
	if swag.IsZero(m.Status) { // not required
		return nil
	}

	if m.Status != nil {
		if err := m.Status.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("status")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("status")
			}
			return err
		}
	}

	return nil
}

func (m *Job) validateStime(formats strfmt.Registry) error {
	if swag.IsZero(m.Stime) { // not required
		return nil
	}

	if err := validate.FormatOf("stime", "body", "date-time", m.Stime.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *Job) validateWorkflow(formats strfmt.Registry) error {
	if swag.IsZero(m.Workflow) { // not required
		return nil
	}

	if m.Workflow != nil {
		if err := m.Workflow.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("workflow")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("workflow")
			}
			return err
		}
	}

	return nil
}

// ContextValidate validate this job based on the context it is used
func (m *Job) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateHistory(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateID(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateMtime(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateStatus(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateStime(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateWorkflow(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Job) contextValidateHistory(ctx context.Context, formats strfmt.Registry) error {

	for i := 0; i < len(m.History); i++ {

		if m.History[i] != nil {

			if swag.IsZero(m.History[i]) { // not required
				return nil
			}

			if err := m.History[i].ContextValidate(ctx, formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("history" + "." + strconv.Itoa(i))
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("history" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

func (m *Job) contextValidateID(ctx context.Context, formats strfmt.Registry) error {

	if err := validate.ReadOnly(ctx, "id", "body", string(m.ID)); err != nil {
		return err
	}

	return nil
}

func (m *Job) contextValidateMtime(ctx context.Context, formats strfmt.Registry) error {

	if err := validate.ReadOnly(ctx, "mtime", "body", m.Mtime); err != nil {
		return err
	}

	return nil
}

func (m *Job) contextValidateStatus(ctx context.Context, formats strfmt.Registry) error {

	if m.Status != nil {

		if swag.IsZero(m.Status) { // not required
			return nil
		}

		if err := m.Status.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("status")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("status")
			}
			return err
		}
	}

	return nil
}

func (m *Job) contextValidateStime(ctx context.Context, formats strfmt.Registry) error {

	if err := validate.ReadOnly(ctx, "stime", "body", m.Stime); err != nil {
		return err
	}

	return nil
}

func (m *Job) contextValidateWorkflow(ctx context.Context, formats strfmt.Registry) error {

	if m.Workflow != nil {

		if swag.IsZero(m.Workflow) { // not required
			return nil
		}

		if err := m.Workflow.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("workflow")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("workflow")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *Job) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Job) UnmarshalBinary(b []byte) error {
	var res Job
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
