package usolana

const (
	LamportsPerSOL          uint64 = 1000000000                               // 1 SOL = 1,000,000,000 Lamports
	MicroLamportsPerLamport uint64 = 1000000                                  // 1 Lamport = 1,000,000 MicroLamports
	MicroLamportsPerSOL     uint64 = MicroLamportsPerLamport * LamportsPerSOL // 1 SOL = 1,000,000,000,000 MicroLamports
	SOLDecimals             uint8  = 9
)

const (
	ProgramIDSystem        = "11111111111111111111111111111111"
	ProgramIDComputeBudget = "ComputeBudget111111111111111111111111111111"
	ProgramIDToken         = "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"
)
