package parsing

// Options passed to the `NewParser` constructor
type ParserOption func(parser *Parser)

func WithReadFileImplementation(readFile func(filename string) ([]byte, error)) ParserOption {
	return func(parser *Parser) {
		parser.readFile = readFile
	}
}

func WithFindByGlobImplementation(findByGlob func(pattern string) ([]string, error)) ParserOption {
	return func(parser *Parser) {
		parser.findByGlob = findByGlob
	}
}

func WithRecursivelyAddImportsImplementation(recursivelyAddImports func(paths []string) ([]string, error)) ParserOption {
	return func(parser *Parser) {
		parser.RecursivelyAddImports = recursivelyAddImports
	}
}
