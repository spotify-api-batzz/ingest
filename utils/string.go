package utils

type StringArgs struct {
	UniqueMap map[string]interface{}
}

func NewStringArgs() StringArgs {
	return StringArgs{
		UniqueMap: make(map[string]interface{}),
	}
}

// Diff computes the symmetric difference between the UniqueMap of the current instance and another StringArgs instance.
func (p *StringArgs) Diff(otherArgs StringArgs) StringArgs {
	newStringArgs := NewStringArgs()

	// Add keys from p that are not in otherArgs
	for id := range p.UniqueMap {
		if _, exists := otherArgs.UniqueMap[id]; !exists {
			newStringArgs.Add(id)
		}
	}

	// Add keys from otherArgs that are not in p
	for id := range otherArgs.UniqueMap {
		if _, exists := p.UniqueMap[id]; !exists {
			newStringArgs.Add(id)
		}
	}

	return newStringArgs
}

func (p *StringArgs) ToString() []string {
	var ids []string
	for id := range p.UniqueMap {
		ids = append(ids, id)
	}

	return ids
}

func (p *StringArgs) Add(value string) {
	if _, ok := p.UniqueMap[value]; !ok {
		p.UniqueMap[value] = value
	}
}

func (p *StringArgs) Set(key string, value interface{}) {
	if _, ok := p.UniqueMap[key]; !ok {
		p.UniqueMap[key] = value
	}
}

func (p *StringArgs) Args() []interface{} {
	args := make([]interface{}, len(p.UniqueMap))
	counter := 0
	for _, val := range p.UniqueMap {
		args[counter] = val
		counter++
	}
	return args
}
