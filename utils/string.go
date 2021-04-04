package utils

type StringArgs struct {
	UniqueMap map[string]interface{}
}

func NewStringArgs() StringArgs {
	return StringArgs{
		UniqueMap: make(map[string]interface{}),
	}
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
