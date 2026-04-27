package eywa

type root interface {
	root()
}
type QueryRoot interface {
	queryRoot()

	root
	SelectionSet
}
type MutationRoot interface {
	mutationRoot()

	root
	SelectionSet
}
