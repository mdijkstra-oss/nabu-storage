package persistence

var stores = map[string]*Store{
	"File": NewStore(),
	"Code": NewStore(),
}

func StoreForNoun(noun string) *Store {
	return stores[noun]
}
