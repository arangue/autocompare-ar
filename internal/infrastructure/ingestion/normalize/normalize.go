package normalize

import "strings"

func Title(input string) string {
	input = strings.TrimSpace(input)
	input = strings.Join(strings.Fields(input), " ")
	input = stripTildes(input)

	return strings.ToUpper(input)
}

var tildes = strings.NewReplacer(
	"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n",
	"Á", "A", "É", "E", "Í", "I", "Ó", "O", "Ú", "U", "Ü", "U", "Ñ", "N",
)

func stripTildes(input string) string {
	return tildes.Replace(input)
}
