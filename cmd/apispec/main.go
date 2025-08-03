package main

import (
	api "github.com/maybemaby/oapibase/apiv2"
)

func main() {
	r := api.SpecRouter()

	if err := r.WriteSchemaTo("spec.yaml"); err != nil {
		panic(err)
	}

}
