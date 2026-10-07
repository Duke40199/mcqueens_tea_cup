package repository

import "github.com/lib/pq"

// quoteIdent safely quotes a SQL identifier (table/column name) for interpolation
// into a query string. Table names come from trusted config, but quoting is cheap
// belt-and-suspenders: it prevents a malformed/typo'd name (e.g. one containing a
// hyphen or reserved word) from producing surprising SQL, and documents that the
// value is an identifier rather than a bound parameter.
func quoteIdent(name string) string {
	return pq.QuoteIdentifier(name)
}
