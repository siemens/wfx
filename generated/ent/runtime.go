// SPDX-FileCopyrightText: The entgo authors
// SPDX-License-Identifier: Apache-2.0

// Code generated by ent, DO NOT EDIT.

package ent

import (
	"time"

	"github.com/siemens/wfx/generated/ent/job"
	"github.com/siemens/wfx/generated/ent/schema"
	"github.com/siemens/wfx/generated/ent/workflow"
)

// The init function reads all schema descriptors with runtime code
// (default values, validators, hooks and policies) and stitches it
// to their package variables.
func init() {
	jobFields := schema.Job{}.Fields()
	_ = jobFields
	// jobDescStime is the schema descriptor for stime field.
	jobDescStime := jobFields[1].Descriptor()
	// job.DefaultStime holds the default value on creation for the stime field.
	job.DefaultStime = jobDescStime.Default.(func() time.Time)
	// jobDescMtime is the schema descriptor for mtime field.
	jobDescMtime := jobFields[2].Descriptor()
	// job.DefaultMtime holds the default value on creation for the mtime field.
	job.DefaultMtime = jobDescMtime.Default.(func() time.Time)
	// job.UpdateDefaultMtime holds the default value on update for the mtime field.
	job.UpdateDefaultMtime = jobDescMtime.UpdateDefault.(func() time.Time)
	// jobDescID is the schema descriptor for id field.
	jobDescID := jobFields[0].Descriptor()
	// job.DefaultID holds the default value on creation for the id field.
	job.DefaultID = jobDescID.Default.(func() string)
	// job.IDValidator is a validator for the "id" field. It is called by the builders before save.
	job.IDValidator = func() func(string) error {
		validators := jobDescID.Validators
		fns := [...]func(string) error{
			validators[0].(func(string) error),
			validators[1].(func(string) error),
		}
		return func(id string) error {
			for _, fn := range fns {
				if err := fn(id); err != nil {
					return err
				}
			}
			return nil
		}
	}()
	workflowFields := schema.Workflow{}.Fields()
	_ = workflowFields
	// workflowDescName is the schema descriptor for name field.
	workflowDescName := workflowFields[0].Descriptor()
	// workflow.NameValidator is a validator for the "name" field. It is called by the builders before save.
	workflow.NameValidator = func() func(string) error {
		validators := workflowDescName.Validators
		fns := [...]func(string) error{
			validators[0].(func(string) error),
			validators[1].(func(string) error),
		}
		return func(name string) error {
			for _, fn := range fns {
				if err := fn(name); err != nil {
					return err
				}
			}
			return nil
		}
	}()
	// workflowDescDescription is the schema descriptor for description field.
	workflowDescDescription := workflowFields[1].Descriptor()
	// workflow.DescriptionValidator is a validator for the "description" field. It is called by the builders before save.
	workflow.DescriptionValidator = workflowDescDescription.Validators[0].(func(string) error)
}
