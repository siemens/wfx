// SPDX-FileCopyrightText: The entgo authors
// SPDX-License-Identifier: Apache-2.0

// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/siemens/wfx/generated/ent/history"
	"github.com/siemens/wfx/generated/ent/job"
	"github.com/siemens/wfx/generated/ent/predicate"
	"github.com/siemens/wfx/generated/model"
)

// HistoryUpdate is the builder for updating History entities.
type HistoryUpdate struct {
	config
	hooks    []Hook
	mutation *HistoryMutation
}

// Where appends a list predicates to the HistoryUpdate builder.
func (hu *HistoryUpdate) Where(ps ...predicate.History) *HistoryUpdate {
	hu.mutation.Where(ps...)
	return hu
}

// SetMtime sets the "mtime" field.
func (hu *HistoryUpdate) SetMtime(t time.Time) *HistoryUpdate {
	hu.mutation.SetMtime(t)
	return hu
}

// SetStatus sets the "status" field.
func (hu *HistoryUpdate) SetStatus(ms model.JobStatus) *HistoryUpdate {
	hu.mutation.SetStatus(ms)
	return hu
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (hu *HistoryUpdate) SetNillableStatus(ms *model.JobStatus) *HistoryUpdate {
	if ms != nil {
		hu.SetStatus(*ms)
	}
	return hu
}

// ClearStatus clears the value of the "status" field.
func (hu *HistoryUpdate) ClearStatus() *HistoryUpdate {
	hu.mutation.ClearStatus()
	return hu
}

// SetDefinition sets the "definition" field.
func (hu *HistoryUpdate) SetDefinition(m map[string]interface{}) *HistoryUpdate {
	hu.mutation.SetDefinition(m)
	return hu
}

// ClearDefinition clears the value of the "definition" field.
func (hu *HistoryUpdate) ClearDefinition() *HistoryUpdate {
	hu.mutation.ClearDefinition()
	return hu
}

// SetJobID sets the "job" edge to the Job entity by ID.
func (hu *HistoryUpdate) SetJobID(id string) *HistoryUpdate {
	hu.mutation.SetJobID(id)
	return hu
}

// SetNillableJobID sets the "job" edge to the Job entity by ID if the given value is not nil.
func (hu *HistoryUpdate) SetNillableJobID(id *string) *HistoryUpdate {
	if id != nil {
		hu = hu.SetJobID(*id)
	}
	return hu
}

// SetJob sets the "job" edge to the Job entity.
func (hu *HistoryUpdate) SetJob(j *Job) *HistoryUpdate {
	return hu.SetJobID(j.ID)
}

// Mutation returns the HistoryMutation object of the builder.
func (hu *HistoryUpdate) Mutation() *HistoryMutation {
	return hu.mutation
}

// ClearJob clears the "job" edge to the Job entity.
func (hu *HistoryUpdate) ClearJob() *HistoryUpdate {
	hu.mutation.ClearJob()
	return hu
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (hu *HistoryUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, hu.sqlSave, hu.mutation, hu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (hu *HistoryUpdate) SaveX(ctx context.Context) int {
	affected, err := hu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (hu *HistoryUpdate) Exec(ctx context.Context) error {
	_, err := hu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (hu *HistoryUpdate) ExecX(ctx context.Context) {
	if err := hu.Exec(ctx); err != nil {
		panic(err)
	}
}

func (hu *HistoryUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := sqlgraph.NewUpdateSpec(history.Table, history.Columns, sqlgraph.NewFieldSpec(history.FieldID, field.TypeInt))
	if ps := hu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := hu.mutation.Mtime(); ok {
		_spec.SetField(history.FieldMtime, field.TypeTime, value)
	}
	if value, ok := hu.mutation.Status(); ok {
		_spec.SetField(history.FieldStatus, field.TypeJSON, value)
	}
	if hu.mutation.StatusCleared() {
		_spec.ClearField(history.FieldStatus, field.TypeJSON)
	}
	if value, ok := hu.mutation.Definition(); ok {
		_spec.SetField(history.FieldDefinition, field.TypeJSON, value)
	}
	if hu.mutation.DefinitionCleared() {
		_spec.ClearField(history.FieldDefinition, field.TypeJSON)
	}
	if hu.mutation.JobCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   history.JobTable,
			Columns: []string{history.JobColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(job.FieldID, field.TypeString),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := hu.mutation.JobIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   history.JobTable,
			Columns: []string{history.JobColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(job.FieldID, field.TypeString),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, hu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{history.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	hu.mutation.done = true
	return n, nil
}

// HistoryUpdateOne is the builder for updating a single History entity.
type HistoryUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *HistoryMutation
}

// SetMtime sets the "mtime" field.
func (huo *HistoryUpdateOne) SetMtime(t time.Time) *HistoryUpdateOne {
	huo.mutation.SetMtime(t)
	return huo
}

// SetStatus sets the "status" field.
func (huo *HistoryUpdateOne) SetStatus(ms model.JobStatus) *HistoryUpdateOne {
	huo.mutation.SetStatus(ms)
	return huo
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (huo *HistoryUpdateOne) SetNillableStatus(ms *model.JobStatus) *HistoryUpdateOne {
	if ms != nil {
		huo.SetStatus(*ms)
	}
	return huo
}

// ClearStatus clears the value of the "status" field.
func (huo *HistoryUpdateOne) ClearStatus() *HistoryUpdateOne {
	huo.mutation.ClearStatus()
	return huo
}

// SetDefinition sets the "definition" field.
func (huo *HistoryUpdateOne) SetDefinition(m map[string]interface{}) *HistoryUpdateOne {
	huo.mutation.SetDefinition(m)
	return huo
}

// ClearDefinition clears the value of the "definition" field.
func (huo *HistoryUpdateOne) ClearDefinition() *HistoryUpdateOne {
	huo.mutation.ClearDefinition()
	return huo
}

// SetJobID sets the "job" edge to the Job entity by ID.
func (huo *HistoryUpdateOne) SetJobID(id string) *HistoryUpdateOne {
	huo.mutation.SetJobID(id)
	return huo
}

// SetNillableJobID sets the "job" edge to the Job entity by ID if the given value is not nil.
func (huo *HistoryUpdateOne) SetNillableJobID(id *string) *HistoryUpdateOne {
	if id != nil {
		huo = huo.SetJobID(*id)
	}
	return huo
}

// SetJob sets the "job" edge to the Job entity.
func (huo *HistoryUpdateOne) SetJob(j *Job) *HistoryUpdateOne {
	return huo.SetJobID(j.ID)
}

// Mutation returns the HistoryMutation object of the builder.
func (huo *HistoryUpdateOne) Mutation() *HistoryMutation {
	return huo.mutation
}

// ClearJob clears the "job" edge to the Job entity.
func (huo *HistoryUpdateOne) ClearJob() *HistoryUpdateOne {
	huo.mutation.ClearJob()
	return huo
}

// Where appends a list predicates to the HistoryUpdate builder.
func (huo *HistoryUpdateOne) Where(ps ...predicate.History) *HistoryUpdateOne {
	huo.mutation.Where(ps...)
	return huo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (huo *HistoryUpdateOne) Select(field string, fields ...string) *HistoryUpdateOne {
	huo.fields = append([]string{field}, fields...)
	return huo
}

// Save executes the query and returns the updated History entity.
func (huo *HistoryUpdateOne) Save(ctx context.Context) (*History, error) {
	return withHooks(ctx, huo.sqlSave, huo.mutation, huo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (huo *HistoryUpdateOne) SaveX(ctx context.Context) *History {
	node, err := huo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (huo *HistoryUpdateOne) Exec(ctx context.Context) error {
	_, err := huo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (huo *HistoryUpdateOne) ExecX(ctx context.Context) {
	if err := huo.Exec(ctx); err != nil {
		panic(err)
	}
}

func (huo *HistoryUpdateOne) sqlSave(ctx context.Context) (_node *History, err error) {
	_spec := sqlgraph.NewUpdateSpec(history.Table, history.Columns, sqlgraph.NewFieldSpec(history.FieldID, field.TypeInt))
	id, ok := huo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "History.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := huo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, history.FieldID)
		for _, f := range fields {
			if !history.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != history.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := huo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := huo.mutation.Mtime(); ok {
		_spec.SetField(history.FieldMtime, field.TypeTime, value)
	}
	if value, ok := huo.mutation.Status(); ok {
		_spec.SetField(history.FieldStatus, field.TypeJSON, value)
	}
	if huo.mutation.StatusCleared() {
		_spec.ClearField(history.FieldStatus, field.TypeJSON)
	}
	if value, ok := huo.mutation.Definition(); ok {
		_spec.SetField(history.FieldDefinition, field.TypeJSON, value)
	}
	if huo.mutation.DefinitionCleared() {
		_spec.ClearField(history.FieldDefinition, field.TypeJSON)
	}
	if huo.mutation.JobCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   history.JobTable,
			Columns: []string{history.JobColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(job.FieldID, field.TypeString),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := huo.mutation.JobIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   history.JobTable,
			Columns: []string{history.JobColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(job.FieldID, field.TypeString),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_node = &History{config: huo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, huo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{history.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	huo.mutation.done = true
	return _node, nil
}