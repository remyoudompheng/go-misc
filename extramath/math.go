/*
package extramath implements toy assembly routines demonstrating MUL and DIV
*/
package extramath

// Divmod computes the quotient and remainder of division of a by b.
func DivmodU64(a, b uint64) (quo, rem uint64)

func MulI64(a, b int64) (hi int64, lo uint64)

// Mul computes the 128-bit product of a by b as hi<<64|lo.
func MulU64(a, b uint64) (hi, lo uint64)
