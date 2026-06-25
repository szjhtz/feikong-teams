package checkpoint

import checkpointstore "fkteams/internal/runtime/checkpoint"

type NamespaceStore = checkpointstore.NamespaceStore

func NewNamespaceStore(namespace string, inner Store) Store {
	return checkpointstore.NewNamespaceStore(namespace, inner)
}
