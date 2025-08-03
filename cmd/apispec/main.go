package main

import (
	api "github.com/maybemaby/oapibase/api"
)

func main() {
	r := api.SpecRouter()

	if err := r.WriteSchemaTo("spec.yaml"); err != nil {
		panic(err)
	}

}
