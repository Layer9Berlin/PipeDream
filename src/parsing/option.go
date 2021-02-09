package parsing

// ParserOption represents an option that can be passed to the `NewParser` constructor
type ParserOption func(parser *Parser)

// WithReadFileImplementation sets the implementation of the function reading a file's content
//
// Useful for tests.
func WithReadFileImplementation(readFile func(filename string) ([]byte, error)) ParserOption {
	return func(parser *Parser) {
		parser.readFile = readFile
	}
}

// WithFindByGlobImplementation sets the implementation of the function finding files by glob
//
// Useful for tests.
func WithFindByGlobImplementation(findByGlob func(pattern string) ([]string, error)) ParserOption {
	return func(parser *Parser) {
		parser.findByGlob = findByGlob
	}
}

// WithRecursivelyAddImportsImplementation sets the implementation of the function that recursively adds imports
//
// Useful for tests.
func WithRecursivelyAddImportsImplementation(recursivelyAddImports func(paths []string) ([]string, error)) ParserOption {
	return func(parser *Parser) {
		parser.RecursivelyAddImports = recursivelyAddImports
	}
}
