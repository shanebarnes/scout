package main

const (
    BIN_   float64 = 1 << iota
    BIN_KI float64 = 1 << (10 * iota)
    BIN_MI
    BIN_GI
    BIN_TI
    BIN_PI
    BIN_EI
)

const (
    DEC_  float64 = 1 << iota
    DEC_K float64 = DEC_  * 1000
    DEC_M         = DEC_K * 1000
    DEC_G         = DEC_M * 1000
    DEC_T         = DEC_G * 1000
    DEC_P         = DEC_T * 1000
    DEC_E         = DEC_P * 1000
)

func ToUnits(val float64, base uint) (float64, string) {
    prefix := DEC_
    mult := DEC_K
    if base == 2 {
        mult = BIN_KI
    }

    for val >= (prefix * mult) {
        prefix = prefix * mult
    }

    return val / prefix, GetUnitPrefix(prefix)
}

func GetUnitPrefix(prefix float64) string {
    var ret string

    switch prefix {
        default:
        case BIN_:
            ret = ""
        case BIN_KI:
            ret = "Ki"
        case BIN_MI:
            ret = "Mi"
        case BIN_GI:
            ret = "Gi"
        case BIN_TI:
            ret = "Ti"
        case BIN_PI:
            ret = "Pi"
        case BIN_EI:
            ret = "Ei"
        case DEC_K:
            ret = "k"
        case DEC_M:
            ret = "M"
        case DEC_G:
            ret = "G"
        case DEC_T:
            ret = "T"
        case DEC_P:
            return "P"
        case DEC_E:
            ret = "E"
    }

    return ret
}
