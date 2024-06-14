package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockModel and MockIdentifiable are mock implementations of the Model and Identifiable interfaces.
type MockModel struct {
	Column1 string `db:"column1"`
	Column2 string `db:"column2"`
	Column3 string `db:"column3"`
}

type MockIdentifiable struct {
	id string
}

func (mi MockIdentifiable) Identifier() string {
	return mi.id
}

func TestColumnNamesExclusive(t *testing.T) {
	model := MockModel{}
	columns := ColumnNamesExclusive(model, "exclude1", "exclude2")
	assert.Equal(t, `"column1","column2","column3"`, columns, "Column names are not as expected")
}

func TestColumnNamesInclusive(t *testing.T) {
	model := MockModel{}
	columns := ColumnNamesInclusive(model, "column1", "column3")
	assert.Equal(t, `"column1","column3"`, columns, "Column names are not as expected")
}

func TestPrepareBatchValuesPG(t *testing.T) {
	values := PrepareBatchValuesPG(3, 4)
	expected := "($1,$2,$3), ($4,$5,$6), ($7,$8,$9), ($10,$11,$12)"
	assert.Equal(t, expected, values, "Batch values are not as expected")
}

func TestPrepareInStringPG(t *testing.T) {
	values := PrepareInStringPG(3, 4, 1)
	expected := "($1,$2,$3), ($4,$5,$6), ($7,$8,$9), ($10,$11,$12)"
	assert.Equal(t, expected, values, "In string values are not as expected")
}

func TestNewStringArgsFromModel(t *testing.T) {
	identifiables := []MockIdentifiable{
		{id: "id1"},
		{id: "id2"},
		{id: "id3"},
	}

	stringArgs := NewStringArgsFromModel(identifiables)

	expMap := make(map[string]interface{})
	expMap["id1"] = "id1"
	expMap["id2"] = "id2"
	expMap["id3"] = "id3"
	expected := StringArgs{UniqueMap: expMap}
	assert.Equal(t, expected, stringArgs, "StringArgs from model are not as expected")
}
