package checkpoint

import storageport "fkteams/internal/ports/storage"

// Store 是 checkpoint 实现包内对端口层契约的别名。
type Store = storageport.CheckpointStore
