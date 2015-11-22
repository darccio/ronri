package ronri

type Context struct {
	values map[string]interface{}
}

func (ctx *Context) Set(name string, value interface{}) {
	ctx.values[name] = value
}

func (ctx *Context) Get(name string) (value interface{}, ok bool) {
	value, ok = ctx.values[name]
	return
}

func NewContext(args ...map[string]interface{}) (ctx *Context) {
	ctx = &Context{}
	ctx.values = make(map[string]interface{})
	for _, arg := range args {
		for key, value := range arg {
			ctx.Set(key, value)
		}
	}
	return
}
