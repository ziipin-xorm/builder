// Copyright 2018 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuilder_Select(t *testing.T) {
	sql, args, err := Select("c, d").From("table1").ToSQL()
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT c, d FROM table1", sql)
	assert.EqualValues(t, []interface{}(nil), args)

	sql, args, err = Select("c, d").From("table1").Where(Eq{"a": 1}).ToSQL()
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT c, d FROM table1 WHERE a=?", sql)
	assert.EqualValues(t, []interface{}{1}, args)

	_, _, err = Select("c, d").ToSQL()
	assert.Error(t, err)
	assert.EqualValues(t, ErrNoTableName, err)
}

func TestBuilderSelectGroupBy(t *testing.T) {
	sql, args, err := Select("c").From("table1").GroupBy("c").Having("count(c)=1").ToSQL()
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT c FROM table1 GROUP BY c HAVING count(c)=1", sql)
	assert.EqualValues(t, 0, len(args))
	fmt.Println(sql, args)

	sql, args, err = Select("c").From("table1").GroupBy("c").Having(Eq{"count(c)": 1}).ToSQL()
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT c FROM table1 GROUP BY c HAVING count(c)=?", sql)
	assert.EqualValues(t, []interface{}{1}, args)
	fmt.Println(sql, args)

	_, _, err = Select("c").From("table1").GroupBy("c").Having(1).ToSQL()
	assert.Error(t, err)
}

func TestBuilderSelectOrderBy(t *testing.T) {
	sql, args, err := Select("c").From("table1").OrderBy("c DESC").ToSQL()
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT c FROM table1 ORDER BY c DESC", sql)
	assert.EqualValues(t, 0, len(args))
	fmt.Println(sql, args)

	sql, args, err = Select("c").From("table1").OrderBy(Expr("CASE WHEN owner_name LIKE ? THEN 0 ELSE 1 END", "a")).ToSQL()
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT c FROM table1 ORDER BY CASE WHEN owner_name LIKE ? THEN 0 ELSE 1 END", sql)
	assert.EqualValues(t, 1, len(args))
	fmt.Println(sql, args)
}

func TestBuilder_From(t *testing.T) {
	// simple one
	sql, args, err := Select("c").From("table1").ToSQL()
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT c FROM table1", sql)
	assert.EqualValues(t, 0, len(args))

	// from sub with alias
	sql, args, err = Select("sub.id").From(Select("id").From("table1").Where(Eq{"a": 1}),
		"sub").Where(Eq{"b": 1}).ToSQL()
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT sub.id FROM (SELECT id FROM table1 WHERE a=?) sub WHERE b=?", sql)
	assert.EqualValues(t, []interface{}{1, 1}, args)

	// from sub without alias and with conditions
	sql, args, err = Select("sub.id").From(Select("id").From("table1").Where(Eq{"a": 1})).Where(Eq{"b": 1}).ToSQL()
	assert.Error(t, err)
	assert.EqualValues(t, ErrUnnamedDerivedTable, err)

	// from sub without alias and conditions
	sql, args, err = Select("sub.id").From(Select("id").From("table1").Where(Eq{"a": 1})).ToSQL()
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT sub.id FROM (SELECT id FROM table1 WHERE a=?)", sql)
	assert.EqualValues(t, []interface{}{1}, args)

	// from union with alias
	sql, args, err = Select("sub.id").From(
		Select("id").From("table1").Where(Eq{"a": 1}).Union(
			"all", Select("id").From("table1").Where(Eq{"a": 2})), "sub").Where(Eq{"b": 1}).ToSQL()
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT sub.id FROM ((SELECT id FROM table1 WHERE a=?) UNION ALL (SELECT id FROM table1 WHERE a=?)) sub WHERE b=?", sql)
	assert.EqualValues(t, []interface{}{1, 2, 1}, args)

	// from union without alias
	_, _, err = Select("sub.id").From(
		Select("id").From("table1").Where(Eq{"a": 1}).Union(
			"all", Select("id").From("table1").Where(Eq{"a": 2}))).Where(Eq{"b": 1}).ToSQL()
	assert.Error(t, err)
	assert.EqualValues(t, ErrUnnamedDerivedTable, err)

	// will raise error
	_, _, err = Select("c").From(Insert(Eq{"a": 1}).From("table1"), "table1").ToSQL()
	assert.Error(t, err)
	assert.EqualValues(t, ErrUnexpectedSubQuery, err)

	// will raise error
	_, _, err = Select("c").From(Delete(Eq{"a": 1}).From("table1"), "table1").ToSQL()
	assert.Error(t, err)
	assert.EqualValues(t, ErrUnexpectedSubQuery, err)

	// from a sub-query in different dialect
	_, _, err = MySQL().Select("sub.id").From(
		Oracle().Select("id").From("table1").Where(Eq{"a": 1}), "sub").Where(Eq{"b": 1}).ToSQL()
	assert.Error(t, err)
	assert.EqualValues(t, ErrInconsistentDialect, err)

	// from a sub-query (dialect set up)
	sql, args, err = MySQL().Select("sub.id").From(
		MySQL().Select("id").From("table1").Where(Eq{"a": 1}), "sub").Where(Eq{"b": 1}).ToSQL()
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT sub.id FROM (SELECT id FROM table1 WHERE a=?) sub WHERE b=?", sql)
	assert.EqualValues(t, []interface{}{1, 1}, args)

	// from a sub-query (dialect not set up)
	sql, args, err = MySQL().Select("sub.id").From(
		Select("id").From("table1").Where(Eq{"a": 1}), "sub").Where(Eq{"b": 1}).ToSQL()
	assert.NoError(t, err)
	assert.EqualValues(t, "SELECT sub.id FROM (SELECT id FROM table1 WHERE a=?) sub WHERE b=?", sql)
	assert.EqualValues(t, []interface{}{1, 1}, args)
}
