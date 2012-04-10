package extramath

// Divmod computes the quotient and remainder of division of a by b.
func Divmod(a, b uint64) (quo, rem uint64)

func MulS(a, b int64) (hi, lo int64)

// Mul computes tha 128-bit product of a by b as hi<<64|lo.
func Mul(a, b uint64) (hi, lo uint64)
