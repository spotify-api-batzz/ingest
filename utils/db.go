package utils

import (
	"fmt"
	"strings"
)

func WrapWithChar(arr []string, char string) []string {
	var res []string
	for _, val := range arr {
		res = append(res, fmt.Sprintf(`%[2]s%[1]s%[2]s`, val, char))
	}

	return res
}

func ColumnNamesExclusive(model interface{}, exclude ...string) string {
	colString := WrapWithChar(ReflectColumns(model), `"`)
	return strings.Join(RemoveExcludedFromSlice(colString, exclude), ",")
}

func ColumnNamesInclusive(model interface{}, include ...string) string {
	colString := WrapWithChar(ReflectColumns(model), `"`)
	return strings.Join(KeepIncludedInSlice(colString, include), ",")
}

func PrepareBatchValuesPG(paramLength int, valueLength int) string {
	counter := 1
	var values string
	for i := 0; i < valueLength; i++ {
		values = fmt.Sprintf("%s, %s", values, genValString(paramLength, &counter))
	}
	return strings.TrimPrefix(values, ", ")
}

func PrepareInStringPG(paramLength int, valueLength int, counter int) string {
	if counter == 0 {
		counter = 1
	}
	var values string
	for i := 0; i < valueLength; i++ {
		values = fmt.Sprintf("%s, %s", values, genValString(paramLength, &counter))
	}
	return strings.TrimPrefix(values, ", ")
}

func genValString(paramLength int, counter *int) string {
	var valString string
	for i := 0; i < paramLength; i++ {
		valString = valString + fmt.Sprintf("$%d,", *counter)
		*counter++
	}
	valString = fmt.Sprintf("(%s)", strings.TrimSuffix(valString, ","))
	return valString
}

type Identifiable interface {
	Identifier() string
}

func NewStringArgsFromModel[T Identifiable](identifiables []T) StringArgs {
	stringArgs := NewStringArgs()
	for _, identifiable := range identifiables {
		stringArgs.Add(identifiable.Identifier())
	}

	return stringArgs
}
