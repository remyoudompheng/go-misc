// Implemented using SSE2 arithmetic.

// func DivmodU64(a, b uint64) (quo, rem uint64)
TEXT ·DivmodU64(SB),7,$0
        JMP ·divmodU64(SB)

// func Mul(a, b uint64) (hi, lo uint64)
TEXT ·MulU64(SB),7,$0
        // Compute alo * blo into lo
        MOVL a+0(FP), AX
        MULL b+8(FP)
        MOVL DX, lo+28(FP)
        MOVL AX, lo+24(FP)

        // Compute ahi * bhi into hi
        MOVL a+4(FP), AX
        MULL b+12(FP)
        MOVL DX, hi+20(FP)
        MOVL AX, hi+16(FP)

        // Compute alo*bhi+ahi*blo
        MOVL alo+0(FP), AX
        MULL bhi+12(FP)
        ADDL DX, hi+16(FP)
        JNC  nocarry1
        INCL  hi+20(FP)
nocarry1:
        ADDL AX, lo+28(FP)
        JNC  nocarry2
        INCL  hi+16(FP)
nocarry2:

        MOVL ahi+4(FP), AX
        MULL blo+8(FP)
        ADDL DX, hi+16(FP)
        JNC  nocarry3
        INCL  hi+20(FP)
nocarry3:
        ADDL AX, lo+28(FP)
        JNC  nocarry4
        INCL  hi+16(FP)
nocarry4:
        RET

// func MulI64(a, b int64) (hi, lo int64)
TEXT ·MulI64(SB),7,$0
        JMP ·mulI64(SB)

