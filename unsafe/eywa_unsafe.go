package unsafe

import "github.com/imperfect-fourth/eywa"

func Get[M eywa.Model, MP eywa.ModelPtr[M]]() eywa.GetQueryBuilder[M, string, eywa.RawField] {
	return eywa.GetQueryBuilder[M, string, eywa.RawField]{
		QuerySkeleton: eywa.QuerySkeleton[M, string, eywa.RawField]{
			ModelName: (*new(M)).ModelName(),
			//			fields:    append(fields, field),
		},
	}
}

func Update[M eywa.Model, MP eywa.ModelPtr[M]]() eywa.UpdateQueryBuilder[M, string, eywa.RawField] {
	return eywa.UpdateQueryBuilder[M, string, eywa.RawField]{
		QuerySkeleton: eywa.QuerySkeleton[M, string, eywa.RawField]{
			ModelName: (*new(M)).ModelName(),
			//			fields:    append(fields, field),
		},
	}
}
