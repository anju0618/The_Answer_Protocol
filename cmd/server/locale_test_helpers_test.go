package main

func en(s string) LocalizedText {
	return LocalizedText{"en": s}
}

func ens(lines ...string) []LocalizedText {
	out := make([]LocalizedText, len(lines))
	for i, line := range lines {
		out[i] = en(line)
	}
	return out
}
