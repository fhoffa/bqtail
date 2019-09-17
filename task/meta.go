package task

import "bqtail/base"

var sourceURLExpandable = map[string]bool{
	"move":   true,
	"delete": true,
}

//bodyAppendable job appendable task
var bodyAppendable = map[string]bool{
	"notify": true,
}

//bqJobs bq jobs
var bqJobs = map[string]bool{
	"copy":   true,
	"query":  true,
	"export": true,
}

//replacements replacements key
var replacements = map[string]string{
	base.ErrorKey:       base.ErrorExpr,
	base.SourceTableKey: base.SourceTableExpr,
	base.DestTableKey:   base.DestTableExpr,
	base.JobIDKey:       base.JobIDExpr,
	base.EventIDKey:     base.EventIDExpr,
}
