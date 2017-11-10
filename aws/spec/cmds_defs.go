package awsspec

import "github.com/wallix/awless/template"

func AWSLookupDefinitions(key string) (t template.Definition, ok bool) {
	t, ok = AWSTemplatesDefinitions[key]
	return
}
