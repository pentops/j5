package j5schema

func stringSliceConvert[A ~string, B ~string](in []A) []B {
	out := make([]B, len(in))
	for i, v := range in {
		out[i] = B(v)
	}
	return out
}
