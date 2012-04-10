// func Divmod(a, b uint64) (quo, rem uint64)
TEXT ·Divmod(SB),7,$0
        MOVQ $0, DX
        MOVQ a+0(FP), AX
        DIVQ b+8(FP),
        MOVQ AX, quo+16(FP)
        MOVQ DX, rem+24(FP)
        RET

// func Mul(a, b uint64) (hi, lo uint64)
TEXT ·Mul(SB),7,$0
        MOVQ a+0(FP), AX
        MULQ b+8(FP),
        MOVQ DX, hi+16(FP)
        MOVQ AX, lo+24(FP)
        RET

// func MulS(a, b int64) (hi, lo int64)
TEXT ·MulS(SB),7,$0
        RET

