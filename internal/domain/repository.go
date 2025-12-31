package domain

// 2. Define the Interface
// The rest of your app will rely on THIS interface, not the concrete "AliasRepo".
type AliasRepo interface {
	Get(key string) (PlayerAlias, bool)
	Set(key, ign, area string) error
	Load() error // You can keep this even if it does nothing for DB
}
